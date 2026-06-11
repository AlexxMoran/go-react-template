// Package jobs is the background-job runtime: a thin wrapper over asynq that
// keeps the broker (Redis) at a single boundary. The Client enqueues tasks; the
// Server runs the worker pool and the periodic scheduler. Modules define their
// own task types and handlers in a jobs.go file and register them here — the
// arch guardrail TestAsynqStaysAtJobsBoundary keeps asynq out of domain code.
//
// Note: enqueueing is not transactional with the database. Producers enqueue
// after their write commits (best-effort); for exactly-once delivery tied to a
// DB change, add a transactional outbox later.
package jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/yourorg/goapp/internal/platform/config"
)

func redisOpt(cfg config.RedisConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB}
}

// Client enqueues tasks onto the queue. It is always available, even when the
// worker server is disabled.
type Client struct {
	c *asynq.Client
}

// NewClient builds an enqueue-only client.
func NewClient(cfg config.RedisConfig) *Client {
	return &Client{c: asynq.NewClient(redisOpt(cfg))}
}

// Enqueue submits a task for asynchronous processing.
func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) error {
	_, err := c.c.EnqueueContext(ctx, task, opts...)
	return err
}

// Close releases the client's Redis connections.
func (c *Client) Close() error { return c.c.Close() }

// Server runs the worker pool (asynq.Server) and the periodic scheduler. Register
// handlers and periodic tasks before calling Start.
type Server struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	scheduler *asynq.Scheduler
	logger    *slog.Logger
}

// NewServer builds the worker server and scheduler with the given concurrency.
func NewServer(cfg config.RedisConfig, jobsCfg config.JobsConfig, logger *slog.Logger) *Server {
	opt := redisOpt(cfg)
	adapter := slogAdapter{logger}
	return &Server{
		server: asynq.NewServer(opt, asynq.Config{
			Concurrency: jobsCfg.Concurrency,
			Logger:      adapter,
		}),
		mux:       asynq.NewServeMux(),
		scheduler: asynq.NewScheduler(opt, &asynq.SchedulerOpts{Logger: adapter}),
		logger:    logger,
	}
}

// HandleFunc registers the handler for a task type (e.g. "auth:purge_tokens").
func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

// RegisterPeriodic schedules a task to be enqueued on a cron spec (asynq cron
// syntax, e.g. "@every 1h" or "0 3 * * *").
func (s *Server) RegisterPeriodic(cronspec string, task *asynq.Task, opts ...asynq.Option) error {
	_, err := s.scheduler.Register(cronspec, task, opts...)
	return err
}

// Start launches the worker pool and the scheduler. It is non-blocking.
func (s *Server) Start() error {
	if err := s.server.Start(s.mux); err != nil {
		return fmt.Errorf("start worker server: %w", err)
	}
	if err := s.scheduler.Start(); err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}
	s.logger.Info("job server started")
	return nil
}

// Stop gracefully shuts down the scheduler (stops enqueueing) and then the worker
// pool (drains in-flight tasks).
func (s *Server) Stop() {
	s.scheduler.Shutdown()
	s.server.Shutdown()
	s.logger.Info("job server stopped")
}

// slogAdapter adapts *slog.Logger to asynq.Logger so job logs match the app's
// structured format.
type slogAdapter struct{ l *slog.Logger }

func (s slogAdapter) Debug(args ...any) { s.l.Debug(fmt.Sprint(args...)) }
func (s slogAdapter) Info(args ...any)  { s.l.Info(fmt.Sprint(args...)) }
func (s slogAdapter) Warn(args ...any)  { s.l.Warn(fmt.Sprint(args...)) }
func (s slogAdapter) Error(args ...any) { s.l.Error(fmt.Sprint(args...)) }
func (s slogAdapter) Fatal(args ...any) { s.l.Error(fmt.Sprint(args...)) }
