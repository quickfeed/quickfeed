package hooks

import (
	"testing"

	"github.com/google/go-github/v45/github"
)

func TestBranchName(t *testing.T) {
	tests := []struct {
		ref        string
		wantBranch string
	}{
		{
			"refs/heads/main",
			"main",
		},
		{
			"refs/heads/master",
			"master",
		},
		{
			"/refs/main",
			"main",
		},
	}

	for _, tt := range tests {
		gotBranch := branchName(tt.ref)
		if gotBranch != tt.wantBranch {
			t.Errorf("expected branch name %s, got %s", tt.wantBranch, gotBranch)
		}
	}
}

func TestDefaultBranch(t *testing.T) {
	tests := []struct {
		ref         string
		repoDefault string
		want        bool
	}{
		{
			"refs/heads/main",
			"main",
			true,
		},
		{
			"refs/heads/master",
			"master",
			true,
		},
		{
			"refs/heads/main",
			"master",
			false,
		},
		{
			"refs/heads/master",
			"main",
			false,
		},
	}

	for _, tt := range tests {
		payload := &github.PushEvent{
			Ref: &tt.ref,
			Repo: &github.PushEventRepository{
				DefaultBranch: &tt.repoDefault,
			},
		}
		if isDefaultBranch(payload) != tt.want {
			t.Errorf("default branch: '%s', ref branch: '%s', expected to match: '%v'",
				tt.repoDefault, tt.ref, tt.want)
		}
	}
}

// func TestGitHubWebHook_updateLastActivityDate(t *testing.T) {

// 	tests := []struct {
// 		name string
// 		repo *qf.Repository
// 	}{}
// }
