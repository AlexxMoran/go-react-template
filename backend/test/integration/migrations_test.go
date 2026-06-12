//go:build integration

package integration

import (
	"context"
	"testing"
)

func TestMigrations_CreateExpectedTables(t *testing.T) {
	tables := []string{"users", "refresh_tokens", "articles"}

	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			var exists bool
			err := pool.QueryRow(context.Background(), `
SELECT EXISTS (
	SELECT 1
	FROM information_schema.tables
	WHERE table_schema = 'public' AND table_name = $1
)`, table).Scan(&exists)
			if err != nil {
				t.Fatalf("query table existence: %v", err)
			}
			if !exists {
				t.Fatalf("table %q does not exist after migrations", table)
			}
		})
	}
}
