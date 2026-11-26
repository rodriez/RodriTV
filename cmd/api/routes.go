package main

import (
	"net/http"

	"github.com/rodriez/RodriTV/internal/health"
)

func initRoutes() {
	http.HandleFunc("GET /healthz", health.Handler)

	http.HandleFunc("GET /api/v1/channels", tvChannelHandler.ListChannels)
	http.HandleFunc("POST /api/v1/channels", tvChannelHandler.AddChannel)
	http.HandleFunc("GET /api/v1/channels/{channel_id}", tvChannelHandler.GetChannelByID)
	http.HandleFunc("PUT /api/v1/channels/{channel_id}", tvChannelHandler.ReplaceChannel)
	http.HandleFunc("DELETE /api/v1/channels/{channel_id}", tvChannelHandler.DeleteChannel)
}
