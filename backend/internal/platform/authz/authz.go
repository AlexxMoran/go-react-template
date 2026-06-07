// Package authz holds the authorization primitives shared by every domain:
// the Actor (the authenticated identity carried in the request context), the
// RBAC Role type, and helpers that turn a policy decision into an error.
//
// Design (mirrors the Python Pundit-style policy layer):
//   - Each domain defines its own policy struct with Can<Action>() bool methods
//     and a Permissions() map[string]bool (the latter is surfaced to the
//     frontend so the UI can show/hide actions).
//   - Handlers (the only place authorization happens) call the policy and pass
//     the boolean to authz.Authorize, which raises a 403 when denied.
//   - Role-gated routes use the RequireRole middleware.
package authz

import (
	"context"

	"github.com/yourorg/goapp/pkg/apperror"
)

// Role is an RBAC role. Roles are coarse, application-wide grants; fine-grained,
// per-record checks (ownership etc.) live in domain policies.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Actor is the authenticated identity, built from the access-token claims. It is
// intentionally small: policies only need the id and role plus a couple of
// flags, so no database lookup is required per request.
type Actor struct {
	ID    int64
	Email string
	Role  Role
}

// IsAdmin reports whether the actor holds the admin role.
func (a *Actor) IsAdmin() bool { return a != nil && a.Role == RoleAdmin }

// Owns reports whether the actor owns a resource with the given owner id.
func (a *Actor) Owns(ownerID int64) bool { return a != nil && a.ID == ownerID }

type ctxKey int

const actorKey ctxKey = iota

// WithActor returns a copy of ctx carrying the actor.
func WithActor(ctx context.Context, actor *Actor) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}

// ActorFrom returns the actor stored in ctx, if any.
func ActorFrom(ctx context.Context) (*Actor, bool) {
	actor, ok := ctx.Value(actorKey).(*Actor)
	return actor, ok && actor != nil
}

// Authorize converts a policy decision into an error. It returns a 403 when the
// action is not allowed and nil otherwise — the Go equivalent of Pundit's
// `policy.deny()`.
func Authorize(allowed bool) error {
	if !allowed {
		return apperror.Forbidden("forbidden", "You do not have permission to perform this action")
	}
	return nil
}

// RequireRole returns nil only if the actor holds the required role (admin
// always satisfies the check).
func RequireRole(actor *Actor, role Role) error {
	if actor == nil {
		return apperror.Unauthorized("unauthorized", "Authentication required")
	}
	if actor.Role == role || actor.IsAdmin() {
		return nil
	}
	return apperror.Forbidden("forbidden", "You do not have permission to perform this action")
}
