package ci_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
)

func TestNewSkeletonRun(t *testing.T) {
	root := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", root)
	course := &qf.Course{Code: "QF101", ScmOrganizationName: "qf101-2025"}
	assignment := &qf.Assignment{Name: "lab1"}

	runData := ci.NewSkeletonRun(course, assignment, "abc1234")
	wantDir := filepath.Join(root, "qf101-2025", qf.AssignmentsRepo)
	if runData.SubmissionDir != wantDir {
		t.Errorf("NewSkeletonRun() submission dir = %q, want %q", runData.SubmissionDir, wantDir)
	}
	if got, want := runData.Repo.Name(), "skeleton-labs"; got != want {
		t.Errorf("NewSkeletonRun() repo name = %q, want %q", got, want)
	}
	// The repository name must not collide with the read-only mounts.
	for _, reserved := range []string{qf.TestsRepo, qf.AssignmentsRepo} {
		if runData.Repo.Name() == reserved {
			t.Errorf("NewSkeletonRun() repo name = %q, which collides with the %q mount", reserved, reserved)
		}
	}
	if runData.JobOwner != "skeleton" {
		t.Errorf("NewSkeletonRun() job owner = %q, want %q", runData.JobOwner, "skeleton")
	}
	if !strings.Contains(runData.String(), "abc123") {
		t.Errorf("NewSkeletonRun() job name = %q, want it to contain the commit ID", runData.String())
	}
}

func TestSkeletonProblem(t *testing.T) {
	scores := func(values ...int32) []*score.Score {
		out := make([]*score.Score, len(values))
		for i, v := range values {
			out[i] = &score.Score{TestName: "Test" + string(rune('A'+i)), Score: v, MaxScore: 10, Weight: 1}
		}
		return out
	}
	tests := []struct {
		name     string
		results  *score.Results
		maxScore uint32
		want     []string // substrings the problem must contain; empty means no problem
	}{
		{
			name:     "ZeroScore",
			results:  &score.Results{Scores: scores(0, 0)},
			maxScore: ci.MaxSkeletonScore,
		},
		{
			name:     "AtThreshold",
			results:  &score.Results{Scores: []*score.Score{{TestName: "TestA", Score: 1, MaxScore: 20, Weight: 1}}},
			maxScore: ci.MaxSkeletonScore,
		},
		{
			name:     "TooHigh",
			results:  &score.Results{Scores: scores(10, 0, 0, 0)},
			maxScore: ci.MaxSkeletonScore,
			want:     []string{"scored 25%", "at most 5%", "TestA (10/10)"},
		},
		{
			name:     "AllTestsPass",
			results:  &score.Results{Scores: scores(10, 10)},
			maxScore: ci.MaxSkeletonScore,
			want:     []string{"scored 100%", "TestA (10/10)", "TestB (10/10)"},
		},
		{
			name: "FailedRun",
			results: &score.Results{
				BuildInfo: &score.BuildInfo{Status: score.RunStatus_NO_SCORES},
				Scores:    scores(0),
			},
			maxScore: ci.MaxSkeletonScore,
			want:     []string{"NO_SCORES"},
		},
		{
			name: "BuildFailure",
			results: &score.Results{
				BuildInfo: &score.BuildInfo{Status: score.RunStatus_BUILD_FAILURE},
			},
			maxScore: ci.MaxSkeletonScore,
			want:     []string{"BUILD_FAILURE"},
		},
		{
			name:     "CustomThreshold",
			results:  &score.Results{Scores: scores(10, 0)},
			maxScore: 60,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ci.SkeletonProblem(tc.results, tc.maxScore)
			if len(tc.want) == 0 {
				if got != "" {
					t.Errorf("SkeletonProblem() = %q, want no problem", got)
				}
				return
			}
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("SkeletonProblem() = %q, want it to contain %q", got, want)
				}
			}
		})
	}
}
