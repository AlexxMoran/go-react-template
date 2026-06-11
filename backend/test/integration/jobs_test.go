//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/internal/platform/jobs"
	"github.com/yourorg/goapp/internal/platform/logger"
	"github.com/yourorg/goapp/internal/user"
	"github.com/yourorg/goapp/internal/user/userapi"
	"github.com/yourorg/goapp/test/testsupport"
)

// TestPurgeExpiredRefreshTokens verifies the cleanup query behind the periodic
// job: expired and revoked tokens are deleted, valid ones are kept.
func TestPurgeExpiredRefreshTokens(t *testing.T) {
	testsupport.Truncate(t, pool)
	ctx := context.Background()
	q := gen.New(pool)

	u, err := user.New(pool).Create(ctx, userapi.CreateParams{
		Email: "tokens@example.com", HashedPassword: "x", Role: authz.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	createToken(t, q, u.ID, "valid-hash", time.Now().Add(time.Hour))
	createToken(t, q, u.ID, "expired-hash", time.Now().Add(-time.Hour))
	createToken(t, q, u.ID, "revoked-hash", time.Now().Add(time.Hour))
	if err := q.RevokeRefreshToken(ctx, "revoked-hash"); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	deleted, err := q.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if deleted != 2 {
		t.Errorf("deleted = %d, want 2 (expired + revoked)", deleted)
	}

	assertTokenExists(t, q, "valid-hash", true)
	assertTokenExists(t, q, "expired-hash", false)
	assertTokenExists(t, q, "revoked-hash", false)
}

// TestJobs_EnqueueAndProcess exercises the real asynq runtime end-to-end: a task
// enqueued via jobs.Client is picked up and run by a jobs.Server worker.
func TestJobs_EnqueueAndProcess(t *testing.T) {
	redis := config.RedisConfig{Addr: redisAddr}
	log := logger.New("test", "error")

	srv := jobs.NewServer(redis, config.JobsConfig{Concurrency: 2}, log)
	processed := make(chan string, 1)
	srv.HandleFunc("test:ping", func(_ context.Context, task *asynq.Task) error {
		processed <- string(task.Payload())
		return nil
	})
	if err := srv.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer srv.Stop()

	client := jobs.NewClient(redis)
	defer func() { _ = client.Close() }()

	if err := client.Enqueue(context.Background(), asynq.NewTask("test:ping", []byte("pong"))); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	select {
	case got := <-processed:
		if got != "pong" {
			t.Errorf("payload = %q, want pong", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("task was not processed within 10s")
	}
}

func createToken(t *testing.T, q *gen.Queries, userID int64, hash string, expires time.Time) {
	t.Helper()
	_, err := q.CreateRefreshToken(context.Background(), gen.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: database.Timestamptz(expires),
	})
	if err != nil {
		t.Fatalf("create token %q: %v", hash, err)
	}
}

func assertTokenExists(t *testing.T, q *gen.Queries, hash string, want bool) {
	t.Helper()
	_, err := q.GetRefreshToken(context.Background(), hash)
	switch {
	case want && err != nil:
		t.Errorf("token %q: expected to exist, got error %v", hash, err)
	case !want && !errors.Is(err, pgx.ErrNoRows):
		t.Errorf("token %q: expected to be gone, got err=%v", hash, err)
	}
}
