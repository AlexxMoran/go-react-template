//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/goapp/internal/article/adapters"
	"github.com/yourorg/goapp/internal/article/app"
	"github.com/yourorg/goapp/internal/platform/cache"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/test/testsupport"
)

// TestArticleCache_ReadThroughAndInvalidation proves the cache-aside behaviour
// against a real Redis: reads are served from cache (observably stale after a
// direct DB change), and a write through the module invalidates the entry.
func TestArticleCache_ReadThroughAndInvalidation(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	author := createAuthor(t)

	redisCache := cache.NewRedis(config.RedisConfig{Addr: redisAddr})
	defer func() { _ = redisCache.Close() }()
	mod := app.NewModule(adapters.NewCachedStore(adapters.NewStore(pool), redisCache, time.Minute))

	created, err := mod.Create(ctx, author.ID, "Original", "body")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Warm the cache for this id.
	if _, err := mod.Get(ctx, created.ID); err != nil {
		t.Fatalf("warm get: %v", err)
	}

	// Change the row directly, bypassing the module and its cache.
	if _, err := pool.Exec(ctx, `UPDATE articles SET title = 'Changed in DB' WHERE id = $1`, created.ID); err != nil {
		t.Fatalf("direct update: %v", err)
	}

	// The read is served from cache, so it still sees the original title.
	cached, err := mod.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("cached get: %v", err)
	}
	if cached.Title != "Original" {
		t.Errorf("title = %q, want Original (cache hit should be stale)", cached.Title)
	}

	// A write through the module invalidates the cache entry.
	if _, err := mod.Update(ctx, created.ID, "Updated via module", "body2"); err != nil {
		t.Fatalf("update: %v", err)
	}
	fresh, err := mod.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("fresh get: %v", err)
	}
	if fresh.Title != "Updated via module" {
		t.Errorf("title = %q, want 'Updated via module' (cache invalidated)", fresh.Title)
	}
}
