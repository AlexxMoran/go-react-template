package publish

import (
	"time"

	"github.com/yourorg/goapp/pkg/apperror"
)

// These mirror the article.Status constants but are duplicated as plain strings
// on purpose: the decision layer must not import the article domain package
// (which pulls in the database driver), keeping these rules pure and unit-
// testable without a database — the same guarantee enforced on the Python
// decisions layer.
const (
	statusDraft     = "draft"
	statusPublished = "published"
)

// Decisions holds the pure business rules for publishing an article.
type Decisions struct{}

// Make validates the snapshot and returns the write plan. It performs no I/O and
// takes the current time as an argument so its output is fully determined by its
// inputs.
func (Decisions) Make(s Snapshot, now time.Time) (Decision, error) {
	if s.Status != statusDraft {
		return Decision{}, apperror.BadRequest(
			"invalid_status_transition",
			"Only draft articles can be published",
		)
	}
	if s.Title == "" || !s.HasContent {
		return Decision{}, apperror.Validation(
			"Article cannot be published",
			map[string]string{"content": "title and content are required to publish"},
		)
	}
	return Decision{
		ArticleID:   s.ArticleID,
		NewStatus:   statusPublished,
		PublishedAt: now,
	}, nil
}
