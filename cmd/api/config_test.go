package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_MissingDatabaseURL(t *testing.T) {
	_, err := LoadConfig()
	require.Error(t, err)
}

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("RODRITV_API_DATABASE_URL", "postgres://localhost/rodritv")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 9291, cfg.Port)
	assert.Equal(t, "/data", cfg.DataDir)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
}

func TestLoadConfig_PortFromEnv(t *testing.T) {
	t.Setenv("RODRITV_API_DATABASE_URL", "postgres://localhost/rodritv")
	t.Setenv("RODRITV_API_PORT", "1234")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 1234, cfg.Port)
}

func TestConfig_Validate(t *testing.T) {
	valid := Config{
		Port:            9291,
		DatabaseURL:     "postgres://localhost/rodritv",
		DataDir:         "/data",
		LogLevel:        "info",
		ShutdownTimeout: 10 * time.Second,
	}

	t.Run("missing database url", func(t *testing.T) {
		cfg := valid
		cfg.DatabaseURL = ""
		require.Error(t, cfg.Validate())
	})

	t.Run("port out of range", func(t *testing.T) {
		for _, port := range []int{0, -1, 65536} {
			cfg := valid
			cfg.Port = port
			require.Error(t, cfg.Validate())
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		cfg := valid
		cfg.LogLevel = "trace"
		require.Error(t, cfg.Validate())
	})

	t.Run("valid", func(t *testing.T) {
		require.NoError(t, valid.Validate())
	})
}
