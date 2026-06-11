package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/internal/user/userapi"
	"github.com/yourorg/goapp/pkg/apperror"
)

const uniqueViolationCode = "23505"

// repository is the write side of the user domain.
type repository struct {
	q *gen.Queries
}

func newRepository(db gen.DBTX) *repository {
	return &repository{q: gen.New(db)}
}

func (r *repository) Create(ctx context.Context, p userapi.CreateParams) (userapi.User, error) {
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
			return userapi.User{}, apperror.Conflict("email_taken", "Email is already registered")
		}
		return userapi.User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *repository) UpdateProfile(ctx context.Context, id int64, firstName, lastName string) (userapi.User, error) {
	row, err := r.q.UpdateUserProfile(ctx, gen.UpdateUserProfileParams{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return userapi.User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}
