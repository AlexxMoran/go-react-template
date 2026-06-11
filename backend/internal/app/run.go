package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/yourorg/goapp/internal/auth"
	"github.com/yourorg/goapp/internal/notifications"
	"github.com/yourorg/goapp/internal/platform/cache"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/database"
	"github.com/yourorg/goapp/internal/platform/jobs"
	"github.com/yourorg/goapp/internal/platform/logger"
	"github.com/yourorg/goapp/internal/platform/mail"
)

// Run loads configuration, opens the database pool, starts the HTTP server and
// the background-job worker, and blocks until ctx is cancelled (e.g.
// SIGINT/SIGTERM), then shuts everything down gracefully. It is the single
// testable entrypoint called from main.
func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Env, cfg.Log.Level)

	pool, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	// Platform services: job client (enqueue), mailer (transport) and cache.
	jobClient := jobs.NewClient(cfg.Redis)
	defer func() { _ = jobClient.Close() }()

	mailer := mail.New(cfg.Mail, log)

	articleCache := cache.New(cfg.Cache, cfg.Redis)
	if closer, ok := articleCache.(io.Closer); ok {
		defer func() { _ = closer.Close() }()
	}

	// notifications provides the concrete implementation of auth.Notifier.
	notifier := notifications.NewEnqueuer(jobClient)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr(),
		Handler:      NewServer(cfg, log, pool, notifier, articleCache),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	// Background worker pool + periodic scheduler. Each module registers its
	// tasks here; the composition root owns the wiring.
	var jobServer *jobs.Server
	if cfg.Jobs.Enabled {
		jobServer = jobs.NewServer(cfg.Redis, cfg.Jobs, log)
		if err := auth.RegisterJobs(jobServer, pool, log); err != nil {
			return fmt.Errorf("register auth jobs: %w", err)
		}
		notifications.RegisterJobs(jobServer, mailer, log)
		if err := jobServer.Start(); err != nil {
			return fmt.Errorf("start job server: %w", err)
		}
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", cfg.HTTP.Addr(), "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		if jobServer != nil {
			jobServer.Stop()
		}
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	// Stop accepting/processing background work before draining HTTP.
	if jobServer != nil {
		jobServer.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("server stopped cleanly")
	return nil
}
