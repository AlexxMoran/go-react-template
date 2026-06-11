package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/goapp/internal/platform/middleware"
)

// newLimited builds a gin engine whose only route is rate-limited. Trusted
// proxies are cleared so ClientIP is the request's RemoteAddr, making the per-IP
// behavior deterministic.
func newLimited(rps float64, burst int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(middleware.RateLimit(rps, burst, nil))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func call(r *gin.Engine, ip string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	req.RemoteAddr = ip + ":1234"
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRateLimit_BlocksAfterBurst(t *testing.T) {
	r := newLimited(0.001, 2) // ~no refill during the test; burst of 2

	if got := call(r, "10.0.0.1"); got != http.StatusOK {
		t.Errorf("request 1: got %d, want 200", got)
	}
	if got := call(r, "10.0.0.1"); got != http.StatusOK {
		t.Errorf("request 2: got %d, want 200", got)
	}
	if got := call(r, "10.0.0.1"); got != http.StatusTooManyRequests {
		t.Errorf("request 3: got %d, want 429", got)
	}
}

func TestRateLimit_SeparateIPsIndependent(t *testing.T) {
	r := newLimited(0.001, 1)

	if got := call(r, "10.0.0.1"); got != http.StatusOK {
		t.Errorf("ip1 first: got %d, want 200", got)
	}
	if got := call(r, "10.0.0.2"); got != http.StatusOK {
		t.Errorf("ip2 first: got %d, want 200", got)
	}
	if got := call(r, "10.0.0.1"); got != http.StatusTooManyRequests {
		t.Errorf("ip1 second: got %d, want 429", got)
	}
}
