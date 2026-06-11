package domain

import "github.com/yourorg/goapp/internal/platform/authz"

// Policy answers authorization questions about an article. Published articles are
// world-readable, drafts are visible only to their author (and admins), and only
// the author or an admin may edit, delete or publish. CanCreate is record-
// agnostic — any authenticated user may create an article.
type Policy struct {
	actor  *authz.Actor
	record Article
}

func NewPolicy(actor *authz.Actor, record Article) Policy {
	return Policy{actor: actor, record: record}
}

// CanCreate is a global permission (no record needed).
func (p Policy) CanCreate() bool {
	return p.actor != nil
}

func (p Policy) CanView() bool {
	if p.record.Status == StatusPublished {
		return true
	}
	return p.isOwnerOrAdmin()
}

func (p Policy) CanEdit() bool   { return p.isOwnerOrAdmin() }
func (p Policy) CanDelete() bool { return p.isOwnerOrAdmin() }

// CanPublish guards the publish operation. Kept separate from CanEdit so the two
// can diverge later (e.g. an editor role that can publish but not rewrite).
func (p Policy) CanPublish() bool { return p.isOwnerOrAdmin() }

func (p Policy) isOwnerOrAdmin() bool {
	return p.actor != nil && (p.actor.IsAdmin() || p.actor.Owns(p.record.AuthorID))
}

// Permissions returns the record-level permission map surfaced to the frontend.
func (p Policy) Permissions() map[string]bool {
	return map[string]bool{
		"view":    p.CanView(),
		"edit":    p.CanEdit(),
		"delete":  p.CanDelete(),
		"publish": p.CanPublish(),
	}
}
