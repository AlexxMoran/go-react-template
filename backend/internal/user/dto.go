package user

import "time"

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

// ToResponse maps the domain model to its API representation.
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
