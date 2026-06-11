//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/goapp/internal/notifications"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/jobs"
	"github.com/yourorg/goapp/internal/platform/logger"
	"github.com/yourorg/goapp/internal/platform/mail"
)

// spyMailer records the messages it is asked to send and signals each on a
// channel, so a test can wait for asynchronous delivery.
type spyMailer struct {
	got chan mail.Message
}

func (s *spyMailer) Send(_ context.Context, m mail.Message) error {
	s.got <- m
	return nil
}

// TestNotifications_WelcomeEmailFlow drives the full producer→queue→consumer→mail
// path: NotifyWelcome enqueues a task that the notifications worker handles by
// composing and sending the welcome email.
func TestNotifications_WelcomeEmailFlow(t *testing.T) {
	redis := config.RedisConfig{Addr: redisAddr}
	log := logger.New("test", "error")
	spy := &spyMailer{got: make(chan mail.Message, 1)}

	srv := jobs.NewServer(redis, config.JobsConfig{Concurrency: 2}, log)
	notifications.RegisterJobs(srv, spy, log)
	if err := srv.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer srv.Stop()

	enq := notifications.NewEnqueuer(jobs.NewClient(redis))
	if err := enq.NotifyWelcome(context.Background(), 1, "ada@example.com", "Ada"); err != nil {
		t.Fatalf("notify: %v", err)
	}

	select {
	case m := <-spy.got:
		if len(m.To) != 1 || m.To[0] != "ada@example.com" {
			t.Errorf("To = %v, want [ada@example.com]", m.To)
		}
		if m.Subject == "" {
			t.Error("expected a non-empty subject")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("welcome email was not sent within 10s")
	}
}
