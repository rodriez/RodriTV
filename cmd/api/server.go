package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
)

type ctxKeyRequestID struct{}

// errNotFound y errConflict son placeholders del mapeo uniforme: las
// definiciones definitivas de los errores de dominio llegan con US-0002
// (internal/channel) y estos casos pasarán a matchear contra ellas.
var (
	errNotFound = errors.New("not found")
	errConflict = errors.New("conflict")
)

// validationError abstrae el error tipado de validación del dominio
// (ErrValidation{Field, Msg} de US-0002) para el mapeo con errors.As.
type validationError interface {
	error
	ValidationFields() map[string]string
}

// NewServer arma el http.Server sobre un ServeMux con las rutas registradas,
// envuelto por el middleware compuesto: requestID -> logging -> recoverPanic.
func NewServer(cfg Config, log *slog.Logger, pool Pinger, logos http.Handler) *http.Server {
	mux := http.NewServeMux()
	registerRoutes(mux, pool, logos, log)

	handler := requestID(logging(log)(recoverPanic(log)(mux)))

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			reqID, _ := r.Context().Value(ctxKeyRequestID{}).(string)
			log.InfoContext(r.Context(), "request",
				"request_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

func recoverPanic(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(r.Context(), "panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
					)
					writeError(w, fmt.Errorf("panic: %v", rec))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

// writeError aplica el mapeo uniforme de errores: validación -> 400,
// not found -> 404, conflicto -> 409, cualquier otro -> 500.
func writeError(w http.ResponseWriter, err error) {
	var verr validationError

	status := http.StatusInternalServerError
	body := errorResponse{Error: errorBody{
		Code:    "internal",
		Message: "internal server error",
	}}

	switch {
	case errors.As(err, &verr):
		status = http.StatusBadRequest
		body.Error = errorBody{Code: "validation", Message: verr.Error(), Fields: verr.ValidationFields()}
	case errors.Is(err, errNotFound):
		status = http.StatusNotFound
		body.Error = errorBody{Code: "not_found", Message: err.Error()}
	case errors.Is(err, errConflict):
		status = http.StatusConflict
		body.Error = errorBody{Code: "conflict", Message: err.Error()}
	}

	writeJSON(w, status, body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
