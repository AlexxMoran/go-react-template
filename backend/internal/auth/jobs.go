package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/internal/platform/jobs"
)

// taskPurgeTokens is auth's own maintenance job.
const taskPurgeTokens = "auth:purge_refresh_tokens" //nolint:gosec // G101: task type name, not a credential

// Notifier is the consumer port auth uses to trigger user-facing notifications
// without knowing how they are composed or delivered. The notifications module
// implements it; the composition root injects the implementation. This keeps
// email/templating out of auth entirely.
type Notifier interface {
	NotifyWelcome(ctx context.Context, userID int64, email, firstName string) error
}

// RegisterJobs wires auth's background work: a periodic refresh-token purge.
func RegisterJobs(srv *jobs.Server, pool *pgxpool.Pool, logger *slog.Logger) error {
	srv.HandleFunc(taskPurgeTokens, handlePurgeTokens(pool, logger))

	// Purge expired/revoked refresh tokens hourly so the table stays bounded.
	return srv.RegisterPeriodic("@every 1h", asynq.NewTask(taskPurgeTokens, nil))
}

func handlePurgeTokens(pool *pgxpool.Pool, logger *slog.Logger) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		deleted, err := gen.New(pool).DeleteExpiredRefreshTokens(ctx)
		if err != nil {
			return fmt.Errorf("purge refresh tokens: %w", err)
		}
		logger.Info("purged refresh tokens", slog.Int64("deleted", deleted))
		return nil
	}
}
