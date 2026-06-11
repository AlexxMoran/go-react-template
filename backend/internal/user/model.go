// Package user is the user domain. Its public contract (DTOs, response shape,
// permissions) lives in the userapi sub-package — that is the only thing other
// modules may import. This parent package holds the module's internals: the
// read (queries) and write (repository) database access and the Module facade
// that composes them. Authentication and token issuance live in the separate
// auth package, which depends on this module solely through userapi.
package user

import (
	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/internal/user/userapi"
)

// fromGen maps the sqlc row type to the published user model.
func fromGen(row gen.User) userapi.User {
	return userapi.User{
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
