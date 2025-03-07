package ui_test

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/internal/ui"
)

func TestBuild(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	if err := ui.Build(false, &tmpDir); err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

// The watch function is exited by the main thread after the test is done.
func TestWatch(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	ch := make(chan error)
	go ui.Watch(ch, false, &tmpDir)
	// Wait for the watch to start
	if err := <-ch; err != nil {
		t.Errorf("Watch failed: %v", err)
	}
}
