package article

import (
	"context"

	"github.com/yourorg/goapp/internal/platform/database/gen"
)

// Service holds the simple, single-step CRUD behavior of the article domain.
// Per the operation-oriented rule, anything multi-step (e.g. publish) lives in
// operation/ instead. Services never perform authorization — that happens at the
// HTTP entrypoint before the service is called.
type Service struct {
	repo    *Repository
	queries *Queries
}

func NewService(db gen.DBTX) *Service {
	return &Service{repo: NewRepository(db), queries: NewQueries(db)}
}

func (s *Service) Get(ctx context.Context, id int64) (Article, error) {
	return s.queries.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, authorID int64, title, content string) (Article, error) {
	return s.repo.Create(ctx, authorID, title, content)
}

func (s *Service) Update(ctx context.Context, id int64, title, content string) (Article, error) {
	return s.repo.Update(ctx, id, title, content)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
