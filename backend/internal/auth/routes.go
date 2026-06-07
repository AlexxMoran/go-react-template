package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes returns the auth sub-router. Public endpoints are open; /me endpoints
// are wrapped with the provided requireAuth middleware.
func (h *Handler) Routes(requireAuth func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)

	r.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/me", h.Me)
		r.Patch("/me", h.UpdateMe)
	})

	return r
}
