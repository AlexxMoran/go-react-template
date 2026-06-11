package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets conservative security response headers on every response.
// HSTS is emitted only when secure is true (the service is served over HTTPS),
// since it is meaningless — and a foot-gun — over plain HTTP in development.
func SecurityHeaders(secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if secure {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		c.Next()
	}
}
