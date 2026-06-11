// Package adapters implements the article application's Store port over
// PostgreSQL (pgx + sqlc-generated queries). It is the only place in the module
// that touches the database driver and maps generated rows to the domain model.
package adapters

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/article/app"
	"github.com/yourorg/goapp/internal/article/domain"
	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Store is the PostgreSQL-backed implementation of app.Store.
type Store struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

// NewStore builds the adapter over a connection pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: gen.New(pool)}
}

// ── reads ───────────────────────────────────────────────────────────────────

func (s *Store) Get(ctx context.Context, id int64) (domain.Article, error) {
	row, err := s.q.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Article{}, apperror.NotFound("not_found", "Article not found")
		}
		return domain.Article{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (s *Store) List(ctx context.Context, f app.ListFilter, skip, limit int) ([]domain.Article, error) {
	rows, err := s.q.ListArticles(ctx, gen.ListArticlesParams{
		Status:   database.TextPtr(statusStr(f.Status)),
		AuthorID: database.Int8Ptr(f.AuthorID),
		Search:   database.TextPtr(f.Search),
		Off:      int32(skip),
		Lim:      int32(limit),
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	articles := make([]domain.Article, 0, len(rows))
	for _, row := range rows {
		articles = append(articles, fromGen(row))
	}
	return articles, nil
}

func (s *Store) Count(ctx context.Context, f app.ListFilter) (int64, error) {
	total, err := s.q.CountArticles(ctx, gen.CountArticlesParams{
		Status:   database.TextPtr(statusStr(f.Status)),
		AuthorID: database.Int8Ptr(f.AuthorID),
		Search:   database.TextPtr(f.Search),
	})
	if err != nil {
		return 0, apperror.Internal(err)
	}
	return total, nil
}

// ── single-statement writes ───────────────────────────────────────────────────

func (s *Store) Create(ctx context.Context, authorID int64, title, content string) (domain.Article, error) {
	row, err := s.q.CreateArticle(ctx, gen.CreateArticleParams{
		AuthorID: authorID,
		Title:    title,
		Content:  content,
	})
	if err != nil {
		return domain.Article{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (s *Store) Update(ctx context.Context, id int64, title, content string) (domain.Article, error) {
	row, err := s.q.UpdateArticle(ctx, gen.UpdateArticleParams{
		ID:      id,
		Title:   title,
		Content: content,
	})
	if err != nil {
		return domain.Article{}, apperror.Internal(err)
	}
	return fromGen(row), nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	if err := s.q.DeleteArticle(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

// ── transactional operation ───────────────────────────────────────────────────

// RunPublish loads the article, asks the caller's pure decider what to do, and
// applies the plan — all inside one transaction. The decider receives plain
// facts (a domain.PublishSnapshot) and never sees the driver.
func (s *Store) RunPublish(
	ctx context.Context,
	id int64,
	decide func(domain.PublishSnapshot) (domain.PublishDecision, error),
) (domain.Article, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Article{}, apperror.Internal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := gen.New(tx)

	row, err := q.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Article{}, apperror.NotFound("not_found", "Article not found")
		}
		return domain.Article{}, apperror.Internal(err)
	}

	decision, err := decide(domain.PublishSnapshot{
		ArticleID:  row.ID,
		Status:     domain.Status(row.Status),
		Title:      row.Title,
		HasContent: row.Content != "",
	})
	if err != nil {
		return domain.Article{}, err
	}

	updated, err := q.SetArticleStatus(ctx, gen.SetArticleStatusParams{
		ID:          decision.ArticleID,
		Status:      string(decision.NewStatus),
		PublishedAt: database.Timestamptz(decision.PublishedAt),
	})
	if err != nil {
		return domain.Article{}, apperror.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Article{}, apperror.Internal(err)
	}
	return fromGen(updated), nil
}

// ── mapping ───────────────────────────────────────────────────────────────────

func fromGen(row gen.Article) domain.Article {
	return domain.Article{
		ID:          row.ID,
		AuthorID:    row.AuthorID,
		Title:       row.Title,
		Content:     row.Content,
		Status:      domain.Status(row.Status),
		PublishedAt: database.TimePtr(row.PublishedAt),
		CreatedAt:   database.TimeOrZero(row.CreatedAt),
		UpdatedAt:   database.TimeOrZero(row.UpdatedAt),
	}
}

func statusStr(s *domain.Status) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}
