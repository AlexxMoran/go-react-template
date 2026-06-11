package article

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts the article routes. List and Get allow anonymous access
// (the policy still hides drafts); mutations require authentication.
func (h *Handler) RegisterRoutes(r gin.IRouter, requireAuth gin.HandlerFunc) {
	r.GET("/", h.List)
	r.GET("/:id", h.Get)

	protected := r.Group("/")
	protected.Use(requireAuth)
	protected.POST("/", h.Create)
	protected.PATCH("/:id", h.Update)
	protected.DELETE("/:id", h.Delete)
	protected.POST("/:id/publish", h.Publish)
}
