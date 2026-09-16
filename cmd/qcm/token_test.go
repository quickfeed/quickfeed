package main

import (
	"os/exec"
	"strings"
	"testing"
)

// ghStub returns a GitHub CLI runner that prints out and fails with err.
func ghStub(out string, err error) ghRunner {
	return func(...string) ([]byte, error) {
		return []byte(out), err
	}
}

// ghNotInstalled mimics exec.Command failing to find the gh binary.
var ghNotInstalled = &exec.Error{Name: "gh", Err: exec.ErrNotFound}

// ghNotSignedIn mimics 'gh auth token' exiting non-zero with its explanation
// on standard error.
var ghNotSignedIn = &exec.ExitError{Stderr: []byte("no oauth token found for github.com\n")}

func TestResolveToken(t *testing.T) {
	tests := []struct {
		name      string
		explicit  string
		gh        ghRunner
		wantToken string
		wantErr   []string // substrings the error must contain
	}{
		{
			name:     "ExplicitTokenWins",
			explicit: "ghp_explicit",
			gh: func(...string) ([]byte, error) {
				t.Fatal("the GitHub CLI must not be asked when a token is given")
				return nil, nil
			},
			wantToken: "ghp_explicit",
		},
		{
			name:      "GitHubCLIToken",
			gh:        ghStub("gho_fromcli\n", nil),
			wantToken: "gho_fromcli",
		},
		{
			name:    "GitHubCLINotInstalled",
			gh:      ghStub("", ghNotInstalled),
			wantErr: []string{"not installed", "gh auth login", tokenEnv, "settings/tokens"},
		},
		{
			name:    "GitHubCLINotSignedIn",
			gh:      ghStub("", ghNotSignedIn),
			wantErr: []string{"no oauth token found", "gh auth login", tokenEnv},
		},
		{
			name:    "GitHubCLIPrintsNothing",
			gh:      ghStub("\n", nil),
			wantErr: []string{"printed no token", "gh auth login"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := resolveToken(tc.explicit, tc.gh)
			if len(tc.wantErr) == 0 {
				if err != nil {
					t.Fatalf("resolveToken() error = %v, want nil", err)
				}
				if token != tc.wantToken {
					t.Errorf("resolveToken() = %q, want %q", token, tc.wantToken)
				}
				return
			}
			if err == nil {
				t.Fatalf("resolveToken() = %q, want an error", token)
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("resolveToken() error %q does not mention %q", err, want)
				}
			}
		})
	}
}

// TestCloneWithoutToken checks that the clone command reports how to
// authenticate when no token is given and the GitHub CLI is not signed in.
func TestCloneWithoutToken(t *testing.T) {
	t.Setenv(tokenEnv, "")
	restore := gh
	gh = ghStub("", ghNotSignedIn)
	t.Cleanup(func() { gh = restore })

	err := run([]string{"clone", "-course", "qf101-2026", "-dir", t.TempDir()}, &strings.Builder{}, &strings.Builder{})
	if err == nil {
		t.Fatal("run(clone) error = nil, want a missing token error")
	}
	for _, want := range []string{"no GitHub access token found", "gh auth login", tokenEnv} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("run(clone) error %q does not mention %q", err, want)
		}
	}
}
