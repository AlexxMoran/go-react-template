package auth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Authenticate parses an optional bearer access token and, when valid, stores
// the resulting Actor in the request context. A missing token is allowed (the
// request proceeds anonymously); a malformed or invalid token is rejected.
//
// This is the counterpart to FastAPI's `current_user_or_none`. Routes that
// require an authenticated user additionally apply authz.RequireAuth.
func Authenticate(jwtm *JWTManager, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				httpx.WriteError(w, logger, apperror.Unauthorized("invalid_token", "Malformed Authorization header"))
				return
			}
			claims, err := jwtm.ParseAccess(token)
			if err != nil {
				httpx.WriteError(w, logger, err)
				return
			}
			actor, err := claims.Actor()
			if err != nil {
				httpx.WriteError(w, logger, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(authz.WithActor(r.Context(), actor)))
		})
	}
}
