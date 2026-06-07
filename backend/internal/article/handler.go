package article

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	skip, limit := httpx.Pagination(r)
	filter := parseListFilter(r)

	articles, err := h.search.List(r.Context(), filter, skip, limit)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	filteredCount, err := h.search.Count(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	totalCount, err := h.search.Count(r.Context(), ListFilter{})
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}

	data := make([]Response, 0, len(articles))
	for _, a := range articles {
		resp := ToResponse(a)
		resp.Permissions = NewPolicy(actor, a).Permissions()
		data = append(data, resp)
	}

	httpx.JSON(w, http.StatusOK, httpx.PaginatedResponse[Response]{
		Data:          data,
		Skip:          skip,
		Limit:         limit,
		FilteredCount: filteredCount,
		TotalCount:    totalCount,
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	a, err := h.service.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	policy := NewPolicy(actor, a)
	if err := authz.Authorize(policy.CanView()); err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	resp := ToResponse(a)
	resp.Permissions = policy.Permissions()
	httpx.Data(w, http.StatusOK, resp)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	if err := authz.Authorize(NewPolicy(actor, Article{}).CanCreate()); err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	req, err := httpx.DecodeValid[CreateArticleRequest](r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	a, err := h.service.Create(r.Context(), actor.ID, req.Title, req.Content)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	resp := ToResponse(a)
	resp.Permissions = NewPolicy(actor, a).Permissions()
	httpx.Data(w, http.StatusCreated, resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	a, err := h.service.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	if err := authz.Authorize(NewPolicy(actor, a).CanEdit()); err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	req, err := httpx.DecodeValid[UpdateArticleRequest](r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	updated, err := h.service.Update(r.Context(), id, req.Title, req.Content)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	resp := ToResponse(updated)
	resp.Permissions = NewPolicy(actor, updated).Permissions()
	httpx.Data(w, http.StatusOK, resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	a, err := h.service.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	if err := authz.Authorize(NewPolicy(actor, a).CanDelete()); err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	httpx.Data(w, http.StatusOK, "Article deleted successfully")
}

// Publish runs the multi-step publish operation.
func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	a, err := h.service.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	if err := authz.Authorize(NewPolicy(actor, a).CanPublish()); err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	row, err := publish.NewScenario(h.pool).Run(r.Context(), publish.Contract{ArticleID: id})
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	updated := fromGen(row)
	resp := ToResponse(updated)
	resp.Permissions = NewPolicy(actor, updated).Permissions()
	httpx.Data(w, http.StatusOK, resp)
}

func parseListFilter(r *http.Request) ListFilter {
	q := r.URL.Query()
	filter := ListFilter{}
	if s := q.Get("status"); s != "" {
		status := Status(s)
		filter.Status = &status
	}
	if a := q.Get("author_id"); a != "" {
		if id, err := strconv.ParseInt(a, 10, 64); err == nil {
			filter.AuthorID = &id
		}
	}
	if s := q.Get("search"); s != "" {
		filter.Search = &s
	}
	return filter
}

func parseIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, apperror.BadRequest("invalid_id", "Invalid resource id")
	}
	return id, nil
}
