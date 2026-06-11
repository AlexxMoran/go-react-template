// Package auth owns authentication: password hashing, JWT issuing/parsing
// (short-lived access tokens + rotating refresh tokens), the authentication
// middleware that turns a bearer token into an authz.Actor, and the HTTP
// endpoints for register/login/refresh/logout/me.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/user/userapi"
	"github.com/yourorg/goapp/pkg/apperror"
)

const (
	audienceAccess  = "goapp:access"
	audienceRefresh = "goapp:refresh"
	signingAlg      = "HS256"
)

// AccessClaims is the payload of an access token: the standard registered claims
// plus the role and email needed to build an Actor without a database lookup.
type AccessClaims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Actor converts validated claims into the request-scoped authorization identity.
func (c *AccessClaims) Actor() (*authz.Actor, error) {
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return nil, apperror.Unauthorized("invalid_token", "Invalid token subject")
	}
	return &authz.Actor{ID: id, Email: c.Email, Role: authz.Role(c.Role)}, nil
}

// JWTManager issues and validates access and refresh tokens.
type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	issuer        string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		issuer:        cfg.Issuer,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

func (m *JWTManager) RefreshTTL() time.Duration { return m.refreshTTL }

// GenerateAccess signs a short-lived access token for the user.
func (m *JWTManager) GenerateAccess(u userapi.User) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		Role:  string(u.Role),
		Email: u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatInt(u.ID, 10),
			Audience:  jwt.ClaimStrings{audienceAccess},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.accessSecret)
	if err != nil {
		return "", apperror.Internal(err)
	}
	return signed, nil
}

// GenerateRefresh signs a long-lived refresh token and returns it with its
// expiry. A random JTI makes each token unique so its hash can be stored and
// revoked independently.
func (m *JWTManager) GenerateRefresh(userID int64) (token string, expiresAt time.Time, err error) {
	now := time.Now()
	expiresAt = now.Add(m.refreshTTL)
	jti, err := randomID()
	if err != nil {
		return "", time.Time{}, apperror.Internal(err)
	}
	claims := jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Subject:   strconv.FormatInt(userID, 10),
		Audience:  jwt.ClaimStrings{audienceRefresh},
		ID:        jti,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.refreshSecret)
	if err != nil {
		return "", time.Time{}, apperror.Internal(err)
	}
	return signed, expiresAt, nil
}

// ParseAccess validates an access token and returns its claims.
func (m *JWTManager) ParseAccess(tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, m.keyFunc(m.accessSecret),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(audienceAccess),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{signingAlg}),
	)
	if err != nil {
		return nil, apperror.Unauthorized("invalid_token", "Invalid or expired token")
	}
	return claims, nil
}

// ParseRefresh validates a refresh token and returns the subject user id.
func (m *JWTManager) ParseRefresh(tokenStr string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, m.keyFunc(m.refreshSecret),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(audienceRefresh),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{signingAlg}),
	)
	if err != nil {
		return 0, apperror.Unauthorized("invalid_token", "Invalid or expired refresh token")
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, apperror.Unauthorized("invalid_token", "Invalid token subject")
	}
	return id, nil
}

func (m *JWTManager) keyFunc(secret []byte) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
