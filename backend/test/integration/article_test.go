//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/yourorg/goapp/internal/article/adapters"
	"github.com/yourorg/goapp/internal/article/app"
	"github.com/yourorg/goapp/internal/article/domain"
	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/user"
	"github.com/yourorg/goapp/internal/user/userapi"
	"github.com/yourorg/goapp/test/testsupport"
)

func newArticleModule() *app.Module {
	return app.NewModule(adapters.NewStore(pool))
}

// createAuthor inserts a user to satisfy the articles.author_id foreign key.
func createAuthor(t *testing.T) userapi.User {
	t.Helper()
	u, err := user.New(pool).Create(context.Background(), userapi.CreateParams{
		Email:          "author@example.com",
		HashedPassword: "fixture", //nolint:gosec // G101: test fixture, not a real secret
		Role:           authz.RoleUser,
		FirstName:      "Test",
	})
	if err != nil {
		t.Fatalf("create author: %v", err)
	}
	return u
}

func TestArticle_CreateGetList(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	mod := newArticleModule()
	author := createAuthor(t)

	created, err := mod.Create(ctx, author.ID, "Hello", "World")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Status != domain.StatusDraft {
		t.Errorf("new article status = %q, want draft", created.Status)
	}

	got, err := mod.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Hello" || got.AuthorID != author.ID {
		t.Errorf("got %+v, want title=Hello author=%d", got, author.ID)
	}

	list, err := mod.List(ctx, app.ListFilter{}, 0, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
	total, err := mod.Count(ctx, app.ListFilter{})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 1 {
		t.Errorf("count = %d, want 1", total)
	}
}

func TestArticle_GetMissing(t *testing.T) {
	testsupport.Truncate(t, pool)
	mod := newArticleModule()

	_, err := mod.Get(context.Background(), 999)
	assertMessageKey(t, err, "not_found")
}

// TestArticle_PublishLifecycle exercises the full multi-step operation against a
// real database and transaction: a draft is published, then re-publishing is
// rejected by the pure decision — which only holds if the first publish actually
// persisted.
func TestArticle_PublishLifecycle(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	mod := newArticleModule()
	author := createAuthor(t)

	draft, err := mod.Create(ctx, author.ID, "Title", "Body")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	published, err := mod.Publish(ctx, draft.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Status != domain.StatusPublished {
		t.Errorf("status = %q, want published", published.Status)
	}
	if published.PublishedAt == nil {
		t.Error("PublishedAt should be set after publish")
	}

	// The change must be durable: a fresh read sees the published state.
	reloaded, err := mod.Get(ctx, draft.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Status != domain.StatusPublished {
		t.Errorf("reloaded status = %q, want published", reloaded.Status)
	}

	// Re-publishing a non-draft is rejected by the domain decision.
	_, err = mod.Publish(ctx, draft.ID)
	assertMessageKey(t, err, "invalid_status_transition")
}

func TestArticle_PublishRequiresContent(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	mod := newArticleModule()
	author := createAuthor(t)

	draft, err := mod.Create(ctx, author.ID, "Title", "") // empty content
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = mod.Publish(ctx, draft.ID)
	assertMessageKey(t, err, "validation_error")

	// And the rejected operation must not have changed the row (tx rolled back).
	reloaded, err := mod.Get(ctx, draft.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Status != domain.StatusDraft {
		t.Errorf("status = %q, want draft (rollback)", reloaded.Status)
	}
}

func TestArticle_FilterByStatus(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	mod := newArticleModule()
	author := createAuthor(t)

	if _, err := mod.Create(ctx, author.ID, "draft one", "body"); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	toPublish, err := mod.Create(ctx, author.ID, "to publish", "body")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := mod.Publish(ctx, toPublish.ID); err != nil {
		t.Fatalf("publish: %v", err)
	}

	pub := domain.StatusPublished
	publishedCount, err := mod.Count(ctx, app.ListFilter{Status: &pub})
	if err != nil {
		t.Fatalf("count published: %v", err)
	}
	if publishedCount != 1 {
		t.Errorf("published count = %d, want 1", publishedCount)
	}

	total, err := mod.Count(ctx, app.ListFilter{})
	if err != nil {
		t.Fatalf("count all: %v", err)
	}
	if total != 2 {
		t.Errorf("total count = %d, want 2", total)
	}
}
