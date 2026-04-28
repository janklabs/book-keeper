package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	cfg     Config
	copier  *Copier
	fsw     *fsnotify.Watcher
	done    chan struct{}
	pending map[string]*time.Timer
	mu      sync.Mutex
}

func NewWatcher(cfg Config, copier *Copier) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		cfg:     cfg,
		copier:  copier,
		fsw:     fsw,
		done:    make(chan struct{}),
		pending: make(map[string]*time.Timer),
	}

	for _, dir := range cfg.WatchDirs {
		if err := w.addRecursive(dir); err != nil {
			fsw.Close()
			return nil, fmt.Errorf("watching %s: %w", dir, err)
		}
	}

	return w, nil
}

func (w *Watcher) Start() {
	log.Printf("Watching %v for new files...", w.cfg.WatchDirs)

	go func() {
		for {
			select {
			case event, ok := <-w.fsw.Events:
				if !ok {
					return
				}
				w.handleEvent(event)
			case err, ok := <-w.fsw.Errors:
				if !ok {
					return
				}
				log.Printf("WARN: watcher error: %v", err)
			}
		}
	}()
}

func (w *Watcher) Stop() {
	w.fsw.Close()

	w.mu.Lock()
	for _, timer := range w.pending {
		timer.Stop()
	}
	w.mu.Unlock()
}

const debounceDelay = 500 * time.Millisecond

func (w *Watcher) handleEvent(event fsnotify.Event) {
	isCreate := event.Op&fsnotify.Create != 0
	isWrite := event.Op&fsnotify.Write != 0

	if !isCreate && !isWrite {
		return
	}

	info, err := os.Stat(event.Name)
	if err != nil {
		return
	}

	if info.IsDir() && isCreate {
		if err := w.addRecursive(event.Name); err != nil {
			log.Printf("WARN: failed to watch new directory %s: %v", event.Name, err)
		}
		return
	}

	if info.IsDir() {
		return
	}

	w.debounce(event.Name)
}

func (w *Watcher) debounce(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if timer, exists := w.pending[path]; exists {
		timer.Reset(debounceDelay)
		return
	}

	w.pending[path] = time.AfterFunc(debounceDelay, func() {
		w.mu.Lock()
		delete(w.pending, path)
		w.mu.Unlock()

		w.copier.Process(path)
	})
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if err := w.fsw.Add(path); err != nil {
				log.Printf("WARN: could not watch %s: %v", path, err)
			}
		} else {
			w.debounce(path)
		}
		return nil
	})
}
