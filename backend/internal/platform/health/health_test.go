package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourorg/goapp/internal/platform/health"
)

func TestChecker_AllPass(t *testing.T) {
	c := health.New()
	c.Register("database", func(context.Context) error { return nil })

	ready, statuses := c.Run(context.Background())
	if !ready {
		t.Errorf("ready = false, want true")
	}
	if statuses["database"] != "ok" {
		t.Errorf("database status = %q, want ok", statuses["database"])
	}
}

func TestChecker_OneFails(t *testing.T) {
	c := health.New()
	c.Register("database", func(context.Context) error { return nil })
	c.Register("cache", func(context.Context) error { return errors.New("connection refused") })

	ready, statuses := c.Run(context.Background())
	if ready {
		t.Errorf("ready = true, want false")
	}
	if statuses["cache"] != "connection refused" {
		t.Errorf("cache status = %q, want %q", statuses["cache"], "connection refused")
	}
}

func TestChecker_NoChecksIsReady(t *testing.T) {
	ready, _ := health.New().Run(context.Background())
	if !ready {
		t.Errorf("ready = false with no checks, want true")
	}
}
