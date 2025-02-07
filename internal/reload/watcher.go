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
	clients map[chan string]bool
	mu      sync.Mutex
}

// NewWatcher creates a new watcher for the given path.
// The watcher listens for file changes and broadcasts them to all connected clients.
// Note: Usually you only ever have one client as this is intended for live-reloading
// a web page in a development environment.
func NewWatcher(ctx context.Context, path string) (*Watcher, error) {
	watcher := &Watcher{
		clients: make(map[chan string]bool),
	}
	watchFunc, err := watcher.start(path)
	if err != nil {
		return nil, err
	}
	go watchFunc(ctx) // Start watching for file changes
	go webpack()      // Start webpack in watch mode
	return watcher, nil
}

func (w *Watcher) start(path string) (func(ctx context.Context), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(path)
	if err != nil {
		return nil, err
	}

	// Start listening for events.
	return func(ctx context.Context) {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return // Watcher closed
				}
				// We only care about writes and creates.
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					// Broadcast event to all clients.
					w.broadcastMessage(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error watching files:", err)
			case <-ctx.Done():
				watcher.Close()
				return
			}
		}
	}, nil
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
	c := exec.Command("webpack", "--mode=development", "--watch")
	c.Dir = "public"
	if err := c.Run(); err != nil {
		log.Print(c.Output())
		log.Print(err)
	}
}
