package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourorg/goapp/internal/article/app"
	"github.com/yourorg/goapp/internal/article/domain"
	"github.com/yourorg/goapp/internal/platform/cache"
)

// CachedStore decorates an app.Store with read-through caching of Get and
// write-through invalidation on mutations. Cache problems are never fatal: on any
// miss or error it falls through to the wrapped store, so the cache can only make
// reads faster, never wrong. It implements app.Store, so the application layer is
// unaware caching exists.
type CachedStore struct {
	inner app.Store
	cache cache.Cache
	ttl   time.Duration
}

// NewCachedStore wraps inner with the given cache and per-entry TTL.
func NewCachedStore(inner app.Store, c cache.Cache, ttl time.Duration) *CachedStore {
	return &CachedStore{inner: inner, cache: c, ttl: ttl}
}

func articleKey(id int64) string { return fmt.Sprintf("article:%d", id) }

func (s *CachedStore) Get(ctx context.Context, id int64) (domain.Article, error) {
	key := articleKey(id)
	if b, ok, _ := s.cache.Get(ctx, key); ok {
		var a domain.Article
		if json.Unmarshal(b, &a) == nil {
			return a, nil
		}
	}
	a, err := s.inner.Get(ctx, id)
	if err != nil {
		return a, err
	}
	if b, e := json.Marshal(a); e == nil {
		_ = s.cache.Set(ctx, key, b, s.ttl)
	}
	return a, nil
}

func (s *CachedStore) Update(ctx context.Context, id int64, title, content string) (domain.Article, error) {
	a, err := s.inner.Update(ctx, id, title, content)
	if err == nil {
		_ = s.cache.Delete(ctx, articleKey(id))
	}
	return a, err
}

func (s *CachedStore) Delete(ctx context.Context, id int64) error {
	err := s.inner.Delete(ctx, id)
	if err == nil {
		_ = s.cache.Delete(ctx, articleKey(id))
	}
	return err
}

func (s *CachedStore) RunPublish(
	ctx context.Context,
	id int64,
	decide func(domain.PublishSnapshot) (domain.PublishDecision, error),
) (domain.Article, error) {
	a, err := s.inner.RunPublish(ctx, id, decide)
	if err == nil {
		_ = s.cache.Delete(ctx, articleKey(id))
	}
	return a, err
}

// List, Count and Create have no single-id cache entry, so they pass through.
func (s *CachedStore) List(ctx context.Context, f app.ListFilter, skip, limit int) ([]domain.Article, error) {
	return s.inner.List(ctx, f, skip, limit)
}

func (s *CachedStore) Count(ctx context.Context, f app.ListFilter) (int64, error) {
	return s.inner.Count(ctx, f)
}

func (s *CachedStore) Create(ctx context.Context, authorID int64, title, content string) (domain.Article, error) {
	return s.inner.Create(ctx, authorID, title, content)
}
