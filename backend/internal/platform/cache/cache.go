// Package cache is a small read-through cache abstraction. The Cache port has a
// Redis-backed implementation (production), an in-memory one (tests / single
// instance) and a no-op one (caching disabled). Values are opaque bytes; callers
// choose the encoding (the article module caches JSON).
package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yourorg/goapp/internal/platform/config"
)

// Cache is the read-through cache port. Get's second result reports a hit; a miss
// is (nil, false, nil). Implementations are safe for concurrent use.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// New selects a Cache from configuration: Redis when enabled, otherwise a no-op.
func New(cfg config.CacheConfig, redisCfg config.RedisConfig) Cache {
	if !cfg.Enabled {
		return Nop{}
	}
	return NewRedis(redisCfg)
}

// ── Redis ─────────────────────────────────────────────────────────────────────

type RedisCache struct {
	client *redis.Client
}

func NewRedis(cfg config.RedisConfig) *RedisCache {
	return &RedisCache{client: redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	b, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Close releases the Redis connections.
func (c *RedisCache) Close() error { return c.client.Close() }

// ── in-memory ───────────────────────────────────────────────────────────────

type entry struct {
	value     []byte
	expiresAt time.Time // zero = no expiry
}

// Memory is an in-process cache, handy for tests and single-instance setups.
type Memory struct {
	mu    sync.Mutex
	items map[string]entry
}

func NewMemory() *Memory {
	return &Memory{items: make(map[string]entry)}
}

func (m *Memory) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.items[key]
	if !ok {
		return nil, false, nil
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		delete(m.items, key)
		return nil, false, nil
	}
	return e.value, true, nil
}

func (m *Memory) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = entry{value: value, expiresAt: expiresAt}
	return nil
}

func (m *Memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
	return nil
}

// ── no-op ─────────────────────────────────────────────────────────────────────

// Nop is a cache that stores nothing: every Get misses. Used when caching is
// disabled so callers need no nil checks.
type Nop struct{}

func (Nop) Get(context.Context, string) ([]byte, bool, error)        { return nil, false, nil }
func (Nop) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (Nop) Delete(context.Context, string) error                     { return nil }
