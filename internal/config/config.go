// Package config reads the application settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every setting the server needs at startup.
type Config struct {
	Port        int
	DatabaseURL string
	LogLevel    string
	SessionTTL  time.Duration
}

// FromEnv reads the configuration, applying defaults for optional values.
func FromEnv() (Config, error) {
	return parse(os.Getenv)
}

func parse(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:        8080,
		DatabaseURL: getenv("DATABASE_URL"),
		LogLevel:    valueOr(getenv("LOG_LEVEL"), "info"),
		SessionTTL:  30 * 24 * time.Hour,
	}
	if v := getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p <= 0 || p > 65535 {
			return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535, got %q", v)
		}
		cfg.Port = p
	}
	if v := getenv("SESSION_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("SESSION_TTL must be a positive duration, got %q", v)
		}
		cfg.SessionTTL = d
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
