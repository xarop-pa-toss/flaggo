// Package config loads runtime configuration from environment variables
// (12-factor). All values have sensible defaults except DATABASE_URL,
// which is required once storage is wired (Phase 2).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the application configuration.
type Config struct {
	Port            string
	LogLevel        string
	Store           string
	DatabaseURL     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// Load reads configuration from the environment.
func Load() (Config, error) {
	cfg := Config{
		Port:            envOr("PORT", "8080"),
		LogLevel:        strings.ToLower(envOr("LOG_LEVEL", "info")),
		Store:           strings.ToLower(envOr("STORE", "postgres")),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		ShutdownTimeout: durOr("SHUTDOWN_TIMEOUT", 10*time.Second),
		ReadTimeout:     durOr("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    durOr("WRITE_TIMEOUT", 15*time.Second),
	}
	if cfg.Port == "" {
		return Config{}, fmt.Errorf("PORT must not be empty")
	}
	switch cfg.Store {
	case "postgres", "memory":
	default:
		return Config{}, fmt.Errorf("STORE must be 'postgres' or 'memory', got %q", cfg.Store)
	}
	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durOr(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		// fall back to seconds if a bare number is given
		if n, err2 := strconv.Atoi(v); err2 == nil {
			return time.Duration(n) * time.Second
		}
		return def
	}
	return d
}
