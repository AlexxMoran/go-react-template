// Package middleware holds cross-cutting Gin middleware that is not tied to a
// specific domain: request IDs, panic recovery, structured request logging and
// CORS.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/pkg/apperror"
)

const requestIDKey = "request_id"

// RequestID attaches a stable request id to the Gin context and response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		c.Set(requestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// GetRequestID returns the request id assigned by RequestID.
func GetRequestID(c *gin.Context) string {
	if id, ok := c.Get(requestIDKey); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// Recoverer converts a panic in a downstream handler into a logged 500 response
// instead of crashing the server.
func Recoverer(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered",
					slog.Any("panic", rec),
					slog.String("path", c.Request.URL.Path),
					slog.String("request_id", GetRequestID(c)),
				)
				httpx.WriteError(c, logger, apperror.Internal(nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RequestLogger logs one structured line per request with method, path, status,
// duration and the request id.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int("bytes", c.Writer.Size()),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", GetRequestID(c)),
		)
	}
}

// CORS is a minimal CORS middleware driven by an allow-list of origins.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && slices.Contains(allowedOrigins, origin) {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b[:])
}
