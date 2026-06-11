package app

import (
	"context"
	"time"

	"github.com/yourorg/goapp/internal/article/domain"
)

// ── write side (operations) ─────────────────────────────────────────────────────

// Create is a lightweight operation: there is no decision to make, so it is a
// single insert. The "operation" scaffold (load → decide → apply) is only worth
// its weight when there is an invariant to enforce — see Publish.
func (m *Module) Create(ctx context.Context, authorID int64, title, content string) (domain.Article, error) {
	return m.store.Create(ctx, authorID, title, content)
}

// Update replaces the editable fields of an article.
func (m *Module) Update(ctx context.Context, id int64, title, content string) (domain.Article, error) {
	return m.store.Update(ctx, id, title, content)
}

// Delete removes an article.
func (m *Module) Delete(ctx context.Context, id int64) error {
	return m.store.Delete(ctx, id)
}

// Publish is the multi-step operation: load the current facts, apply the pure
// publish decision, and persist the result atomically. The decision
// (domain.DecidePublish) runs inside the adapter's transaction via the callback,
// so it stays free of any database type while the whole thing remains atomic.
func (m *Module) Publish(ctx context.Context, id int64) (domain.Article, error) {
	return m.store.RunPublish(ctx, id, func(s domain.PublishSnapshot) (domain.PublishDecision, error) {
		return domain.DecidePublish(s, time.Now())
	})
}
