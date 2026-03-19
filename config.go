package main

import (
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
		WatchDir:           envOrDefault("WATCH_DIR", "/watch"),
		IngestionDir:       envOrDefault("INGESTION_DIR", "/ingestion"),
		RecordsFile:        envOrDefault("RECORDS_FILE", "/data/records.json"),
		StabilizationDelay: time.Duration(envIntOrDefault("STABILIZATION_SECS", 5)) * time.Second,
		ScanOnStartup:      envBoolOrDefault("SCAN_ON_STARTUP", true),
		PollInterval:       time.Duration(envIntOrDefault("POLL_INTERVAL_SECS", 0)) * time.Second,
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
