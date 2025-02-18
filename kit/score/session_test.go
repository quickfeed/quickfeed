package score

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// To run this test, use this command:
//
//	QUICKFEED_SESSION_SECRET=just-testing go test -v -run TestNewSocket
func TestNewSocket(t *testing.T) {
	go func() {
		if err := NewSocket(context.Background(), sessionSecret); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	env := []string{"QUICKFEED_SESSION_SECRET=" + sessionSecret}
	sh.MustRunEnv(true, env, "testdata/session", "go", "test", "-v", "-run", "TestFibonacci")
}

func TestNewSocketTimeout(t *testing.T) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if err := NewSocket(ctx, sessionSecret); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// try to dial the socket after it should have been closed
	socketPath := filepath.Join(rootSocketDir, sessionSecret+".sock")
	_, err := net.Dial("unix", socketPath)
	wantErr := fmt.Sprintf("dial unix %s: connect: no such file or directory", socketPath)
	if err == nil {
		t.Errorf("socket should not be available, want error: %s", wantErr)
	}
	if err.Error() != wantErr {
		t.Errorf("got error: %v, want error: %s", err, wantErr)
	}
}
