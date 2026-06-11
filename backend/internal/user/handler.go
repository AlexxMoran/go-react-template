package user

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/internal/user/userapi"
)

// Handler exposes the user HTTP endpoints (the authenticated user directory).
type Handler struct {
	mod    *Module
	logger *slog.Logger
}

func NewHandler(mod *Module, logger *slog.Logger) *Handler {
	return &Handler{mod: mod, logger: logger}
}

// RegisterRoutes mounts the user routes. Listing the directory requires
// authentication (any signed-in user); tighten with authz.RequireRole if you
// want it admin-only.
func (h *Handler) RegisterRoutes(r gin.IRouter, requireAuth gin.HandlerFunc) {
	r.Use(requireAuth)
	r.GET("/", h.List)
}

func (h *Handler) List(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	skip, limit := httpx.Pagination(c)

	users, err := h.mod.List(c.Request.Context(), skip, limit)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	total, err := h.mod.Count(c.Request.Context())
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}

	data := make([]userapi.Response, 0, len(users))
	for _, u := range users {
		resp := userapi.ToResponse(u)
		resp.Permissions = userapi.Permissions(actor, u)
		data = append(data, resp)
	}

	httpx.JSON(c, http.StatusOK, httpx.PaginatedResponse[userapi.Response]{
		Data:          data,
		Skip:          skip,
		Limit:         limit,
		FilteredCount: total,
		TotalCount:    total,
	})
}
