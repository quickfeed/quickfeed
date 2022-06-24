package hooks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractChanges(t *testing.T) {
	modifiedFiles := []string{
		"go.mod",
		"go.sum",
		"exercise.go",
		"README.md",
		"lab2/fib.go",
		"lab3/detector/fd.go",
		"paxos/proposer.go",
		"/hallo",
		"",
	}
	want := map[string]bool{
		"lab2":  true,
		"lab3":  true,
		"paxos": true,
	}
	got := make(map[string]bool)
	extractChanges(modifiedFiles, got)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("content mismatch (-want +got):\n%s", diff)
	}
}
