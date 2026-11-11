package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rodriez/RodriTV/internal/channel"
	"github.com/rodriez/RodriTV/internal/platform"
	"github.com/rodriez/RodriTV/migrations"
)

// TestChannelsAPI_EndToEnd levanta PostgreSQL con testcontainers y ejercita
// el flujo completo crear -> listar -> obtener -> actualizar -> eliminar
// contra el mux real de la API, incluyendo el reemplazo total de sources
// en PUT y el formato uniforme de error.
func TestChannelsAPI_EndToEnd(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("rodritv_test"),
		postgres.WithUsername("rodritv"),
		postgres.WithPassword("rodritv"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, platform.MigrateUp(dsn, migrations.FS))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	repo := channel.NewPostgresRepository(pool)
	svc := channel.NewService(repo, func() time.Time { return time.Now().UTC() })
	logos := channel.NewFSLogoStore(t.TempDir())
	svc.SetLogoStore(logos)
	handler := channel.NewHandler(svc, writeError)

	mux := http.NewServeMux()
	registerRoutes(mux, pool, handler, logos, slog.Default())
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	doJSON := func(t *testing.T, method, path string, payload any, wantStatus int) map[string]any {
		t.Helper()

		var body bytes.Buffer
		if payload != nil {
			require.NoError(t, json.NewEncoder(&body).Encode(payload))
		}

		req, err := http.NewRequest(method, srv.URL+path, &body)
		require.NoError(t, err)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, wantStatus, resp.StatusCode)

		var decoded map[string]any
		if resp.StatusCode != http.StatusNoContent {
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
		}

		return decoded
	}

	createPayload := map[string]any{
		"number": 7,
		"name":   "Canal 7",
		"sources": []any{
			map[string]any{"slot": 1, "url": "http://stream.example.com/a.m3u8", "enabled": true},
			nil,
			map[string]any{"slot": 3, "url": "http://stream.example.com/b.m3u8", "enabled": true},
		},
	}

	created := doJSON(t, http.MethodPost, "/v1/channels", createPayload, http.StatusCreated)
	id, ok := created["id"].(string)
	require.True(t, ok)
	require.NotEqual(t, uuid.Nil, uuid.MustParse(id))
	assert.Equal(t, float64(7), created["number"])
	assert.Equal(t, float64(channel.DefaultSegmentSeconds), created["segment_seconds"])
	assert.Equal(t, true, created["enabled"])

	sources, ok := created["sources"].([]any)
	require.True(t, ok)
	require.Len(t, sources, channel.SourceSlots)
	sourceID := sources[0].(map[string]any)["id"].(string)
	require.NotEmpty(t, sourceID)

	list := doJSON(t, http.MethodGet, "/v1/channels", nil, http.StatusOK)
	channels, ok := list["channels"].([]any)
	require.True(t, ok)
	assert.Len(t, channels, 1)

	got := doJSON(t, http.MethodGet, "/v1/channels/"+id, nil, http.StatusOK)
	assert.Equal(t, created, got)

	// PUT reemplaza totalmente el array de sources: el source del slot 3 no
	// se envía y debe desaparecer; el del slot 1 conserva su id.
	updatePayload := map[string]any{
		"number": 7,
		"name":   "Canal 7 HD",
		"sources": []any{
			map[string]any{"id": sourceID, "slot": 1, "url": "http://stream.example.com/a2.m3u8", "enabled": true},
		},
	}
	updated := doJSON(t, http.MethodPut, "/v1/channels/"+id, updatePayload, http.StatusOK)
	assert.Equal(t, "Canal 7 HD", updated["name"])

	updatedSources, ok := updated["sources"].([]any)
	require.True(t, ok)
	require.Len(t, updatedSources, channel.SourceSlots)
	assert.Equal(t, sourceID, updatedSources[0].(map[string]any)["id"])
	for i := 1; i < channel.SourceSlots; i++ {
		assert.Nil(t, updatedSources[i], "slot %d debió quedar vacío tras el PUT", i+1)
	}

	doJSON(t, http.MethodDelete, "/v1/channels/"+id, nil, http.StatusNoContent)

	missing := doJSON(t, http.MethodGet, "/v1/channels/"+id, nil, http.StatusNotFound)
	errBody, ok := missing["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "not_found", errBody["code"])

	conflict := doJSON(t, http.MethodPost, "/v1/channels", map[string]any{"number": 0, "name": ""}, http.StatusBadRequest)
	conflictErr, ok := conflict["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "validation", conflictErr["code"])
}
