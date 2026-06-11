package domain

import (
	"time"

	"github.com/yourorg/goapp/pkg/apperror"
)

// PublishSnapshot is the set of facts the publish decision needs, captured from
// the database by the adapter. It is a plain value type with no driver types so
// the decision below stays pure and testable.
type PublishSnapshot struct {
	ArticleID  int64
	Status     Status
	Title      string
	HasContent bool
}

// PublishDecision is the write plan produced by DecidePublish and applied by the
// adapter inside a transaction. It carries every value the write needs (including
// the timestamp) so applying it is deterministic.
type PublishDecision struct {
	ArticleID   int64
	NewStatus   Status
	PublishedAt time.Time
}

// DecidePublish validates the snapshot and returns the write plan. It performs no
// I/O and takes the current time as an argument, so its output is fully
// determined by its inputs — the payoff of keeping the decision pure.
func DecidePublish(s PublishSnapshot, now time.Time) (PublishDecision, error) {
	if s.Status != StatusDraft {
		return PublishDecision{}, apperror.BadRequest(
			"invalid_status_transition",
			"Only draft articles can be published",
		)
	}
	if s.Title == "" || !s.HasContent {
		return PublishDecision{}, apperror.Validation(
			"Article cannot be published",
			map[string]string{"content": "title and content are required to publish"},
		)
	}
	return PublishDecision{
		ArticleID:   s.ArticleID,
		NewStatus:   StatusPublished,
		PublishedAt: now,
	}, nil
}
