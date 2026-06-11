// Package domain is the article module's pure core: the model, the status state
// machine, the authorization policy and the publish decision. It performs no I/O
// — no database driver, no HTTP, no router — so every rule here is unit-testable
// without a database. The arch guardrail TestDomainIsPure enforces that purity.
package domain

import "time"

// Article is the domain model.
type Article struct {
	ID          int64
	AuthorID    int64
	Title       string
	Content     string
	Status      Status
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
