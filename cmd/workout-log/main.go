package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // user time zones work without zoneinfo on the host

	"github.com/sorin-work/workout-log/internal/api"
	"github.com/sorin-work/workout-log/internal/config"
	"github.com/sorin-work/workout-log/internal/database"
	"github.com/sorin-work/workout-log/internal/seed"
	"github.com/sorin-work/workout-log/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}
	logger.Info("database ready")
	if cfg.SeedDemo {
		created, err := seed.Demo(ctx, pool, time.Now())
		if err != nil {
			return fmt.Errorf("seed demo data: %w", err)
		}
		if created {
			logger.Info("created demo account", "email", seed.DemoEmail)
		}
	}

	st := store.New(pool)
	go cleanupSessions(ctx, st, logger)

	handler := api.NewServer(st, api.Options{
		SessionTTL:    cfg.SessionTTL,
		SecureCookies: cfg.SecureCookies,
		WebDir:        cfg.WebDir,
	})
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	errc := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", srv.Addr, "version", api.Version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
		close(errc)
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func newLogger(level string) *slog.Logger {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}

// cleanupSessions deletes expired sessions and closes workouts that were
// left open, once an hour until ctx ends.
func cleanupSessions(ctx context.Context, st *store.Store, logger *slog.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		n, err := st.DeleteExpiredSessions(ctx)
		if err != nil && ctx.Err() == nil {
			logger.Warn("delete expired sessions", "err", err)
		} else if n > 0 {
			logger.Info("deleted expired sessions", "count", n)
		}
		closed, err := st.CloseStaleWorkouts(ctx, 12*time.Hour)
		if err != nil && ctx.Err() == nil {
			logger.Warn("close stale workouts", "err", err)
		} else if closed > 0 {
			logger.Info("closed workouts left open", "count", closed)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
