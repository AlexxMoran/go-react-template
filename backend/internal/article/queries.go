package article

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Queries is the read side of the article domain.
type Queries struct {
	q *gen.Queries
}

func NewQueries(db gen.DBTX) *Queries {
	return &Queries{q: gen.New(db)}
}

func (r *Queries) GetByID(ctx context.Context, id int64) (Article, error) {
	row, err := r.q.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Article{}, apperror.NotFound("not_found", "Article not found")
		}
		return Article{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}
