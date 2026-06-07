package authz

import (
	"log/slog"
	"net/http"

	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/pkg/apperror"
)

// RequireAuth rejects requests that have no authenticated actor in context.
// It runs after the authentication middleware, which populates the actor from a
// valid access token. This is the equivalent of FastAPI's `current_user`
// dependency, as opposed to the optional `current_user_or_none`.
func RequireAuth(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := ActorFrom(r.Context()); !ok {
				httpx.WriteError(w, logger, apperror.Unauthorized("unauthorized", "Authentication required"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleMW rejects requests whose actor does not hold the given role.
func RequireRoleMW(logger *slog.Logger, role Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, _ := ActorFrom(r.Context())
			if err := RequireRole(actor, role); err != nil {
				httpx.WriteError(w, logger, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
