package article

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes returns the article sub-router. List and Get allow anonymous access
// (the policy still hides drafts); mutations require authentication.
func (h *Handler) Routes(requireAuth func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.List)
	r.Get("/{id}", h.Get)

	r.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/publish", h.Publish)
	})

	return r
}
