package sh_test

import (
	"fmt"
	"testing"

	"github.com/autograde/quickfeed/kit/sh"
)

func TestRun(t *testing.T) {
	err := sh.Run("cat doesnotexist.txt")
	if err == nil {
		t.Error(err)
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
	s, err = sh.Output("golangci-lint run --tests=true --disable structcheck --disable unused --disable deadcode --disable varcheck")
	if err != nil {
		t.Error(err)
	}
	if s != "" {
		fmt.Println(s)
	}
}
