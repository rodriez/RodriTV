package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/rodriez/RodriTV/internal/channel"
)

const healthzPingTimeout = 2 * time.Second

// Pinger es la dependencia de salud de la base de datos; lo satisface
// *pgxpool.Pool y permite testear /healthz sin una DB real.
type Pinger interface {
	Ping(ctx context.Context) error
}

// registerRoutes registra las rutas del ServeMux. channels es el handler
// REST del dominio de canales y logos el de GET /v1/logos/{filename}
// (channel.FSLogoStore); si alguno es nil sus rutas no se registran.
func registerRoutes(mux *http.ServeMux, pool Pinger, channels *channel.Handler, logos http.Handler, log *slog.Logger) {
	mux.HandleFunc("GET /healthz", healthz(pool, log))
	if channels != nil {
		mux.HandleFunc("GET /v1/channels", channels.List)
		mux.HandleFunc("POST /v1/channels", channels.Create)
		mux.HandleFunc("GET /v1/channels/{id}", channels.Get)
		mux.HandleFunc("PUT /v1/channels/{id}", channels.Update)
		mux.HandleFunc("DELETE /v1/channels/{id}", channels.Delete)
	}
	if logos != nil {
		mux.Handle("GET /v1/logos/{filename}", logos)
	}
}

func healthz(pool Pinger, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthzPingTimeout)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			log.WarnContext(ctx, "healthz: database ping failed", "err", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
