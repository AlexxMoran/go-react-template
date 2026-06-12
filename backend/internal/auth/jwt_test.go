package auth

import (
	"testing"
	"time"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/user/userapi"
)

func testJWTManager() *JWTManager {
	return NewJWTManager(config.JWTConfig{
		AccessSecret:  "test-access-secret-test-access-secret",
		RefreshSecret: "test-refresh-secret-test-refresh-secret",
		Issuer:        "goapp-test",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Hour,
	})
}

func TestJWTManagerAccessRoundTrip(t *testing.T) {
	manager := testJWTManager()

	token, err := manager.GenerateAccess(userapi.User{
		ID:    42,
		Email: "ada@example.com",
		Role:  authz.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("generate access: %v", err)
	}

	claims, err := manager.ParseAccess(token)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	actor, err := claims.Actor()
	if err != nil {
		t.Fatalf("actor: %v", err)
	}
	if actor.ID != 42 || actor.Email != "ada@example.com" || actor.Role != authz.RoleAdmin {
		t.Fatalf("actor = %+v, want id/email/admin", actor)
	}
}

func TestJWTManagerRejectsRefreshTokenAsAccessToken(t *testing.T) {
	manager := testJWTManager()

	token, _, err := manager.GenerateRefresh(42)
	if err != nil {
		t.Fatalf("generate refresh: %v", err)
	}

	if _, err := manager.ParseAccess(token); err == nil {
		t.Fatal("expected refresh token to be rejected as an access token")
	}
}

func TestJWTManagerRefreshRoundTrip(t *testing.T) {
	manager := testJWTManager()

	token, expiresAt, err := manager.GenerateRefresh(42)
	if err != nil {
		t.Fatalf("generate refresh: %v", err)
	}
	if time.Until(expiresAt) <= 0 {
		t.Fatalf("refresh expiry should be in the future: %v", expiresAt)
	}

	userID, err := manager.ParseRefresh(token)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	if userID != 42 {
		t.Fatalf("userID = %d, want 42", userID)
	}
}
