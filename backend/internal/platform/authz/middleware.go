package authz

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/pkg/apperror"
)

// RequireAuth rejects requests that have no authenticated actor in context.
// It runs after the authentication middleware, which populates the actor from a
// valid access token. This is the equivalent of FastAPI's `current_user`
// dependency, as opposed to the optional `current_user_or_none`.
func RequireAuth(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := ActorFrom(c.Request.Context()); !ok {
			httpx.WriteError(c, logger, apperror.Unauthorized("unauthorized", "Authentication required"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRoleMW rejects requests whose actor does not hold the given role.
func RequireRoleMW(logger *slog.Logger, role Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, _ := ActorFrom(c.Request.Context())
		if err := RequireRole(actor, role); err != nil {
			httpx.WriteError(c, logger, err)
			c.Abort()
			return
		}
		c.Next()
	}
}
