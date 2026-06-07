package article

import (
	"context"
	"time"

	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Repository is the write side of the article domain.
type Repository struct {
	q *gen.Queries
}

func NewRepository(db gen.DBTX) *Repository {
	return &Repository{q: gen.New(db)}
}

func (r *Repository) Create(ctx context.Context, authorID int64, title, content string) (Article, error) {
	row, err := r.q.CreateArticle(ctx, gen.CreateArticleParams{
		AuthorID: authorID,
		Title:    title,
		Content:  content,
	})
	if err != nil {
		return Article{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *Repository) Update(ctx context.Context, id int64, title, content string) (Article, error) {
	row, err := r.q.UpdateArticle(ctx, gen.UpdateArticleParams{
		ID:      id,
		Title:   title,
		Content: content,
	})
	if err != nil {
		return Article{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *Repository) SetStatus(ctx context.Context, id int64, status Status, publishedAt *time.Time) (Article, error) {
	row, err := r.q.SetArticleStatus(ctx, gen.SetArticleStatusParams{
		ID:          id,
		Status:      string(status),
		PublishedAt: database.TimestamptzPtr(publishedAt),
	})
	if err != nil {
		return Article{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if err := r.q.DeleteArticle(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}
