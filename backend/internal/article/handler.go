package article

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/article/operation/publish"
	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Handler exposes the article HTTP endpoints. Authorization is performed here, at
// the entrypoint, never inside services or operations.
type Handler struct {
	service *Service
	search  *Search
	pool    *pgxpool.Pool
	logger  *slog.Logger
}

func NewHandler(pool *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{
		service: NewService(pool),
		search:  NewSearch(pool),
		pool:    pool,
		logger:  logger,
	}
}

func (h *Handler) List(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	skip, limit := httpx.Pagination(c)
	filter := parseListFilter(c)

	articles, err := h.search.List(c.Request.Context(), filter, skip, limit)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	filteredCount, err := h.search.Count(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	totalCount, err := h.search.Count(c.Request.Context(), ListFilter{})
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}

	data := make([]Response, 0, len(articles))
	for _, a := range articles {
		resp := ToResponse(a)
		resp.Permissions = NewPolicy(actor, a).Permissions()
		data = append(data, resp)
	}

	httpx.JSON(c, http.StatusOK, httpx.PaginatedResponse[Response]{
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
	a, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	policy := NewPolicy(actor, a)
	if err := authz.Authorize(policy.CanView()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := ToResponse(a)
	resp.Permissions = policy.Permissions()
	httpx.Data(c, http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	if err := authz.Authorize(NewPolicy(actor, Article{}).CanCreate()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	req, err := httpx.DecodeValid[CreateArticleRequest](c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.service.Create(c.Request.Context(), actor.ID, req.Title, req.Content)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := ToResponse(a)
	resp.Permissions = NewPolicy(actor, a).Permissions()
	httpx.Data(c, http.StatusCreated, resp)
}

func (h *Handler) Update(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	id, err := parseIDParam(c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := authz.Authorize(NewPolicy(actor, a).CanEdit()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	req, err := httpx.DecodeValid[UpdateArticleRequest](c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	updated, err := h.service.Update(c.Request.Context(), id, req.Title, req.Content)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := ToResponse(updated)
	resp.Permissions = NewPolicy(actor, updated).Permissions()
	httpx.Data(c, http.StatusOK, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	id, err := parseIDParam(c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	a, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := authz.Authorize(NewPolicy(actor, a).CanDelete()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
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
	a, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	if err := authz.Authorize(NewPolicy(actor, a).CanPublish()); err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	row, err := publish.NewScenario(h.pool).Run(c.Request.Context(), publish.Contract{ArticleID: id})
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	updated := fromGen(row)
	resp := ToResponse(updated)
	resp.Permissions = NewPolicy(actor, updated).Permissions()
	httpx.Data(c, http.StatusOK, resp)
}

func parseListFilter(c *gin.Context) ListFilter {
	filter := ListFilter{}
	if s := c.Query("status"); s != "" {
		status := Status(s)
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
