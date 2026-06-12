//go:build integration

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/goapp/internal/auth"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/user"
	"github.com/yourorg/goapp/test/testsupport"
)

func newAuthService() *auth.Service {
	jwt := auth.NewJWTManager(config.JWTConfig{
		AccessSecret:  "test-access-secret-test-access-secret",
		RefreshSecret: "test-refresh-secret-test-refresh-secret",
		Issuer:        "goapp-test",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
	})
	// The real auth↔user seam: auth consumes the concrete *user.Module through
	// its port, all against the live database.
	return auth.NewService(pool, jwt, user.New(pool))
}

// TestAuth_RegisterLoginRefreshRotation walks the full credential lifecycle and
// asserts refresh-token rotation: refreshing issues a new token and revokes the
// old one, so reusing the rotated token fails. This only passes if the
// refresh_tokens table is written, hashed and revoked correctly.
func TestAuth_RegisterLoginRefreshRotation(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	svc := newAuthService()

	if _, err := svc.Register(ctx, "user@example.com", "password123", "Ada", "Lovelace"); err != nil {
		t.Fatalf("register: %v", err)
	}

	_, pair, err := svc.Login(ctx, "user@example.com", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected a non-empty token pair")
	}

	newPair, err := svc.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if newPair.RefreshToken == pair.RefreshToken {
		t.Error("refresh token should rotate to a new value")
	}

	if _, err := svc.Refresh(ctx, pair.RefreshToken); err == nil {
		t.Error("expected reuse of the rotated (revoked) refresh token to fail")
	}
}

func TestAuth_ConcurrentRefreshAllowsOnlyOneRotation(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	svc := newAuthService()

	if _, err := svc.Register(ctx, "race@example.com", "password123", "Ada", "Lovelace"); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, pair, err := svc.Login(ctx, "race@example.com", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.Refresh(context.Background(), pair.RefreshToken)
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	successes := 0
	failures := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		failures++
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("concurrent refresh results: successes=%d failures=%d, want 1/1", successes, failures)
	}
}

func TestAuth_LoginWrongPassword(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	svc := newAuthService()

	if _, err := svc.Register(ctx, "user@example.com", "password123", "Ada", ""); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, _, err := svc.Login(ctx, "user@example.com", "wrong-password")
	assertMessageKey(t, err, "invalid_credentials")
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	svc := newAuthService()

	if _, err := svc.Register(ctx, "dup@example.com", "password123", "A", ""); err != nil {
		t.Fatalf("first register: %v", err)
	}
	_, err := svc.Register(ctx, "dup@example.com", "password123", "B", "")
	assertMessageKey(t, err, "email_taken")
}
