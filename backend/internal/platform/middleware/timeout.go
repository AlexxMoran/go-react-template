package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout attaches a deadline to each request's context. Downstream work that
// honors context cancellation (the pgx pool does) is aborted once the deadline
// passes. It is cooperative: an already-running handler is not preempted, but its
// next context-aware call returns context.DeadlineExceeded. A non-positive d
// disables the timeout.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d <= 0 {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
