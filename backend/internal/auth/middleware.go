package auth

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

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
func Authenticate(jwtm *JWTManager, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			httpx.WriteError(c, logger, apperror.Unauthorized("invalid_token", "Malformed Authorization header"))
			c.Abort()
			return
		}
		claims, err := jwtm.ParseAccess(token)
		if err != nil {
			httpx.WriteError(c, logger, err)
			c.Abort()
			return
		}
		actor, err := claims.Actor()
		if err != nil {
			httpx.WriteError(c, logger, err)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(authz.WithActor(c.Request.Context(), actor))
		c.Next()
	}
}
