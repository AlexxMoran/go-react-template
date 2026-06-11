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

// pool is the shared connection to the suite's PostgreSQL container. It is
// started once for the whole package by TestMain; individual tests isolate
// themselves with testsupport.Truncate.
var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := testsupport.StartPostgres(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "integration setup failed:", err)
		os.Exit(1)
	}
	pool = pg.Pool

	code := m.Run()

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
