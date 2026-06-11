//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/goapp/internal/app"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/logger"
)

// newTestServer builds the full HTTP handler (all middleware + routes) over the
// suite's live database.
func newTestServer() http.Handler {
	cfg := config.Config{
		Env: "test",
		HTTP: config.HTTPConfig{
			RequestTimeout: 5 * time.Second,
			MaxBodyBytes:   1 << 20,
		},
		JWT: config.JWTConfig{
			AccessSecret:  "test-access-secret-test-access-secret",
			RefreshSecret: "test-refresh-secret-test-refresh-secret",
			Issuer:        "goapp-test",
			AccessTTL:     15 * time.Minute,
			RefreshTTL:    24 * time.Hour,
		},
		RateLimit: config.RateLimitConfig{Enabled: false},
	}
	return app.NewServer(cfg, logger.New("test", "error"), pool)
}

func TestServer_HealthLive(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/live", nil)
	newTestServer().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// TestServer_HealthReady checks that readiness actually probes the database: with
// a healthy pool it reports 200 and database=ok.
func TestServer_HealthReady(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/ready", nil)
	newTestServer().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want 200", w.Code)
	}
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" || body.Checks["database"] != "ok" {
		t.Errorf("ready body = %+v, want status=ok database=ok", body)
	}
}

func TestServer_SecurityHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/live", nil)
	newTestServer().ServeHTTP(w, req)

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", got)
	}
}
