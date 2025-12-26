package reload

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	clients   map[chan string]bool
	mu        sync.Mutex
	// List of files to watch for changes.
	// Only changes to these files will be broadcast to clients.
	watchList []string
}

// NewWatcher creates a new watcher for the given path.
// The watcher listens for file changes for the specified watchlist and broadcasts them to all connected clients.
// While a single client is the most common use case, multiple clients can connect to
// the same watcher, e.g., for live-reloading the web page in different browsers.
func NewWatcher(ctx context.Context, path string, watchList ...string) (*Watcher, error) {
	if len(watchList) == 0 {
		return nil, fmt.Errorf("nothing to watch")
	}
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err = fsWatcher.Add(path); err != nil {
		return nil, err
	}
	watcher := &Watcher{
		fsWatcher: fsWatcher,
		clients:   make(map[chan string]bool),
		watchList: watchList,
	}
	go watcher.start(ctx) // Start watching for file changes
	return watcher, nil
}

// start listening for events and broadcast them to clients.
// The watcher will stop when the context is canceled.
func (w *Watcher) start(ctx context.Context) {
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return // Watcher closed
			}
			// We only care about writes and creates.
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				// Compare only the base filename since event.Name is a path.
				if !slices.Contains(w.watchList, filepath.Base(event.Name)) {
					continue // Ignore non-watched files and keep watching
				}
				// Broadcast event to all clients.
				w.broadcastMessage(event.Name)
			}
		case err := <-w.fsWatcher.Errors:
			log.Println("error watching files:", err)
		case <-ctx.Done():
			w.fsWatcher.Close()
			return
		}
	}
}

// Add a new client
func (w *Watcher) addClient(ch chan string) {
	w.mu.Lock()
	w.clients[ch] = true
	w.mu.Unlock()
}

// Remove a disconnected client
func (w *Watcher) removeClient(ch chan string) {
	w.mu.Lock()
	delete(w.clients, ch)
	close(ch)
	w.mu.Unlock()
}

// Broadcast a message to all clients
func (w *Watcher) broadcastMessage(msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for ch := range w.clients {
		select {
		case ch <- msg:
			continue
		default:
			// do nothing
		}
	}
}

func (watcher *Watcher) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a new channel for this client
	client := make(chan string, 1)
	watcher.addClient(client)
	// Listen for changes and send updates
	for {
		select {
		case msg := <-client:
			// message must end with \n\n to mark the end of the event
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		case <-r.Context().Done(): // Client disconnected
			watcher.removeClient(client)
			return
		}
	}
}
