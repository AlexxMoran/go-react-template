package user

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/httpx"
)

// Handler exposes the user HTTP endpoints (the authenticated user directory).
type Handler struct {
	queries *Queries
	logger  *slog.Logger
}

func NewHandler(pool *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{queries: NewQueries(pool), logger: logger}
}

// Routes returns the user sub-router. Listing the directory requires
// authentication (any signed-in user); tighten with authz.RequireRoleMW if you
// want it admin-only.
func (h *Handler) Routes(requireAuth func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/", h.List)
	})
	return r
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	skip, limit := httpx.Pagination(r)

	users, err := h.queries.List(r.Context(), skip, limit)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	total, err := h.queries.Count(r.Context())
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}

	data := make([]Response, 0, len(users))
	for _, u := range users {
		resp := ToResponse(u)
		resp.Permissions = NewPolicy(actor, u).Permissions()
		data = append(data, resp)
	}

	httpx.JSON(w, http.StatusOK, httpx.PaginatedResponse[Response]{
		Data:          data,
		Skip:          skip,
		Limit:         limit,
		FilteredCount: total,
		TotalCount:    total,
	})
}
