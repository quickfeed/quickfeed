package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// pingStub is a Docker runner stand-in whose Ping returns err.
type pingStub struct{ err error }

func (p pingStub) Ping(context.Context) error { return p.err }

func TestCheckDocker(t *testing.T) {
	if err := checkDocker(t.Context(), pingStub{}); err != nil {
		t.Errorf("checkDocker() with a running daemon = %v, want nil", err)
	}

	daemonDown := errors.New("dial unix /var/run/docker.sock: connect: no such file or directory")
	err := checkDocker(t.Context(), pingStub{err: daemonDown})
	if err == nil {
		t.Fatal("checkDocker() with a stopped daemon = nil, want an error")
	}
	for _, want := range []string{"Docker daemon is not available", daemonDown.Error(), "Start Docker"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("checkDocker() error %q does not mention %q", err, want)
		}
	}
	if !errors.Is(err, daemonDown) {
		t.Errorf("checkDocker() error does not wrap the daemon error: %v", err)
	}
}
