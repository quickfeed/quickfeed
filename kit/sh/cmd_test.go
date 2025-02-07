package sh_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/kit/sh"
)

func TestRun(t *testing.T) {
	if err := sh.Run("cat doesnotexist.txt"); err == nil {
		t.Error("expected: exit status 1")
	}
}

func TestOutput(t *testing.T) {
	s, err := sh.Output("ls -la")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(s)
}

func TestLintAG(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		// since we don't have golangci-lint installed in the GitHub Actions runner
		t.Skip("Skipping test since it is running in GitHub Actions")
	}

	// check formatting using goimports
	s, err := sh.Output("golangci-lint run --tests=false --disable-all --enable goimports")
	if err != nil {
		t.Error(err)
	}
	if s != "" {
		t.Error("Formatting check failed: goimports")
	}

	// check for TODO/BUG/FIXME comments using godox
	s, err = sh.Output("golangci-lint run --tests=false --disable-all --enable godox")
	if err == nil {
		t.Error("Expected golangci-lint to return error with 'exit status 1'")
	}
	if s == "" {
		t.Error("Expected golangci-lint to return message 'TODO(meling) test golangci-lint godox check for TODO items'")
	}

	// check many things: probably too aggressive for DAT320
	s, err = sh.Output("golangci-lint run --tests=true --disable unused")
	if err != nil {
		t.Error(err)
	}
	if s != "" {
		fmt.Println(s)
	}
}
