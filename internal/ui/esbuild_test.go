package ui_test

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/internal/ui"
)

func TestBuild(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping %s when running on GitHub", t.Name())
	}
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)
	if err := ui.Build(tmpDir, true); err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

// The watch go routine is exited by the main thread after the test is done.
func TestWatch(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping %s when running on GitHub", t.Name())
	}
	if err := ui.Watch(); err != nil {
		t.Errorf("Watch failed: %v", err)
	}
}
