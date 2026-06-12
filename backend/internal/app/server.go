// Package app wires the platform and domains together into an HTTP application
// and runs it. NewServer builds the fully configured http.Handler (the Mat
// Ryer / Grafana pattern: a single constructor that takes its dependencies and
// returns a handler); Run owns process lifecycle and graceful shutdown.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/api"
	"github.com/yourorg/goapp/internal/article"
	"github.com/yourorg/goapp/internal/auth"
	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/cache"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/health"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/internal/platform/metrics"
	"github.com/yourorg/goapp/internal/platform/middleware"
	"github.com/yourorg/goapp/internal/user"
)

// NewServer constructs the application's HTTP handler with all routes and
// middleware wired in. All dependencies are passed as arguments; nothing is read
// from globals.
func NewServer(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool, notifier auth.Notifier, articleCache cache.Cache) http.Handler {
	// ── compose dependencies ────────────────────────────────────────────────
	// The composition root is the one place allowed to construct concrete
	// modules and wire them together. auth consumes the user module and the
	// notifications module through its own ports (auth.Users, auth.Notifier);
	// the concrete implementations are injected here.
	jwtManager := auth.NewJWTManager(cfg.JWT)
	userModule := user.New(pool)
	authService := auth.NewService(pool, jwtManager, userModule)

	authHandler := auth.NewHandler(authService, userModule, notifier, cfg.Cookie, jwtManager.RefreshTTL(), logger)
	userHandler := user.NewHandler(userModule, logger)
	articleHandler := article.NewHandler(pool, logger, articleCache, cfg.Cache.TTL)

	authenticate := auth.Authenticate(jwtManager, logger)
	requireAuth := authz.RequireAuth(logger)

	// ── router ──────────────────────────────────────────────────────────────
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// Derive ClientIP from the real connection only; behind a trusted proxy,
	// replace nil with the proxy CIDRs so rate limiting can't be spoofed via
	// X-Forwarded-For.
	_ = r.SetTrustedProxies(nil)

	// Global middleware — applies to everything below, including health probes.
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.SecurityHeaders(cfg.Cookie.Secure))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	if cfg.Metrics.Enabled {
		registry := metrics.New()
		r.Use(metrics.Middleware(registry))
		r.GET("/metrics", metrics.Handler(registry))
	}

	// Health/readiness probes are registered before the rate limiter and request
	// timeout, so frequent orchestrator probes are never throttled.
	checker := health.New()
	checker.Register("database", func(ctx context.Context) error { return pool.Ping(ctx) })
	r.GET("/health", live)                 // liveness alias (kept for convenience)
	r.GET("/health/live", live)            // is the process up?
	r.GET("/health/ready", ready(checker)) // can it serve traffic? (pings the DB)

	// API contract: the raw spec plus an interactive docs UI.
	r.GET("/openapi.yaml", openapiSpec)
	r.GET("/docs", docsUI)

	// Heavier protections apply to the API surface, not to health probes.
	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimit(cfg.RateLimit.RPS, cfg.RateLimit.Burst, logger))
	}
	r.Use(middleware.BodySizeLimit(cfg.HTTP.MaxBodyBytes))
	r.Use(middleware.Timeout(cfg.HTTP.RequestTimeout))

	api := r.Group("/api")
	// Optional authentication for the whole API surface: populates the actor
	// when a valid token is present; protected routes additionally use
	// requireAuth.
	api.Use(authenticate)

	authHandler.RegisterRoutes(api.Group("/auth"), requireAuth)

	v1 := api.Group("/v1")
	userHandler.RegisterRoutes(v1.Group("/users"), requireAuth)
	articleHandler.RegisterRoutes(v1.Group("/articles"), requireAuth)

	return r
}

type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// live reports that the process is up and able to answer requests.
func live(c *gin.Context) {
	httpx.JSON(c, http.StatusOK, healthResponse{Status: "ok"})
}

// openapiSpec serves the embedded OpenAPI document.
func openapiSpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", api.Spec)
}

// docsUI serves an interactive API reference (Scalar) that loads the spec.
func docsUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(docsHTML))
}

const docsHTML = `<!doctype html>
<html>
  <head>
    <title>goapp API</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/openapi.yaml"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

// ready runs the dependency checks (e.g. a DB ping) and reports 200 when the
// service can serve traffic, 503 otherwise.
func ready(checker *health.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		ok, statuses := checker.Run(ctx)
		status := http.StatusOK
		word := "ok"
		if !ok {
			status = http.StatusServiceUnavailable
			word = "unavailable"
		}
		httpx.JSON(c, status, healthResponse{Status: word, Checks: statuses})
	}
}
