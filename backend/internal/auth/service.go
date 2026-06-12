package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/internal/user/userapi"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Users is the consumer-defined port over the user module: auth depends only on
// the slice of user behavior it actually needs, expressed in terms of the user
// module's published contract (userapi). The composition root injects the
// concrete *user.Module, which satisfies this interface structurally.
type Users interface {
	Create(ctx context.Context, p userapi.CreateParams) (userapi.User, error)
	UpdateProfile(ctx context.Context, id int64, firstName, lastName string) (userapi.User, error)
	GetByID(ctx context.Context, id int64) (userapi.User, error)
	GetByEmail(ctx context.Context, email string) (userapi.User, error)
}

// TokenPair is the result of a successful login or refresh.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// Service implements the authentication use-cases. Refresh tokens are persisted
// as SHA-256 hashes so they can be rotated and revoked (the raw token never
// touches the database).
type Service struct {
	pool  *pgxpool.Pool
	users Users
	q     *gen.Queries
	jwt   *JWTManager
}

func NewService(pool *pgxpool.Pool, jwt *JWTManager, users Users) *Service {
	return &Service{
		pool:  pool,
		users: users,
		q:     gen.New(pool),
		jwt:   jwt,
	}
}

// Register creates a new account with the default user role.
func (s *Service) Register(ctx context.Context, email, password, firstName, lastName string) (userapi.User, error) {
	hashed, err := HashPassword(password)
	if err != nil {
		return userapi.User{}, err
	}
	return s.users.Create(ctx, userapi.CreateParams{
		Email:          email,
		HashedPassword: hashed,
		FirstName:      firstName,
		LastName:       lastName,
	})
}

// Login verifies credentials and issues a fresh token pair.
func (s *Service) Login(ctx context.Context, email, password string) (userapi.User, TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return userapi.User{}, TokenPair{}, apperror.Unauthorized("invalid_credentials", "Incorrect email or password")
	}
	if !CheckPassword(u.HashedPassword, password) {
		return userapi.User{}, TokenPair{}, apperror.Unauthorized("invalid_credentials", "Incorrect email or password")
	}
	if !u.IsActive {
		return userapi.User{}, TokenPair{}, apperror.Forbidden("inactive_user", "Account is inactive")
	}
	pair, err := s.issueTokens(ctx, u)
	if err != nil {
		return userapi.User{}, TokenPair{}, err
	}
	return u, pair, nil
}

// Refresh validates a refresh token, rotates it (revoking the old one), and
// issues a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	userID, err := s.jwt.ParseRefresh(refreshToken)
	if err != nil {
		return TokenPair{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return TokenPair{}, apperror.Internal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := s.q.WithTx(tx)
	tokenHash := hashToken(refreshToken)

	stored, err := qtx.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, apperror.Unauthorized("invalid_token", "Refresh token is not recognized")
		}
		return TokenPair{}, apperror.Internal(err)
	}
	if stored.UserID != userID {
		return TokenPair{}, apperror.Unauthorized("invalid_token", "Refresh token user mismatch")
	}
	if stored.RevokedAt.Valid {
		return TokenPair{}, apperror.Unauthorized("token_revoked", "Refresh token has been revoked")
	}
	if stored.ExpiresAt.Valid && time.Now().After(stored.ExpiresAt.Time) {
		return TokenPair{}, apperror.Unauthorized("invalid_token", "Refresh token has expired")
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return TokenPair{}, err
	}

	access, err := s.jwt.GenerateAccess(u)
	if err != nil {
		return TokenPair{}, err
	}
	newRefresh, expiresAt, err := s.jwt.GenerateRefresh(u.ID)
	if err != nil {
		return TokenPair{}, err
	}

	if err := qtx.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return TokenPair{}, apperror.Internal(err)
	}
	if _, err := qtx.CreateRefreshToken(ctx, gen.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: hashToken(newRefresh),
		ExpiresAt: database.Timestamptz(expiresAt),
	}); err != nil {
		return TokenPair{}, apperror.Internal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return TokenPair{}, apperror.Internal(err)
	}

	return TokenPair{AccessToken: access, RefreshToken: newRefresh, RefreshExpiresAt: expiresAt}, nil
}

// UpdateProfile updates the current user's editable profile fields.
func (s *Service) UpdateProfile(ctx context.Context, id int64, firstName, lastName string) (userapi.User, error) {
	return s.users.UpdateProfile(ctx, id, firstName, lastName)
}

// Logout revokes a refresh token (idempotent).
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	if err := s.q.RevokeRefreshToken(ctx, hashToken(refreshToken)); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) issueTokens(ctx context.Context, u userapi.User) (TokenPair, error) {
	access, err := s.jwt.GenerateAccess(u)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, expiresAt, err := s.jwt.GenerateRefresh(u.ID)
	if err != nil {
		return TokenPair{}, err
	}
	if _, err := s.q.CreateRefreshToken(ctx, gen.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: hashToken(refresh),
		ExpiresAt: database.Timestamptz(expiresAt),
	}); err != nil {
		return TokenPair{}, apperror.Internal(err)
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, RefreshExpiresAt: expiresAt}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
