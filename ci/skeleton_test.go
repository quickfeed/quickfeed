package ci_test

import (
	"path/filepath"
	"slices"
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

// TestNewSkeletonRunFromSnapshot checks that a snapshot run reads the skeleton
// code from the snapshot and never from the course's clone directory.
func TestNewSkeletonRunFromSnapshot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", root)
	course := &qf.Course{Code: "QF101", ScmOrganizationName: "qf101-2025"}
	assignment := &qf.Assignment{Name: "lab1"}
	snapshotDir := t.TempDir()

	runData := ci.NewSkeletonRunFromSnapshot(course, assignment, "abc1234", snapshotDir)
	wantDir := filepath.Join(snapshotDir, qf.AssignmentsRepo)
	if runData.SubmissionDir != wantDir {
		t.Errorf("NewSkeletonRunFromSnapshot() submission dir = %q, want %q", runData.SubmissionDir, wantDir)
	}
	if strings.HasPrefix(runData.SubmissionDir, course.CloneDir()) {
		t.Errorf("NewSkeletonRunFromSnapshot() submission dir = %q, want it outside the course clone dir %q",
			runData.SubmissionDir, course.CloneDir())
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

// TestCheckSkeleton checks the structured verdict that cmd/qcm formats one line
// per passing test, and that its one-line form is what SkeletonProblem returns.
func TestCheckSkeleton(t *testing.T) {
	results := &score.Results{Scores: []*score.Score{
		{TestName: "TestA", Score: 10, MaxScore: 10, Weight: 1},
		{TestName: "TestB", Score: 0, MaxScore: 10, Weight: 1},
		{TestName: "TestC", Score: 3, MaxScore: 10, Weight: 1},
	}}
	report := ci.CheckSkeleton(results, ci.MaxSkeletonScore)
	if report.Healthy() {
		t.Fatalf("CheckSkeleton(43%%) = %+v, want a problem", report)
	}
	wantPassing := []string{"TestA (10/10)", "TestC (3/10)"}
	if !slices.Equal(report.PassingTests, wantPassing) {
		t.Errorf("CheckSkeleton() passing tests = %q, want %q", report.PassingTests, wantPassing)
	}
	if got, want := report.String(), ci.SkeletonProblem(results, ci.MaxSkeletonScore); got != want {
		t.Errorf("CheckSkeleton().String() = %q, want SkeletonProblem's %q", got, want)
	}
	if !strings.HasPrefix(report.String(), report.Problem+"; tests passing on the skeleton code: TestA (10/10), TestC (3/10)") {
		t.Errorf("CheckSkeleton().String() = %q, want the problem followed by the passing tests", report.String())
	}

	// A healthy skeleton may still have a passing test; the report names it
	// without declaring a problem.
	healthy := ci.CheckSkeleton(&score.Results{Scores: []*score.Score{
		{TestName: "TestLint", Score: 1, MaxScore: 1, Weight: 1},
		{TestName: "TestWork", Score: 0, MaxScore: 100, Weight: 30},
	}}, ci.MaxSkeletonScore)
	if !healthy.Healthy() || healthy.String() != "" {
		t.Errorf("CheckSkeleton(healthy) = %+v, want no problem and an empty String()", healthy)
	}
	if !slices.Equal(healthy.PassingTests, []string{"TestLint (1/1)"}) {
		t.Errorf("CheckSkeleton(healthy) passing tests = %q, want TestLint", healthy.PassingTests)
	}
}
