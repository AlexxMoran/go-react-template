package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodySizeLimit caps the request body at maxBytes. Reads past the limit fail, so
// a handler that decodes an oversized body returns an error instead of buffering
// unbounded input into memory. A non-positive maxBytes disables the limit.
func BodySizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
