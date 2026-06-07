package article

import (
	"context"

	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

// ListFilter holds the optional filters for listing articles. A nil field means
// "do not filter on this attribute".
type ListFilter struct {
	Status   *Status
	AuthorID *int64
	Search   *string
}

// Search is the read side for filtered, paginated article listings. It mirrors
// the Python BaseSearch: results() + filtered_count() + total_count().
type Search struct {
	q *gen.Queries
}

func NewSearch(db gen.DBTX) *Search {
	return &Search{q: gen.New(db)}
}

func (s *Search) List(ctx context.Context, f ListFilter, skip, limit int) ([]Article, error) {
	rows, err := s.q.ListArticles(ctx, gen.ListArticlesParams{
		Status:   database.TextPtr(statusStr(f.Status)),
		AuthorID: database.Int8Ptr(f.AuthorID),
		Search:   database.TextPtr(f.Search),
		Off:      int32(skip),
		Lim:      int32(limit),
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	articles := make([]Article, 0, len(rows))
	for _, row := range rows {
		articles = append(articles, fromGen(row))
	}
	return articles, nil
}

func (s *Search) Count(ctx context.Context, f ListFilter) (int64, error) {
	total, err := s.q.CountArticles(ctx, gen.CountArticlesParams{
		Status:   database.TextPtr(statusStr(f.Status)),
		AuthorID: database.Int8Ptr(f.AuthorID),
		Search:   database.TextPtr(f.Search),
	})
	if err != nil {
		return 0, apperror.Internal(err)
	}
	return total, nil
}

func statusStr(s *Status) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}
