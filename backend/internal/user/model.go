// Package user is the user domain: the domain model, its read (queries) and
// write (repository) database access, its authorization policy, and its API
// DTOs. Authentication and token issuance live in the separate auth package.
package user

import (
	"time"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
)

// User is the domain representation of an account. HashedPassword stays inside
// the backend and is never serialized into an API response (see dto.go).
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

// fromGen maps the sqlc row type to the domain model.
func fromGen(row gen.User) User {
	return User{
		ID:             row.ID,
		Email:          row.Email,
		HashedPassword: row.HashedPassword,
		Role:           authz.Role(row.Role),
		FirstName:      row.FirstName,
		LastName:       row.LastName,
		IsActive:       row.IsActive,
		IsVerified:     row.IsVerified,
		CreatedAt:      database.TimeOrZero(row.CreatedAt),
		UpdatedAt:      database.TimeOrZero(row.UpdatedAt),
	}
}

// AsActor builds the lightweight authorization identity for this user.
func (u User) AsActor() *authz.Actor {
	return &authz.Actor{ID: u.ID, Email: u.Email, Role: u.Role}
}
