package score

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"
)

// NewSocketRegistry creates a new score registry that connects to the session socket
// for the current session. If the connection fails, it falls back to standard stdout
// score reporting.
func NewSocketRegistry() *registry { // skipcq: RVV-B0011
	socketPath := filepath.Join(rootSocketDir, sessionSecret+".sock")
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Printf("failed to connect to session socket: %v\n", err)
		// if we failed to connect, we fall back to standard stdout score reporting
	}
	return &registry{
		testNames: make([]string, 0),
		scores:    make(map[string]*Score),
		conn:      conn,
	}
}

// PrintToSocket prints the score object to the session socket.
func (s *registry) PrintToSocket(t *testing.T, sc *Score) {
	// print JSON score object: {"Secret":"my secret code","TestName": ...}
	fmt.Fprintln(s.conn, sc.JSON())
	s.conn.Close()
}
