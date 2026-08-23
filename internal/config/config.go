package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL    string
	HTTPAddr       string
	SessionTTL     time.Duration
	WorkerInterval time.Duration
}

func Load() Config {
	ttl := 12 * time.Hour
	if raw := os.Getenv("SESSION_TTL"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			ttl = parsed
		}
	}
	interval := 2 * time.Second
	if raw := os.Getenv("WORKER_INTERVAL"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			interval = parsed
		}
	}
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		db = "./data/autumn.db"
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{DatabaseURL: db, HTTPAddr: addr, SessionTTL: ttl, WorkerInterval: interval}
}
