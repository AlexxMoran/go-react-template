package app_test

import (
	"context"
	"testing"

	"github.com/yourorg/goapp/internal/article/app"
	"github.com/yourorg/goapp/internal/article/domain"
	"github.com/yourorg/goapp/pkg/apperror"
)

// fakeStore is an in-memory app.Store for fast, DB-free unit tests of the
// application layer. Because the write side depends on a port (app.Store) rather
// than a concrete database, the publish operation and its pure decision can be
// exercised together without Postgres. Only the methods used by these tests do
// real work; the rest satisfy the interface.
type fakeStore struct {
	snapshot domain.PublishSnapshot // facts RunPublish hands to the decider
	applied  *domain.PublishDecision
}

func (f *fakeStore) Get(context.Context, int64) (domain.Article, error) {
	return domain.Article{}, nil
}

func (f *fakeStore) List(context.Context, app.ListFilter, int, int) ([]domain.Article, error) {
	return nil, nil
}

func (f *fakeStore) Count(context.Context, app.ListFilter) (int64, error) { return 0, nil }

func (f *fakeStore) Create(context.Context, int64, string, string) (domain.Article, error) {
	return domain.Article{}, nil
}

func (f *fakeStore) Update(context.Context, int64, string, string) (domain.Article, error) {
	return domain.Article{}, nil
}

func (f *fakeStore) Delete(context.Context, int64) error { return nil }

func (f *fakeStore) RunPublish(
	_ context.Context,
	_ int64,
	decide func(domain.PublishSnapshot) (domain.PublishDecision, error),
) (domain.Article, error) {
	decision, err := decide(f.snapshot)
	if err != nil {
		return domain.Article{}, err
	}
	f.applied = &decision
	return domain.Article{
		ID:          decision.ArticleID,
		Status:      decision.NewStatus,
		PublishedAt: &decision.PublishedAt,
	}, nil
}

func TestModule_Publish_PublishesDraft(t *testing.T) {
	store := &fakeStore{snapshot: domain.PublishSnapshot{
		ArticleID:  42,
		Status:     domain.StatusDraft,
		Title:      "Hello",
		HasContent: true,
	}}
	mod := app.NewModule(store)

	got, err := mod.Publish(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != domain.StatusPublished {
		t.Errorf("status = %q, want %q", got.Status, domain.StatusPublished)
	}
	if store.applied == nil || store.applied.NewStatus != domain.StatusPublished {
		t.Errorf("expected the published decision to be applied, got %+v", store.applied)
	}
	if got.PublishedAt == nil {
		t.Error("expected PublishedAt to be set")
	}
}

func TestModule_Publish_RejectsNonDraft(t *testing.T) {
	store := &fakeStore{snapshot: domain.PublishSnapshot{
		ArticleID:  42,
		Status:     domain.StatusPublished, // already published → invariant rejects
		Title:      "Hello",
		HasContent: true,
	}}
	mod := app.NewModule(store)

	_, err := mod.Publish(context.Background(), 42)
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("expected *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.MessageKey != "invalid_status_transition" {
		t.Errorf("message_key = %q, want %q", appErr.MessageKey, "invalid_status_transition")
	}
	if store.applied != nil {
		t.Errorf("no decision should have been applied, got %+v", store.applied)
	}
}
