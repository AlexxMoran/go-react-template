package article

import (
	"context"
	"strings"
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
