// Package userapi is the user module's published contract: the only thing other
// modules are allowed to import from user. It carries the cross-module DTOs
// (User, CreateParams), the public JSON representation (Response/ToResponse) and
// the frontend permission projection (Permissions). The module's internals
// (repository, queries, the Module facade) live in the parent package and stay
// private to the module — the arch guardrail TestModuleBoundaries enforces that
// consumers reach the user module only through this package.
package userapi

import (
	"time"

	"github.com/yourorg/goapp/internal/platform/authz"
)

// User is the canonical user value passed across the module boundary. HashedPassword
// is included because the auth module verifies credentials against it; it is
// backend-only and is never serialized (see Response, which omits it).
type User struct {
	ID             int64
	Email          string
	HashedPassword string
	Role           authz.Role
	FirstName      string
	LastName       string
	IsActive       bool
	IsVerified     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreateParams holds the fields required to create an account. A zero Role is
// defaulted to authz.RoleUser by the module.
type CreateParams struct {
	Email          string
	HashedPassword string
	Role           authz.Role
	FirstName      string
	LastName       string
}

// Response is the public representation of a user. It deliberately omits the
// hashed password and any other sensitive fields.
type Response struct {
	ID          int64           `json:"id"`
	Email       string          `json:"email"`
	Role        string          `json:"role"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	IsActive    bool            `json:"is_active"`
	IsVerified  bool            `json:"is_verified"`
	CreatedAt   time.Time       `json:"created_at"`
	Permissions map[string]bool `json:"permissions,omitempty"`
}

// ToResponse maps the cross-module model to its API representation.
func ToResponse(u User) Response {
	return Response{
		ID:         u.ID,
		Email:      u.Email,
		Role:       string(u.Role),
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		IsActive:   u.IsActive,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt,
	}
}

// Permissions returns the record-level permission map surfaced to the frontend.
// The actor may act on its own record; admins may act on anyone. This mirrors
// the Python UserPolicy.
func Permissions(actor *authz.Actor, u User) map[string]bool {
	canManage := actor != nil && (actor.IsAdmin() || actor.Owns(u.ID))
	return map[string]bool{
		"view": canManage,
		"edit": canManage,
	}
}
