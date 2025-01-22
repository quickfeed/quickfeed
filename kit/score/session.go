package score

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// rootSocketDir is the root directory for the Unix domain sockets.
const rootSocketDir = "/tmp/quickfeed-sessions"

// session encapsulates the state of a session.
type session struct {
	wg                sync.WaitGroup
	mu                sync.Mutex
	activeConnections int
	socketPath        string
	listener          net.Listener
}

// newSession creates a new session for the given socketPath.
func newSession(sessionSecret string) (*session, error) {
	socketPath := filepath.Join(rootSocketDir, sessionSecret+".sock")
	// remove an existing socket file, if it exists (shouldn't happen)
	_ = os.Remove(socketPath)

	// create listener for the session-specific Unix domain socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return &session{
		socketPath: socketPath,
		listener:   listener,
	}, nil
}

// incActiveConnections increments the active connection count.
func (s *session) incActiveConnections() {
	s.wg.Add(1)
	s.mu.Lock()
	s.activeConnections++
	s.mu.Unlock()
}

// decActiveConnections decrements the active connection count.
// If no connections remain, it cleans up the socket file.
func (s *session) decActiveConnections() {
	s.mu.Lock()
	s.activeConnections--
	if s.activeConnections == 0 {
		// fmt.Println("No active connections; cleaning up socket and closing listener.")
		// close the listener to stop accepting new connections and remove the socket file
		_ = s.listener.Close()
		_ = os.Remove(s.socketPath)
	}
	s.mu.Unlock()
	s.wg.Done()
}

// NewSocket creates a new Unix domain socket for the session.
// This will block until the listener is closed by the last client,
// or the context is canceled.
func NewSocket(ctx context.Context, sessionSecret string) error {
	// ensure the root socket directory exists
	if err := os.MkdirAll(rootSocketDir, 0o700); err != nil {
		return fmt.Errorf("failed to create root socket directory: %w", err)
	}

	// create Unix domain socket for the session
	sess, err := newSession(sessionSecret)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	go func() {
		<-ctx.Done()
		// fmt.Println("Context canceled; shutting down listener.")
		// Close the listener to break the accept loop
		_ = sess.listener.Close()
	}()

	for {
		conn, err := sess.listener.Accept()
		if err != nil {
			// fmt.Printf("Error accepting connection on socket %s: %v\n", sess.socketPath, err)
			// if the listener is closed, stop accepting new connections
			return nil
		}

		sess.incActiveConnections()

		go func(conn net.Conn) {
			defer sess.decActiveConnections()
			defer func() { _ = conn.Close() }() // ignore any errors on close

			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				line := scanner.Text()
				score, err := parse(line, sessionSecret)
				if err != nil {
					// fmt.Printf("Connection closed: invalid message: %s (error: %v)\n", line, err)
					// invalid message; closing connection
					return
				}
				fmt.Printf("Valid score received: %v\n", score)
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("Error reading from connection: %v\n", err)
			}
		}(conn)
	}
}
