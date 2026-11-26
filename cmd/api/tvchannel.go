package main

import (
	"github.com/rodriez/RodriTV/internal/sqlc"
	"github.com/rodriez/RodriTV/internal/tvchannels"
)

func initTvChannelHandler() {
	db := sqlc.New(dbpool)
	tvChannelRepository = tvchannels.NewRepository(db)
	tvChannelService = tvchannels.NewService(&tvChannelRepository)
	tvChannelHandler = tvchannels.NewHandler(&tvChannelService)
}
