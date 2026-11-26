package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodriez/RodriTV/internal/tvchannels"
)

var (
	conf                Config
	dbpool              *pgxpool.Pool
	tvChannelRepository tvchannels.TvChannelsStoreRepository
	tvChannelService    tvchannels.Service
	tvChannelHandler    tvchannels.Handler
)

func main() {
	ctx := context.Background()

	initConfig()
	initPoolDB(ctx, &conf)
	initTvChannelHandler()
	initRoutes()

	defer dbpool.Close()

	fmt.Println("Initializing RodriTV api sever...")

	log.Fatal(http.ListenAndServe(":9290", nil))
}
