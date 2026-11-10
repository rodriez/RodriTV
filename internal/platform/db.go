package platform

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pingTimeout  = 30 * time.Second
	initialWait  = 250 * time.Millisecond
	maxWait      = 5 * time.Second
	pingMaxConns = 10
	pingMinConns = 2
)

// NewPool crea un pool pgx con MinConns=2 y MaxConns=10, y espera a que la
// base de datos responda a Ping con backoff exponencial (250ms inicial, x2,
// tope 5s) hasta un maximo de 30s. Agotado el tiempo, devuelve error.
func NewPool(ctx context.Context, dsn string, log *slog.Logger) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}
	cfg.MinConns = pingMinConns
	cfg.MaxConns = pingMaxConns

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	deadline := time.Now().Add(pingTimeout)
	wait := initialWait
	for {
		if err := pool.Ping(ctx); err == nil {
			return pool, nil
		} else {
			log.WarnContext(ctx, "database ping failed, retrying", "wait", wait, "err", err)
		}

		if time.Now().Add(wait).After(deadline) {
			pool.Close()
			return nil, fmt.Errorf("database not available after %s: %w", pingTimeout, errors.New("ping retries exhausted"))
		}

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("database not available: %w", ctx.Err())
		case <-time.After(wait):
		}

		wait = min(wait*2, maxWait)
	}
}
