package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

func ScanExisting(cfg Config, copier *Copier) error {
	log.Printf("Scanning %v for existing files...", cfg.WatchDirs)

	workers := runtime.NumCPU()
	paths := make(chan string, workers*2)
	var wg sync.WaitGroup
	var count atomic.Int64

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				copier.Process(path)
			}
		}()
	}

	for _, dir := range cfg.WatchDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("WARN: error accessing %s: %v", path, err)
				return nil
			}

			if info.IsDir() {
				return nil
			}

			count.Add(1)
			paths <- path
			return nil
		})
		if err != nil {
			close(paths)
			return err
		}
	}

	close(paths)
	wg.Wait()

	log.Printf("Initial scan complete. Found %d files.", count.Load())
	return nil
}
