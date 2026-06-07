package user

import "github.com/yourorg/goapp/internal/platform/authz"

// Policy answers authorization questions about a single user record. It mirrors
// the Python UserPolicy: the actor may act on its own record, and admins may act
// on anyone.
type Policy struct {
	actor  *authz.Actor
	record User
}

func NewPolicy(actor *authz.Actor, record User) Policy {
	return Policy{actor: actor, record: record}
}

func (p Policy) CanView() bool {
	return p.actor != nil && (p.actor.IsAdmin() || p.actor.Owns(p.record.ID))
}

func (p Policy) CanEdit() bool {
	return p.actor != nil && (p.actor.IsAdmin() || p.actor.Owns(p.record.ID))
}

// Permissions returns the record-level permission map surfaced to the frontend.
func (p Policy) Permissions() map[string]bool {
	return map[string]bool{
		"view": p.CanView(),
		"edit": p.CanEdit(),
	}
}
