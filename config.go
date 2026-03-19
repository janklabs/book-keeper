package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	WatchDir           string
	IngestionDir       string
	RecordsFile        string
	StabilizationDelay time.Duration
	ScanOnStartup      bool
	PollInterval       time.Duration
}

func LoadConfig() (Config, error) {
	cfg := Config{
		WatchDir:           os.Getenv("WATCH_DIR"),
		IngestionDir:       os.Getenv("INGESTION_DIR"),
		RecordsFile:        envOrDefault("RECORDS_FILE", "/data/records.json"),
		StabilizationDelay: time.Duration(envIntOrDefault("STABILIZATION_SECS", 5)) * time.Second,
		ScanOnStartup:      envBoolOrDefault("SCAN_ON_STARTUP", true),
		PollInterval:       time.Duration(envIntOrDefault("POLL_INTERVAL_SECS", 0)) * time.Second,
	}

	if cfg.WatchDir == "" {
		return cfg, fmt.Errorf("WATCH_DIR environment variable is required")
	}
	if cfg.IngestionDir == "" {
		return cfg, fmt.Errorf("INGESTION_DIR environment variable is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envBoolOrDefault(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
