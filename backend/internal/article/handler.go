// Package article is the article module — the template's reference implementation
// of the synthesis architecture. The module is split into:
//
//   - domain/      pure core: model, status state machine, policy, publish decision
//   - app/         application: operations (write) + queries (read) behind a facade
//   - adapters/    PostgreSQL implementation of the app's Store port
//   - articleapi/  the published response contract
//
// This parent package is the HTTP boundary (handlers + routes + request DTOs) and
// the composition wiring (NewHandler). Authorization happens here, at the
// entrypoint, never inside the application or the adapters.
package article

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/article/adapters"
	"github.com/yourorg/goapp/internal/article/app"
	"github.com/yourorg/goapp/internal/article/articleapi"
	"github.com/yourorg/goapp/internal/article/domain"
	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Handler exposes the article HTTP endpoints. Authorization is performed here, at
// the entrypoint, never inside the application or adapters.
type Handler struct {
	mod    *app.Module
	logger *slog.Logger
}

// NewHandler wires the module: a PostgreSQL adapter behind the application facade.
func NewHandler(pool *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{
		mod:    app.NewModule(adapters.NewStore(pool)),
		logger: logger,
	}
}

func (h *Handler) List(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	skip, limit := httpx.Pagination(c)
	filter := parseListFilter(c)

	articles, err := h.mod.List(c.Request.Context(), filter, skip, limit)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	filteredCount, err := h.mod.Count(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	totalCount, err := h.mod.Count(c.Request.Context(), app.ListFilter{})
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}

	data := make([]articleapi.Response, 0, len(articles))
	for _, a := range articles {
		resp := articleapi.ToResponse(a)
		resp.Permissions = domain.NewPolicy(actor, a).Permissions()
		data = append(data, resp)
	}

	httpx.JSON(c, http.StatusOK, httpx.PaginatedResponse[articleapi.Response]{
		Data:          data,
		Skip:          skip,
		Limit:         limit,
		FilteredCount: filteredCount,
		TotalCount:    totalCount,
	})
}

func (h *Handler) Get(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	id, err := parseIDParam(c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.mod.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	policy := domain.NewPolicy(actor, a)
	if err := authz.Authorize(policy.CanView()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := articleapi.ToResponse(a)
	resp.Permissions = policy.Permissions()
	httpx.Data(c, http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	if err := authz.Authorize(domain.NewPolicy(actor, domain.Article{}).CanCreate()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	req, err := httpx.DecodeValid[CreateArticleRequest](c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.mod.Create(c.Request.Context(), actor.ID, req.Title, req.Content)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := articleapi.ToResponse(a)
	resp.Permissions = domain.NewPolicy(actor, a).Permissions()
	httpx.Data(c, http.StatusCreated, resp)
}

func (h *Handler) Update(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	id, err := parseIDParam(c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.mod.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := authz.Authorize(domain.NewPolicy(actor, a).CanEdit()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	req, err := httpx.DecodeValid[UpdateArticleRequest](c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	updated, err := h.mod.Update(c.Request.Context(), id, req.Title, req.Content)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := articleapi.ToResponse(updated)
	resp.Permissions = domain.NewPolicy(actor, updated).Permissions()
	httpx.Data(c, http.StatusOK, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	id, err := parseIDParam(c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.mod.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := authz.Authorize(domain.NewPolicy(actor, a).CanDelete()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := h.mod.Delete(c.Request.Context(), id); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	httpx.Data(c, http.StatusOK, "Article deleted successfully")
}

// Publish runs the multi-step publish operation.
func (h *Handler) Publish(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	id, err := parseIDParam(c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.mod.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := authz.Authorize(domain.NewPolicy(actor, a).CanPublish()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	updated, err := h.mod.Publish(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := articleapi.ToResponse(updated)
	resp.Permissions = domain.NewPolicy(actor, updated).Permissions()
	httpx.Data(c, http.StatusOK, resp)
}

func parseListFilter(c *gin.Context) app.ListFilter {
	filter := app.ListFilter{}
	if s := c.Query("status"); s != "" {
		status := domain.Status(s)
		filter.Status = &status
	}
	if a := c.Query("author_id"); a != "" {
		if id, err := strconv.ParseInt(a, 10, 64); err == nil {
			filter.AuthorID = &id
		}
	}
	if s := c.Query("search"); s != "" {
		filter.Search = &s
	}
	return filter
}

func parseIDParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, apperror.BadRequest("invalid_id", "Invalid resource id")
	}
	return id, nil
}
