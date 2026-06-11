// Package app is the article module's application layer. Under the synthesis
// architecture every write is an operation (a command) and every read is a query,
// both exposed through the Module facade. The application orchestrates; the pure
// rules live in domain and all I/O sits behind the Store port, implemented by the
// adapters package. Authorization is never performed here — it happens at the
// HTTP entrypoint before a method is called.
package app

import (
	"context"

	"github.com/yourorg/goapp/internal/article/domain"
)

// ListFilter holds the optional filters for listing articles. A nil field means
// "do not filter on this attribute".
type ListFilter struct {
	Status   *domain.Status
	AuthorID *int64
	Search   *string
}

// Store is the data port the application depends on. Reads and single-statement
// writes are direct; the multi-step publish operation uses RunPublish, which owns
// the transaction and calls the pure decider back mid-transaction so the decision
// never sees the database.
type Store interface {
	Get(ctx context.Context, id int64) (domain.Article, error)
	List(ctx context.Context, f ListFilter, skip, limit int) ([]domain.Article, error)
	Count(ctx context.Context, f ListFilter) (int64, error)
	Create(ctx context.Context, authorID int64, title, content string) (domain.Article, error)
	Update(ctx context.Context, id int64, title, content string) (domain.Article, error)
	Delete(ctx context.Context, id int64) error
	RunPublish(ctx context.Context, id int64, decide func(domain.PublishSnapshot) (domain.PublishDecision, error)) (domain.Article, error)
}

// Module is the article module's application facade: a single concrete entry
// point exposing the module's commands and queries.
type Module struct {
	store Store
}

// NewModule builds the application over the given data port.
func NewModule(store Store) *Module {
	return &Module{store: store}
}
