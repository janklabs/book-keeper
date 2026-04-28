package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Copier struct {
	cfg   Config
	store *RecordStore
}

func NewCopier(cfg Config, store *RecordStore) *Copier {
	return &Copier{cfg: cfg, store: store}
}

func (c *Copier) Process(absolutePath string) {
	relPath, err := c.resolveRelPath(absolutePath)
	if err != nil {
		log.Printf("ERROR: resolving relative path for %s: %v", absolutePath, err)
		return
	}

	// Fast path: skip hashing if path and size match an existing record.
	if info, err := os.Stat(absolutePath); err == nil {
		if rec, ok := c.store.GetByPath(relPath); ok && rec.SizeBytes == info.Size() {
			log.Printf("SKIP: %s (unchanged, size %d)", relPath, info.Size())
			return
		}
	}

	if err := c.waitUntilStable(absolutePath); err != nil {
		log.Printf("ERROR: waiting for file to stabilize %s: %v", relPath, err)
		return
	}

	hash, err := HashFile(absolutePath)
	if err != nil {
		log.Printf("ERROR: hashing %s: %v", relPath, err)
		return
	}

	if c.store.HasHash(hash) {
		log.Printf("SKIP: %s (duplicate content, hash %s)", relPath, hash[:20]+"...")
		return
	}

	destPath := filepath.Join(c.cfg.IngestionDir, relPath)

	if info, err := os.Stat(destPath); err == nil && !info.IsDir() {
		log.Printf("WARN: destination %s already exists, will be overwritten by %s", relPath, absolutePath)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		log.Printf("ERROR: creating destination directory for %s: %v", relPath, err)
		return
	}

	info, err := copyFile(absolutePath, destPath)
	if err != nil {
		log.Printf("ERROR: copying %s: %v", relPath, err)
		return
	}

	record := Record{
		SourcePath: relPath,
		Hash:       hash,
		CopiedAt:   time.Now().UTC(),
		SizeBytes:  info.Size(),
	}

	if err := c.store.Add(record); err != nil {
		log.Printf("ERROR: recording %s: %v", relPath, err)
		return
	}

	log.Printf("COPIED: %s (%d bytes)", relPath, info.Size())
}

func (c *Copier) waitUntilStable(path string) error {
	if c.cfg.StabilizationDelay <= 0 {
		return nil
	}

	checkInterval := 1 * time.Second
	stableFor := time.Duration(0)
	lastSize := int64(-1)

	for stableFor < c.cfg.StabilizationDelay {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat failed: %w", err)
		}

		currentSize := info.Size()
		if currentSize == lastSize {
			stableFor += checkInterval
		} else {
			stableFor = 0
			lastSize = currentSize
		}

		if stableFor < c.cfg.StabilizationDelay {
			time.Sleep(checkInterval)
		}
	}

	return nil
}

func copyFile(src, dst string) (os.FileInfo, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("opening source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("creating destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return nil, fmt.Errorf("copying data: %w", err)
	}

	if err := dstFile.Sync(); err != nil {
		return nil, fmt.Errorf("syncing destination: %w", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		return nil, fmt.Errorf("stat destination: %w", err)
	}

	return info, nil
}

// resolveRelPath finds which watch directory contains the file and returns the relative path.
func (c *Copier) resolveRelPath(absolutePath string) (string, error) {
	for _, dir := range c.cfg.WatchDirs {
		rel, err := filepath.Rel(dir, absolutePath)
		if err != nil {
			continue
		}
		if len(rel) >= 2 && rel[:2] == ".." {
			continue
		}
		return rel, nil
	}
	return "", fmt.Errorf("path %s is not under any watch directory", absolutePath)
}
