package watcher

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var watchedExts = map[string]bool{
	".go":   true,
	".sh":   true,
	".bash": true,
	".toml": true,
	".yaml": true,
	".yml":  true,
	".json": true,
	".env":  true,
}

// Watcher monitors a directory for file changes, debounces rapid events,
// and emits file paths for supported extensions.
type Watcher struct {
	fsw      *fsnotify.Watcher
	events   chan string
	debounce time.Duration

	mu     sync.Mutex
	timers map[string]*time.Timer
	done   chan struct{}
}

// New creates a Watcher monitoring root and its subtree.
// Events are debounced by debounceDur per file path.
func New(root string, debounceDur time.Duration) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := fsw.Add(root); err != nil {
		fsw.Close()
		return nil, err
	}

	w := &Watcher{
		fsw:      fsw,
		events:   make(chan string, 64),
		debounce: debounceDur,
		timers:   make(map[string]*time.Timer),
		done:     make(chan struct{}),
	}

	go w.loop()
	return w, nil
}

// Events returns the channel that receives debounced file paths.
func (w *Watcher) Events() <-chan string {
	return w.events
}

// Close stops the watcher and releases resources.
func (w *Watcher) Close() error {
	close(w.done)
	err := w.fsw.Close()
	w.mu.Lock()
	for _, t := range w.timers {
		t.Stop()
	}
	w.mu.Unlock()
	return err
}

func (w *Watcher) loop() {
	for {
		select {
		case <-w.done:
			return
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Create|fsnotify.Write) == 0 {
				continue
			}
			ext := filepath.Ext(ev.Name)
			if !watchedExts[ext] {
				continue
			}
			w.scheduleSend(ev.Name)
		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) scheduleSend(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if t, ok := w.timers[path]; ok {
		t.Reset(w.debounce)
		return
	}

	w.timers[path] = time.AfterFunc(w.debounce, func() {
		w.mu.Lock()
		delete(w.timers, path)
		w.mu.Unlock()

		select {
		case w.events <- path:
		case <-w.done:
		}
	})
}
