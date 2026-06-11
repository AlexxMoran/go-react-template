//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/pkg/apperror"
	"github.com/yourorg/goapp/test/testsupport"
)

// Shared per-package fixtures, started once by TestMain. Tests isolate the
// database with testsupport.Truncate.
var (
	pool      *pgxpool.Pool
	redisAddr string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := testsupport.StartPostgres(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "integration setup failed (postgres):", err)
		os.Exit(1)
	}
	pool = pg.Pool

	rd, err := testsupport.StartRedis(ctx)
	if err != nil {
		pg.Close(ctx)
		fmt.Fprintln(os.Stderr, "integration setup failed (redis):", err)
		os.Exit(1)
	}
	redisAddr = rd.Addr

	code := m.Run()

	rd.Close(ctx)
	pg.Close(ctx)
	os.Exit(code)
}

// assertMessageKey fails unless err is an *apperror.Error carrying want.
func assertMessageKey(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with message_key %q, got nil", want)
	}
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("expected *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.MessageKey != want {
		t.Errorf("message_key = %q, want %q", appErr.MessageKey, want)
	}
}
