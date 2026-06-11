//go:build integration

package testsupport

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Truncate clears all application tables and resets identity sequences. Call it
// at the start of each test so every test begins from an empty, predictable
// schema. CASCADE handles the foreign keys (articles/refresh_tokens → users).
func Truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`TRUNCATE articles, refresh_tokens, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
