package reload

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/rand"
)

func TestWatcher(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, rand.String())

	// create a file to modify later
	// to trigger WRITE events in the watcher
	file, err := os.Create(filename)
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	watcher, err := NewWatcher(ctx, dir)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	// create a server with a handler the client can connect to
	mux := http.NewServeMux()
	mux.HandleFunc("/watch", watcher.Handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	go func() {
		// wait some time before writing to the file
		time.Sleep(2 * time.Second)
		// try to write to the file multiple times
		// in case the server is slow to start
		for range 3 {
			// write to the file to trigger the watcher
			// to send an event to the client
			_, err = file.WriteString(rand.String())
			if err != nil {
				t.Errorf("failed to write to file: %v", err)
			}
		}
	}()

	// connect to the server
	resp, err := server.Client().Get(server.URL + "/watch")
	if err != nil {
		t.Fatalf("failed to get from server: %v", err)
	}
	defer resp.Body.Close()

	// continuously read from the response body
	eventChan := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			// send the message to the event channel
			eventChan <- line
			break
		}
	}()

	// we want the received message to contain the filename
	want := fmt.Sprintf("data: %s\n", filename)
	select {
	case msg := <-eventChan:
		if msg != want {
			t.Errorf("Expected event message: %s, got: %s", want, msg)
		}
		return
	case <-time.After(10 * time.Second):
		// if we don't receive an event in 10 seconds, fail the test
		t.Error("Timeout: No event received")
	}
}
