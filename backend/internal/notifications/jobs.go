package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/yourorg/goapp/internal/platform/jobs"
	"github.com/yourorg/goapp/internal/platform/mail"
)

const taskWelcomeEmail = "notifications:welcome_email"

type welcomePayload struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

// Enqueuer is the producer side of notifications. Its methods match the consumer
// ports other modules declare (e.g. auth.Notifier), so the composition root can
// inject it without those modules importing this one.
type Enqueuer struct {
	client *jobs.Client
}

// NewEnqueuer builds the notifications producer over the shared job client.
func NewEnqueuer(client *jobs.Client) *Enqueuer {
	return &Enqueuer{client: client}
}

// NotifyWelcome enqueues a welcome email for a newly registered user.
func (e *Enqueuer) NotifyWelcome(ctx context.Context, userID int64, email, firstName string) error {
	payload, err := json.Marshal(welcomePayload{UserID: userID, Email: email, FirstName: firstName})
	if err != nil {
		return err
	}
	return e.client.Enqueue(ctx, asynq.NewTask(taskWelcomeEmail, payload), asynq.MaxRetry(5))
}

// RegisterJobs wires the notifications worker handlers into the job server.
func RegisterJobs(srv *jobs.Server, mailer mail.Mailer, logger *slog.Logger) {
	srv.HandleFunc(taskWelcomeEmail, handleWelcome(mailer, logger))
}

func handleWelcome(mailer mail.Mailer, logger *slog.Logger) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p welcomePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			// A malformed payload will never succeed — don't retry it.
			return fmt.Errorf("%w: welcome payload unmarshal: %w", asynq.SkipRetry, err)
		}
		if err := mailer.Send(ctx, renderWelcome(p.FirstName, p.Email)); err != nil {
			return fmt.Errorf("send welcome email: %w", err)
		}
		logger.Info("welcome email sent",
			slog.Int64("user_id", p.UserID),
			slog.String("email", p.Email))
		return nil
	}
}
