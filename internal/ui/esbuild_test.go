package ui_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/ui"
)

func TestBuild(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping %s when running on GitHub", t.Name())
	}
	if err := ui.Build(t.TempDir(), true); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	html, err := os.ReadFile(env.Root("public", "assets", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	versionedTailwind := regexp.MustCompile(`href="/static/tailwind\.css\?v=[0-9a-f]{8}"`)
	if !versionedTailwind.Match(html) {
		t.Errorf("index.html does not link a versioned Tailwind stylesheet:\n%s", html)
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
