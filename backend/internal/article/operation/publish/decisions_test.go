package publish

import (
	"testing"
	"time"

	"github.com/yourorg/goapp/pkg/apperror"
)

// These tests exercise the business rules with no database and no HTTP — the
// payoff of keeping the decision layer pure.
func TestDecisions_Make_PublishesDraft(t *testing.T) {
	now := time.Now()
	got, err := Decisions{}.Make(Snapshot{
		ArticleID:  7,
		Status:     statusDraft,
		Title:      "Hello",
		HasContent: true,
	}, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.NewStatus != statusPublished {
		t.Errorf("NewStatus = %q, want %q", got.NewStatus, statusPublished)
	}
	if !got.PublishedAt.Equal(now) {
		t.Errorf("PublishedAt = %v, want %v", got.PublishedAt, now)
	}
	if got.ArticleID != 7 {
		t.Errorf("ArticleID = %d, want 7", got.ArticleID)
	}
}

func TestDecisions_Make_RejectsNonDraft(t *testing.T) {
	_, err := Decisions{}.Make(Snapshot{
		Status:     statusPublished,
		Title:      "Hello",
		HasContent: true,
	}, time.Now())
	assertMessageKey(t, err, "invalid_status_transition")
}

func TestDecisions_Make_RejectsEmptyContent(t *testing.T) {
	_, err := Decisions{}.Make(Snapshot{
		Status:     statusDraft,
		Title:      "Hello",
		HasContent: false,
	}, time.Now())
	assertMessageKey(t, err, "validation_error")
}

func assertMessageKey(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with message_key %q, got nil", want)
	}
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("expected *apperror.Error, got %T", err)
	}
	if appErr.MessageKey != want {
		t.Errorf("message_key = %q, want %q", appErr.MessageKey, want)
	}
}
