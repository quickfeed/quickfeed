package reload

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	clients   map[chan string]bool
	mu        sync.Mutex
}

// NewWatcher creates a new watcher for the given path.
// The watcher listens for file changes and broadcasts them to all connected clients.
// While a single client is the most common use case, multiple clients can connect to
// the same watcher, e.g., for live-reloading the web page in different browsers.
func NewWatcher(ctx context.Context, path string) (*Watcher, error) {
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
	}
	go watcher.start(ctx) // Start watching for file changes
	go webpack()          // Start webpack in watch mode
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

func webpack() {
	log.Println("Running webpack...")
	c := exec.Command("npx", "webpack", "--mode=development", "--watch")
	c.Dir = "public"
	if err := c.Run(); err != nil {
		log.Print(c.Output())
		log.Print(err)
	}
}
