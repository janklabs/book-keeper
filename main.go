package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var version = "dev"

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	store, err := NewRecordStore(cfg.RecordsFile)
	if err != nil {
		log.Fatalf("Failed to load records: %v", err)
	}

	copier := NewCopier(cfg, store)

	if cfg.ScanOnStartup {
		if err := ScanExisting(cfg, copier); err != nil {
			log.Fatalf("Initial scan failed: %v", err)
		}
	}

	watcher, err := NewWatcher(cfg, copier)
	if err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}
	watcher.Start()

	log.Printf("book-keeper %s is running. WATCH_DIRS=[%s] INGESTION_DIR=%s", version, strings.Join(cfg.WatchDirs, ", "), cfg.IngestionDir)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	watcher.Stop()
	log.Println("Stopped.")
}
