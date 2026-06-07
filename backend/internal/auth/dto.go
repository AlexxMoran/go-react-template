package auth

import (
	"context"
	"strings"
)

const minPasswordLength = 8

// RegisterRequest is the body of POST /auth/register.
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (r RegisterRequest) Valid(_ context.Context) map[string]string {
	problems := map[string]string{}
	if !looksLikeEmail(r.Email) {
		problems["email"] = "must be a valid email address"
	}
	if len(r.Password) < minPasswordLength {
		problems["password"] = "must be at least 8 characters"
	}
	return problems
}

// LoginRequest is the body of POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Valid(_ context.Context) map[string]string {
	problems := map[string]string{}
	if strings.TrimSpace(r.Email) == "" {
		problems["email"] = "is required"
	}
	if r.Password == "" {
		problems["password"] = "is required"
	}
	return problems
}

// UpdateProfileRequest is the body of PATCH /auth/me.
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (r UpdateProfileRequest) Valid(_ context.Context) map[string]string {
	return map[string]string{}
}

// TokenResponse is returned on login/refresh. The refresh token is delivered
// separately as an httpOnly cookie, never in the body.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func looksLikeEmail(s string) bool {
	s = strings.TrimSpace(s)
	at := strings.IndexByte(s, '@')
	return at > 0 && at < len(s)-1 && strings.IndexByte(s[at+1:], '.') >= 0
}
