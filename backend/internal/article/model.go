// Package article is the article domain. It demonstrates the full layering used
// for a non-trivial resource: domain model + status state machine, read
// (queries/search) and write (repository) database access, an authorization
// policy, simple CRUD in service.go, and a multi-step business operation under
// operation/ (see operation/publish).
package article

import (
	"time"

	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
)

// Article is the domain model.
type Article struct {
	ID          int64
	AuthorID    int64
	Title       string
	Content     string
	Status      Status
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// fromGen maps the sqlc row type to the domain model.
func fromGen(row gen.Article) Article {
	return Article{
		ID:          row.ID,
		AuthorID:    row.AuthorID,
		Title:       row.Title,
		Content:     row.Content,
		Status:      Status(row.Status),
		PublishedAt: database.TimePtr(row.PublishedAt),
		CreatedAt:   database.TimeOrZero(row.CreatedAt),
		UpdatedAt:   database.TimeOrZero(row.UpdatedAt),
	}
}
