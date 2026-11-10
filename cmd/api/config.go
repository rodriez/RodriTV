package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port            int
	DatabaseURL     string
	DataDir         string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func LoadConfig() (Config, error) {
	v := viper.New()
	v.SetEnvPrefix("RODRITV_API")
	v.AutomaticEnv()

	v.SetDefault("PORT", 9291)
	v.SetDefault("DATA_DIR", "/data")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("SHUTDOWN_TIMEOUT", 10*time.Second)

	for _, key := range []string{"PORT", "DATABASE_URL", "DATA_DIR", "LOG_LEVEL", "SHUTDOWN_TIMEOUT"} {
		if err := v.BindEnv(key); err != nil {
			return Config{}, fmt.Errorf("bind env %s: %w", key, err)
		}
	}

	cfg := Config{
		Port:            v.GetInt("PORT"),
		DatabaseURL:     v.GetString("DATABASE_URL"),
		DataDir:         v.GetString("DATA_DIR"),
		LogLevel:        v.GetString("LOG_LEVEL"),
		ShutdownTimeout: v.GetDuration("SHUTDOWN_TIMEOUT"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", c.Port)
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be debug, info, warn or error, got %q", c.LogLevel)
	}
	return nil
}
