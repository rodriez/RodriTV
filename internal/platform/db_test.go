package platform

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewPool_InvalidDSN(t *testing.T) {
	pool, err := NewPool(context.Background(), "not-a-valid-dsn", testLogger())

	require.Error(t, err)
	assert.Nil(t, pool)
}

func TestNewPool_ReachableDB(t *testing.T) {
	dsn := os.Getenv("RODRITV_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("RODRITV_TEST_DATABASE_URL not set")
	}

	pool, err := NewPool(context.Background(), dsn, testLogger())
	require.NoError(t, err)
	defer pool.Close()

	var one int
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT 1").Scan(&one))
	assert.Equal(t, 1, one)
}

func TestNewPool_UnreachableDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 30s retry test in short mode")
	}

	start := time.Now()
	pool, err := NewPool(context.Background(), "postgres://user:pass@127.0.0.1:1/db?connect_timeout=1", testLogger())
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Nil(t, pool)
	assert.GreaterOrEqual(t, elapsed, 25*time.Second)
	assert.Less(t, elapsed, 40*time.Second)
}

func TestNewPool_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, "postgres://user:pass@127.0.0.1:1/db?connect_timeout=1", testLogger())

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, pool)
}
