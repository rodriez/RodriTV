package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const healthzPingTimeout = 2 * time.Second

// Pinger es la dependencia de salud de la base de datos; lo satisface
// *pgxpool.Pool y permite testear /healthz sin una DB real.
type Pinger interface {
	Ping(ctx context.Context) error
}

// registerRoutes registra las rutas del ServeMux. logos es el handler de
// GET /v1/logos/{filename} (channel.FSLogoStore); si es nil la ruta no se
// registra — el wiring real del directorio de logos llega con T-0013.
func registerRoutes(mux *http.ServeMux, pool Pinger, logos http.Handler, log *slog.Logger) {
	mux.HandleFunc("GET /healthz", healthz(pool, log))
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
