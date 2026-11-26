package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func initPoolDB(ctx context.Context, conf *Config) {
	var err error

	dbpool, err = pgxpool.New(ctx, conf.DatabaseURL)
	if err != nil {
		log.Fatalf("fail to create db pool: %v", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		log.Fatalf("fail to connect postgres: %v", err)
	}
}
