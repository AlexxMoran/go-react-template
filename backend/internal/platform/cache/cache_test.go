package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/goapp/internal/platform/cache"
)

func TestMemory_SetGetDelete(t *testing.T) {
	ctx := context.Background()
	c := cache.NewMemory()

	if _, ok, _ := c.Get(ctx, "k"); ok {
		t.Fatal("expected miss on empty cache")
	}

	if err := c.Set(ctx, "k", []byte("v"), time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, ok, err := c.Get(ctx, "k")
	if err != nil || !ok || string(got) != "v" {
		t.Fatalf("get = (%q, %v, %v), want (v, true, nil)", got, ok, err)
	}

	if err := c.Delete(ctx, "k"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok, _ := c.Get(ctx, "k"); ok {
		t.Error("expected miss after delete")
	}
}

func TestMemory_Expiry(t *testing.T) {
	ctx := context.Background()
	c := cache.NewMemory()

	if err := c.Set(ctx, "k", []byte("v"), 10*time.Millisecond); err != nil {
		t.Fatalf("set: %v", err)
	}
	time.Sleep(25 * time.Millisecond)
	if _, ok, _ := c.Get(ctx, "k"); ok {
		t.Error("expected miss after TTL expiry")
	}
}

func TestNop_AlwaysMisses(t *testing.T) {
	ctx := context.Background()
	var c cache.Cache = cache.Nop{}

	if err := c.Set(ctx, "k", []byte("v"), time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, ok, _ := c.Get(ctx, "k"); ok {
		t.Error("nop cache must always miss")
	}
}
