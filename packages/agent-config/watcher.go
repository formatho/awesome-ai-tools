package agentconfig

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a configuration file for changes.
// When the file changes, it calls the registered callback.
type Watcher struct {
	mu       sync.Mutex
	watcher  *fsnotify.Watcher
	filePath string
	callback func()

	// debounce timer
	timer    *time.Timer
	debounce time.Duration
}

// WatcherOption is a functional option for configuring the Watcher.
type WatcherOption func(*Watcher)

// WithDebounce sets the debounce duration for file change events.
// This prevents multiple callbacks for a single file save operation.
func WithDebounce(d time.Duration) WatcherOption {
	return func(w *Watcher) {
		w.debounce = d
	}
}

// NewWatcher creates a new file watcher for the specified path.
func NewWatcher(path string, callback func(), opts ...WatcherOption) (*Watcher, error) {
	// Get absolute path for reliable watching
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create fsnotify watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		watcher:  fsWatcher,
		filePath: absPath,
		callback: callback,
		debounce: 100 * time.Millisecond, // default debounce
	}

	// Apply options
	for _, opt := range opts {
		opt(w)
	}

	// Watch the directory (more reliable than watching the file directly)
	dir := filepath.Dir(absPath)
	if err := fsWatcher.Add(dir); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}

	// Start the event loop
	go w.eventLoop()

	return w, nil
}

// eventLoop handles file system events.
func (w *Watcher) eventLoop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Check if this is our file
			if filepath.Base(event.Name) != filepath.Base(w.filePath) {
				continue
			}

			// Check for write or create events
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				w.scheduleCallback()
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err // silence unused variable warning

		}
	}
}

// scheduleCallback schedules the callback with debouncing.
func (w *Watcher) scheduleCallback() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Cancel existing timer
	if w.timer != nil {
		w.timer.Stop()
	}

	// Schedule new callback
	w.timer = time.AfterFunc(w.debounce, func() {
		if w.callback != nil {
			w.callback()
		}
	})
}

// Close stops the watcher and cleans up resources.
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timer != nil {
		w.timer.Stop()
	}

	if w.watcher != nil {
		return w.watcher.Close()
	}

	return nil
}

// FilePath returns the path being watched.
func (w *Watcher) FilePath() string {
	return w.filePath
}

// WatcherStats holds statistics about the watcher.
type WatcherStats struct {
	FilePath   string        `json:"file_path"`
	IsWatching bool          `json:"is_watching"`
	Debounce   time.Duration `json:"debounce"`
}

// Stats returns current watcher statistics.
func (w *Watcher) Stats() WatcherStats {
	w.mu.Lock()
	defer w.mu.Unlock()

	return WatcherStats{
		FilePath:   w.filePath,
		IsWatching: w.watcher != nil,
		Debounce:   w.debounce,
	}
}
