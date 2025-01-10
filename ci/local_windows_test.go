//go:build windows

package ci_test

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
)

func TestLocalWindows(t *testing.T) {
	const (
		cmd      = `printf "Hello World"`
		expected = "Hello World"
	)

	out := runCmd(t, []string{cmd})

	if expected != out {
		t.Errorf("Have %#v want %#v", out, expected)
	}
}

func runCmd(t *testing.T, cmds []string) string {
	local := ci.Local{}
	out, err := local.Run(context.Background(), &ci.Job{
		Commands: cmds,
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func checkOwner(t *testing.T, _ string) {
	t.Helper()
	t.Log("Skipping checkOwner on Windows")
}
