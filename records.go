package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Record struct {
	SourcePath string    `json:"source_path"`
	Hash       string    `json:"hash"`
	CopiedAt   time.Time `json:"copied_at"`
	SizeBytes  int64     `json:"size_bytes"`
}

type RecordStore struct {
	mu        sync.Mutex
	filePath  string
	Records   []Record `json:"records"`
	hashSet   map[string]bool
	pathIndex map[string]Record
}

func NewRecordStore(filePath string) (*RecordStore, error) {
	store := &RecordStore{
		filePath:  filePath,
		hashSet:   make(map[string]bool),
		pathIndex: make(map[string]Record),
	}

	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading records file: %w", err)
	}

	if err := json.Unmarshal(data, store); err != nil {
		return nil, fmt.Errorf("parsing records file: %w", err)
	}

	for _, r := range store.Records {
		store.hashSet[r.Hash] = true
		store.pathIndex[r.SourcePath] = r
	}

	return store, nil
}

func (s *RecordStore) HasHash(hash string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hashSet[hash]
}

func (s *RecordStore) GetByPath(relPath string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.pathIndex[relPath]
	return r, ok
}

func (s *RecordStore) Add(record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Records = append(s.Records, record)
	s.hashSet[record.Hash] = true
	s.pathIndex[record.SourcePath] = record

	return s.flush()
}

func (s *RecordStore) flush() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling records: %w", err)
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("writing temp records file: %w", err)
	}

	if err := os.Rename(tmpPath, s.filePath); err != nil {
		return fmt.Errorf("renaming temp records file: %w", err)
	}

	return nil
}
