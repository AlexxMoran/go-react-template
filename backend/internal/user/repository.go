package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

const uniqueViolationCode = "23505"

// Repository is the write side of the user domain.
type Repository struct {
	q *gen.Queries
}

func NewRepository(db gen.DBTX) *Repository {
	return &Repository{q: gen.New(db)}
}

// CreateParams holds the fields required to create an account.
type CreateParams struct {
	Email          string
	HashedPassword string
	Role           authz.Role
	FirstName      string
	LastName       string
}

func (r *Repository) Create(ctx context.Context, p CreateParams) (User, error) {
	role := p.Role
	if role == "" {
		role = authz.RoleUser
	}
	row, err := r.q.CreateUser(ctx, gen.CreateUserParams{
		Email:          p.Email,
		HashedPassword: p.HashedPassword,
		Role:           string(role),
		FirstName:      p.FirstName,
		LastName:       p.LastName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return User{}, apperror.Conflict("email_taken", "Email is already registered")
		}
		return User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *Repository) UpdateProfile(ctx context.Context, id int64, firstName, lastName string) (User, error) {
	row, err := r.q.UpdateUserProfile(ctx, gen.UpdateUserProfileParams{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}
