package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr, DatabasePath                      string
	SessionTTL, WorkerInterval, ShutdownTimeout time.Duration
	WorkerBatch                                 int
}

func Load() (Config, error) {
	cfg := Config{HTTPAddr: env("TOURISM_HTTP_ADDR", ":52822"), DatabasePath: env("TOURISM_DATABASE_PATH", "nanjing.db"), SessionTTL: 8 * time.Hour, WorkerInterval: 2 * time.Second, ShutdownTimeout: 10 * time.Second, WorkerBatch: 20}
	var err error
	if cfg.SessionTTL, err = duration("TOURISM_SESSION_TTL", cfg.SessionTTL); err != nil {
		return Config{}, err
	}
	if cfg.WorkerInterval, err = duration("TOURISM_WORKER_INTERVAL", cfg.WorkerInterval); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = duration("TOURISM_SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout); err != nil {
		return Config{}, err
	}
	if raw := os.Getenv("TOURISM_WORKER_BATCH"); raw != "" {
		cfg.WorkerBatch, err = strconv.Atoi(raw)
		if err != nil || cfg.WorkerBatch < 1 {
			return Config{}, fmt.Errorf("invalid TOURISM_WORKER_BATCH")
		}
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}
