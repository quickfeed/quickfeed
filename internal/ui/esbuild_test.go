package ui_test

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/internal/ui"
)

func TestBuild(t *testing.T) {
	if os.Getenv("ESBUILD_TESTS") == "" {
		t.SkipNow()
	}

	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	if err := ui.Build(&tmpDir); err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

// The watch function is exited by the main thread after the test is done.
func TestWatch(t *testing.T) {
	if os.Getenv("ESBUILD_TESTS") == "" {
		t.SkipNow()
	}
	ch := make(chan error)
	go ui.Watch(ch)
	// Only wait for the first error, which is expected to be nil
	if err := <-ch; err != nil {
		t.Errorf("Watch failed: %v", err)
	}
}
