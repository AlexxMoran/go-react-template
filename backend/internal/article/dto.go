package article

import (
	"context"
	"strings"
	"time"
)

const maxTitleLength = 200

// CreateArticleRequest is the body of POST /articles.
type CreateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (r CreateArticleRequest) Valid(_ context.Context) map[string]string {
	problems := map[string]string{}
	title := strings.TrimSpace(r.Title)
	if title == "" {
		problems["title"] = "is required"
	} else if len(title) > maxTitleLength {
		problems["title"] = "must be at most 200 characters"
	}
	return problems
}

// UpdateArticleRequest is the body of PATCH /articles/{id}.
type UpdateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (r UpdateArticleRequest) Valid(_ context.Context) map[string]string {
	problems := map[string]string{}
	title := strings.TrimSpace(r.Title)
	if title == "" {
		problems["title"] = "is required"
	} else if len(title) > maxTitleLength {
		problems["title"] = "must be at most 200 characters"
	}
	return problems
}

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

// ToResponse maps the domain model to its API representation.
func ToResponse(a Article) Response {
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
