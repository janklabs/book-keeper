package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	WatchDirs          []string
	IngestionDir       string
	RecordsFile        string
	StabilizationDelay time.Duration
	ScanOnStartup      bool
	PollInterval       time.Duration
}

func LoadConfig() (Config, error) {
	cfg := Config{
		WatchDirs:          parseWatchDirs(),
		IngestionDir:       envOrDefault("INGESTION_DIR", "/ingestion"),
		RecordsFile:        envOrDefault("RECORDS_FILE", "/data/records.json"),
		StabilizationDelay: time.Duration(envIntOrDefault("STABILIZATION_SECS", 5)) * time.Second,
		ScanOnStartup:      envBoolOrDefault("SCAN_ON_STARTUP", true),
		PollInterval:       time.Duration(envIntOrDefault("POLL_INTERVAL_SECS", 0)) * time.Second,
	}

	if len(cfg.WatchDirs) == 0 {
		return cfg, fmt.Errorf("no watch directories configured")
	}

	return cfg, nil
}

// parseWatchDirs reads WATCH_DIRS (newline-separated) with fallback to WATCH_DIR.
func parseWatchDirs() []string {
	if v := os.Getenv("WATCH_DIRS"); v != "" {
		var dirs []string
		for _, line := range strings.Split(v, "\n") {
			d := strings.TrimSpace(line)
			if d != "" {
				dirs = append(dirs, d)
			}
		}
		return dirs
	}
	return []string{envOrDefault("WATCH_DIR", "/watch")}
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
