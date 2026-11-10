package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	for input, want := range cases {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, want, parseLogLevel(input))
		})
	}
}

func TestParseLogLevel_UnknownDefaultsToInfo(t *testing.T) {
	assert.Equal(t, slog.LevelInfo, parseLogLevel("bogus"))
}

func TestRunServer_GracefulShutdownOnContextCancel(t *testing.T) {
	srv := &http.Server{Addr: "127.0.0.1:0", Handler: http.NewServeMux()}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- runServer(ctx, srv, 5*time.Second, newTestLogger(&bytes.Buffer{}))
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("runServer did not return after context cancel")
	}
}

func TestRunServer_StartErrorPropagates(t *testing.T) {
	srv := &http.Server{Addr: "invalid-addr", Handler: http.NewServeMux()}

	err := runServer(context.Background(), srv, 5*time.Second, newTestLogger(&bytes.Buffer{}))

	require.Error(t, err)
}
