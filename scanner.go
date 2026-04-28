package main

import (
	"log"
	"os"
	"path/filepath"
)

func ScanExisting(cfg Config, copier *Copier) error {
	log.Printf("Scanning %v for existing files...", cfg.WatchDirs)

	count := 0
	for _, dir := range cfg.WatchDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("WARN: error accessing %s: %v", path, err)
				return nil
			}

			if info.IsDir() {
				return nil
			}

			count++
			copier.Process(path)
			return nil
		})
		if err != nil {
			return err
		}
	}

	log.Printf("Initial scan complete. Found %d files.", count)
	return nil
}
