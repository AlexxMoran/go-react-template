//go:build integration

package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the database/sql "pgx" driver for goose
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Postgres is a running PostgreSQL container with the schema applied, paired with
// a pgx pool connected to it.
type Postgres struct {
	Pool      *pgxpool.Pool
	container *postgres.PostgresContainer
}

// StartPostgres boots a disposable PostgreSQL container, applies all goose
// migrations and returns a ready-to-use pool. Call Close to tear it down (the
// integration suite does this once per package from TestMain).
func StartPostgres(ctx context.Context) (*Postgres, error) {
	container, err := postgres.Run(ctx, "postgres:17",
		postgres.WithDatabase("goapp_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("connection string: %w", err)
	}

	if err := migrate(dsn); err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("migrate: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("open pool: %w", err)
	}

	return &Postgres{Pool: pool, container: container}, nil
}

// Close releases the pool and terminates the container.
func (p *Postgres) Close(ctx context.Context) {
	if p.Pool != nil {
		p.Pool.Close()
	}
	if p.container != nil {
		_ = p.container.Terminate(ctx)
	}
}

// migrate applies the project's goose migrations to a fresh database.
func migrate(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, migrationsDir())
}

// migrationsDir resolves the repo's db/migrations directory from the module root,
// so the suite works regardless of the test's working directory.
func migrationsDir() string {
	return filepath.Join(repoRoot(), "db", "migrations")
}

// repoRoot walks up from the working directory to the backend module root (the
// directory containing go.mod).
func repoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}
