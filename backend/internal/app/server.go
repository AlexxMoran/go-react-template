// Package app wires the platform and domains together into an HTTP application
// and runs it. NewServer builds the fully configured http.Handler (the Mat
// Ryer / Grafana pattern: a single constructor that takes its dependencies and
// returns a handler); Run owns process lifecycle and graceful shutdown.
package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/article"
	"github.com/yourorg/goapp/internal/auth"
	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/internal/platform/middleware"
	"github.com/yourorg/goapp/internal/user"
)

// NewServer constructs the application's HTTP handler with all routes and
// middleware wired in. All dependencies are passed as arguments; nothing is read
// from globals.
func NewServer(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) http.Handler {
	// ── compose dependencies ────────────────────────────────────────────────
	// The composition root is the one place allowed to construct concrete
	// modules and wire them together. auth consumes the user module through its
	// own port (auth.Users); *user.Module satisfies it structurally.
	jwtManager := auth.NewJWTManager(cfg.JWT)
	userModule := user.New(pool)
	authService := auth.NewService(pool, jwtManager, userModule)

	authHandler := auth.NewHandler(authService, userModule, cfg.Cookie, jwtManager.RefreshTTL(), logger)
	userHandler := user.NewHandler(userModule, logger)
	articleHandler := article.NewHandler(pool, logger)

	authenticate := auth.Authenticate(jwtManager, logger)
	requireAuth := authz.RequireAuth(logger)

	// ── router ──────────────────────────────────────────────────────────────
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	r.GET("/health", health)

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

func health(c *gin.Context) {
	httpx.JSON(c, http.StatusOK, map[string]string{"status": "ok"})
}
