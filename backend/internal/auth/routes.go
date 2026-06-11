package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts the auth routes. Public endpoints are open; /me
// endpoints are wrapped with the provided requireAuth middleware.
func (h *Handler) RegisterRoutes(r gin.IRouter, requireAuth gin.HandlerFunc) {
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
	r.POST("/logout", h.Logout)

	protected := r.Group("/")
	protected.Use(requireAuth)
	protected.GET("/me", h.Me)
	protected.PATCH("/me", h.UpdateMe)
}
