package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/internal/user/userapi"
	"github.com/yourorg/goapp/pkg/apperror"
)

// queries is the read side of the user domain.
type queries struct {
	q *gen.Queries
}

func newQueries(db gen.DBTX) *queries {
	return &queries{q: gen.New(db)}
}

func (r *queries) GetByID(ctx context.Context, id int64) (userapi.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userapi.User{}, apperror.NotFound("not_found", "User not found")
		}
		return userapi.User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *queries) GetByEmail(ctx context.Context, email string) (userapi.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userapi.User{}, apperror.NotFound("not_found", "User not found")
		}
		return userapi.User{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *queries) List(ctx context.Context, skip, limit int) ([]userapi.User, error) {
	rows, err := r.q.ListUsers(ctx, gen.ListUsersParams{Off: int32(skip), Lim: int32(limit)})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	users := make([]userapi.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, fromGen(row))
	}
	return users, nil
}

func (r *queries) Count(ctx context.Context) (int64, error) {
	total, err := r.q.CountUsers(ctx)
	if err != nil {
		return 0, apperror.Internal(err)
	}
	return total, nil
}
