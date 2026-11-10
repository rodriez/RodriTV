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

	"github.com/rodriez/RodriTV/internal/platform"
	"github.com/rodriez/RodriTV/migrations"
)

func main() {
	os.Exit(run())
}

// run compone el binario: config -> logger -> pool -> migraciones -> server,
// y devuelve el exit code del proceso (0 OK, 1 error de arranque).
func run() int {
	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}

	log := newLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := platform.NewPool(ctx, cfg.DatabaseURL, log)
	if err != nil {
		log.Error("connect database", "err", err)
		return 1
	}
	defer pool.Close()

	if err := platform.MigrateUp(cfg.DatabaseURL, migrations.FS); err != nil {
		log.Error("apply migrations", "err", err)
		return 1
	}

	srv := NewServer(cfg, log, pool, nil)

	log.Info("server starting", "port", cfg.Port)
	if err := runServer(ctx, srv, cfg.ShutdownTimeout, log); err != nil {
		log.Error("server stopped with error", "err", err)
		return 1
	}

	log.Info("server stopped gracefully")
	return 0
}

func newLogger(cfg Config) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// runServer arranca el servidor en una goroutine y bloquea hasta que el
// contexto se cancele (señal) o el servidor falle. Al cancelarse, hace
// Shutdown drenando conexiones en vuelo dentro del timeout.
func runServer(ctx context.Context, srv *http.Server, timeout time.Duration, log *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		log.Info("shutdown signal received, draining connections", "timeout", timeout)
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	}
}
