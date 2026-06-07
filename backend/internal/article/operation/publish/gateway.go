package publish

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Gateway performs all database I/O for the operation, bound to a single
// transaction. It returns the raw sqlc row from Apply; the caller maps it to the
// domain model. Keeping the gateway free of the article package avoids an import
// cycle (article -> publish -> article).
type Gateway struct {
	q *gen.Queries
}

func NewGateway(tx pgx.Tx) *Gateway {
	return &Gateway{q: gen.New(tx)}
}

// Load reads the article and captures the facts the decision layer needs.
func (g *Gateway) Load(ctx context.Context, c Contract) (Snapshot, error) {
	row, err := g.q.GetArticleByID(ctx, c.ArticleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Snapshot{}, apperror.NotFound("not_found", "Article not found")
		}
		return Snapshot{}, apperror.Internal(err)
	}
	return Snapshot{
		ArticleID:  row.ID,
		Status:     row.Status,
		Title:      row.Title,
		HasContent: row.Content != "",
	}, nil
}

// Apply executes the decision's write plan and returns the updated row.
func (g *Gateway) Apply(ctx context.Context, d Decision) (gen.Article, error) {
	row, err := g.q.SetArticleStatus(ctx, gen.SetArticleStatusParams{
		ID:          d.ArticleID,
		Status:      d.NewStatus,
		PublishedAt: database.Timestamptz(d.PublishedAt),
	})
	if err != nil {
		return gen.Article{}, apperror.Internal(err)
	}
	return row, nil
}
