// Command seed inserts demo data for local development: an admin user, a regular
// user and a couple of articles. It is idempotent — users are matched by email
// and articles are only seeded into an empty table — so it is safe to re-run.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yourorg/goapp/internal/article/adapters"
	"github.com/yourorg/goapp/internal/article/app"
	"github.com/yourorg/goapp/internal/auth"
	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/user"
	"github.com/yourorg/goapp/internal/user/userapi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	users := user.New(pool)

	admin, err := ensureUser(ctx, users, "admin@example.com", "password123", authz.RoleAdmin, "Admin", "User")
	if err != nil {
		return err
	}
	if _, err := ensureUser(ctx, users, "user@example.com", "password123", authz.RoleUser, "Regular", "User"); err != nil {
		return err
	}

	store := adapters.NewStore(pool)
	count, err := store.Count(ctx, app.ListFilter{})
	if err != nil {
		return fmt.Errorf("count articles: %w", err)
	}
	if count == 0 {
		for _, a := range []struct{ title, content string }{
			{"Welcome to goapp", "This is a seeded article you can edit or delete."},
			{"Second article", "More seeded content for the list view."},
		} {
			if _, err := store.Create(ctx, admin.ID, a.title, a.content); err != nil {
				return fmt.Errorf("create article %q: %w", a.title, err)
			}
		}
	}

	fmt.Println("seed complete:")
	fmt.Println("  admin@example.com / password123 (admin)")
	fmt.Println("  user@example.com  / password123 (user)")
	return nil
}

// ensureUser returns the existing user with the email, or creates it.
func ensureUser(ctx context.Context, users *user.Module, email, password string, role authz.Role, first, last string) (userapi.User, error) {
	if existing, err := users.GetByEmail(ctx, email); err == nil {
		return existing, nil
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return userapi.User{}, err
	}
	return users.Create(ctx, userapi.CreateParams{
		Email:          email,
		HashedPassword: hash,
		Role:           role,
		FirstName:      first,
		LastName:       last,
	})
}
