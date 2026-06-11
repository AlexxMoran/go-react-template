package user

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
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

// RegisterRoutes mounts the user routes. Listing the directory requires
// authentication (any signed-in user); tighten with authz.RequireRoleMW if you
// want it admin-only.
func (h *Handler) RegisterRoutes(r gin.IRouter, requireAuth gin.HandlerFunc) {
	r.Use(requireAuth)
	r.GET("/", h.List)
}

func (h *Handler) List(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	skip, limit := httpx.Pagination(c)

	users, err := h.queries.List(c.Request.Context(), skip, limit)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	total, err := h.queries.Count(c.Request.Context())
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}

	data := make([]Response, 0, len(users))
	for _, u := range users {
		resp := ToResponse(u)
		resp.Permissions = NewPolicy(actor, u).Permissions()
		data = append(data, resp)
	}

	httpx.JSON(c, http.StatusOK, httpx.PaginatedResponse[Response]{
		Data:          data,
		Skip:          skip,
		Limit:         limit,
		FilteredCount: total,
		TotalCount:    total,
	})
}
