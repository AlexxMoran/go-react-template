package app

import (
	"context"

	"github.com/yourorg/goapp/internal/article/domain"
)

// ── read side (queries) ────────────────────────────────────────────────────────

// Get returns a single article by id.
func (m *Module) Get(ctx context.Context, id int64) (domain.Article, error) {
	return m.store.Get(ctx, id)
}

// List returns a filtered, paginated page of articles.
func (m *Module) List(ctx context.Context, f ListFilter, skip, limit int) ([]domain.Article, error) {
	return m.store.List(ctx, f, skip, limit)
}

// Count returns the number of articles matching the filter.
func (m *Module) Count(ctx context.Context, f ListFilter) (int64, error) {
	return m.store.Count(ctx, f)
}
