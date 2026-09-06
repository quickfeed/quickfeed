package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestRunRejectsInvalidFlags checks the flag combinations that must be rejected
// before anything is generated or the trust store is touched. run reports these
// as errUsage and returns before loading the environment, so no valid
// combination is exercised here.
func TestRunRejectsInvalidFlags(t *testing.T) {
	tests := []struct {
		name                 string
		gen, add, rem, force bool
	}{
		{name: "no action"},
		{name: "gen and add", gen: true, add: true},
		{name: "gen and remove", gen: true, rem: true},
		{name: "add and remove", add: true, rem: true},
		{name: "all actions", gen: true, add: true, rem: true},
		{name: "force without action", force: true},
		{name: "force with add", add: true, force: true},
		{name: "force with remove", rem: true, force: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.gen, tt.add, tt.rem, tt.force)
			if !errors.Is(err, errUsage) {
				t.Errorf("run(%t, %t, %t, %t) = %v, want %v", tt.gen, tt.add, tt.rem, tt.force, err, errUsage)
			}
		})
	}
}

func TestActions(t *testing.T) {
	tests := []struct {
		flags []bool
		want  int
	}{
		{flags: []bool{false, false, false}, want: 0},
		{flags: []bool{true, false, false}, want: 1},
		{flags: []bool{false, true, false}, want: 1},
		{flags: []bool{true, true, false}, want: 2},
		{flags: []bool{true, true, true}, want: 3},
	}
	for _, tt := range tests {
		if got := actions(tt.flags...); got != tt.want {
			t.Errorf("actions(%v) = %d, want %d", tt.flags, got, tt.want)
		}
	}
}

func TestExistingFiles(t *testing.T) {
	dir := t.TempDir()
	present := filepath.Join(dir, "fullchain.pem")
	if err := os.WriteFile(present, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "privkey.pem")

	if got := existingFiles(missing); got != nil {
		t.Errorf("existingFiles(%q) = %v, want nil", missing, got)
	}
	want := []string{"fullchain.pem"}
	if got := existingFiles(present, missing); !slices.Equal(got, want) {
		t.Errorf("existingFiles(%q, %q) = %v, want %v", present, missing, got, want)
	}
}
