// Package app wires the platform and domains together into an HTTP application
// and runs it. NewServer builds the fully configured http.Handler (the Mat
// Ryer / Grafana pattern: a single constructor that takes its dependencies and
// returns a handler); Run owns process lifecycle and graceful shutdown.
package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

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
	jwtManager := auth.NewJWTManager(cfg.JWT)
	authService := auth.NewService(pool, jwtManager)
	userQueries := user.NewQueries(pool)

	authHandler := auth.NewHandler(authService, userQueries, cfg.Cookie, jwtManager.RefreshTTL(), logger)
	userHandler := user.NewHandler(pool, logger)
	articleHandler := article.NewHandler(pool, logger)

	authenticate := auth.Authenticate(jwtManager, logger)
	requireAuth := authz.RequireAuth(logger)

	// ── router ──────────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	r.Get("/health", health)

	r.Route("/api", func(r chi.Router) {
		// Optional authentication for the whole API surface: populates the actor
		// when a valid token is present; protected routes additionally use
		// requireAuth.
		r.Use(authenticate)

		r.Mount("/auth", authHandler.Routes(requireAuth))

		r.Route("/v1", func(r chi.Router) {
			r.Mount("/users", userHandler.Routes(requireAuth))
			r.Mount("/articles", articleHandler.Routes(requireAuth))
		})
	})

	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
