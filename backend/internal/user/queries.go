package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Queries is the read side of the user domain.
type Queries struct {
	q *gen.Queries
}

func NewQueries(db gen.DBTX) *Queries {
	return &Queries{q: gen.New(db)}
}

func (r *Queries) GetByID(ctx context.Context, id int64) (User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, apperror.NotFound("not_found", "User not found")
		}
		return User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *Queries) GetByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, apperror.NotFound("not_found", "User not found")
		}
		return User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *Queries) List(ctx context.Context, skip, limit int) ([]User, error) {
	rows, err := r.q.ListUsers(ctx, gen.ListUsersParams{Off: int32(skip), Lim: int32(limit)})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, fromGen(row))
	}
	return users, nil
}

func (r *Queries) Count(ctx context.Context) (int64, error) {
	total, err := r.q.CountUsers(ctx)
	if err != nil {
		return 0, apperror.Internal(err)
	}
	return total, nil
}
