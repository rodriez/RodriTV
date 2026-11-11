package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodriez/RodriTV/internal/channel"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(ctx context.Context) error { return f.err }

func TestLogosRoute_ServesAndValidates(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "3f6b2c1e.png"), []byte("fake-image"), 0o644))

	mux := http.NewServeMux()
	registerRoutes(mux, fakePinger{}, channel.NewFSLogoStore(dir), slog.Default())

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/logos/3f6b2c1e.png", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/logos/evil.exe", nil))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHealthz_OK(t *testing.T) {
	mux := http.NewServeMux()
	registerRoutes(mux, fakePinger{}, nil, slog.Default())

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
}

func TestHealthz_DatabaseDown(t *testing.T) {
	mux := http.NewServeMux()
	registerRoutes(mux, fakePinger{err: errors.New("connection refused")}, nil, slog.Default())

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
}
