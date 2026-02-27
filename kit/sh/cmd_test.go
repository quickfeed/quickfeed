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

func TestCodeQuality(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		// since we don't have golangci-lint installed in the GitHub Actions runner
		t.Skip("Skipping test since it is running in GitHub Actions")
	}

	tests := []struct {
		name     string
		cmd      string
		msg      string
		wantFail bool
	}{
		{
			name:     "todo",
			cmd:      "golangci-lint run --tests=false --show-stats=false --enable-only godox",
			msg:      "TODO checker failed: TODO comments must be removed when a task has been completed",
			wantFail: true,
		},
		{
			name: "lint",
			cmd:  "golangci-lint run --show-stats=false --disable errcheck",
			msg:  "Lint checker failed",
		},
		{
			name: "vet",
			cmd:  "go vet ./...",
			msg:  "Vet checker failed",
		},
		{
			name: "gofmt",
			cmd:  "golangci-lint fmt --diff-colored --enable gofmt",
			msg:  "Formatting checker failed: please configure your editor to format code on save",
		},
		{
			name: "goimports",
			cmd:  "golangci-lint fmt --diff-colored --enable goimports",
			msg:  "Formatting checker failed: please configure your editor to format code on save",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := sh.Output(tt.cmd)
			if tt.wantFail {
				if err == nil {
					t.Error("Expected golangci-lint to return error with 'exit status 1'")
				}
				if s == "" {
					t.Error("Expected golangci-lint to return message 'TODO(meling) test golangci-lint godox check for TODO items'")
				}
				return
			}
			if err != nil {
				t.Error(err)
			}
			if s != "" {
				t.Error(tt.msg)
				fmt.Println(s)
			}
		})
	}
}
