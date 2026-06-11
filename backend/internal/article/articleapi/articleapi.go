// Package articleapi is the article module's published contract: the JSON
// response shape that the HTTP boundary and any future consumer module use. It
// depends only on the pure domain model — never on the application or adapters —
// so importing it never drags in the database driver.
package articleapi

import (
	"time"

	"github.com/yourorg/goapp/internal/article/domain"
)

// Response is the public representation of an article.
type Response struct {
	ID          int64           `json:"id"`
	AuthorID    int64           `json:"author_id"`
	Title       string          `json:"title"`
	Content     string          `json:"content"`
	Status      string          `json:"status"`
	PublishedAt *time.Time      `json:"published_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Permissions map[string]bool `json:"permissions,omitempty"`
}

// ToResponse maps the domain model to its API representation. Permissions are
// attached by the caller (the HTTP handler) from the domain policy.
func ToResponse(a domain.Article) Response {
	return Response{
		ID:          a.ID,
		AuthorID:    a.AuthorID,
		Title:       a.Title,
		Content:     a.Content,
		Status:      string(a.Status),
		PublishedAt: a.PublishedAt,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
