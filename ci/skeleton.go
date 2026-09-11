package ci

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
)

// skeletonOwner is the job owner recorded for a skeleton run. It also names the
// repository the skeleton code is copied to inside the job's temporary
// directory, which must not collide with the "tests" and "assignments"
// read-only mounts; see RunData.parseTestRunnerScript.
const skeletonOwner = "skeleton"

// MaxSkeletonScore is the highest total score the skeleton (handout) code in
// the course's assignments repository may achieve before the test environment
// is suspected of being broken.
//
// The skeleton code is what students start from, so it should fail nearly every
// test. A higher score means the tests are not actually exercising the parts
// the students are asked to write, or that the solution leaked into the
// handout. The threshold is not zero because an exact zero is hard to reach in
// practice: a test may award points for compiling, for a case the skeleton
// happens to satisfy, or for a table entry whose expected value is the zero
// value. A few percent is therefore expected; more than that is a problem.
const MaxSkeletonScore uint32 = 5

// NewSkeletonRun returns the run data for testing the course's skeleton
// (handout) code, taken from the local clone of the assignments repository,
// against the course's own tests.
//
// Callers must pass a commit ID that is unique for the run, because
// RunData.String names the container, and starting a container whose name is
// already taken fails with ErrConflict.
func NewSkeletonRun(course *qf.Course, assignment *qf.Assignment, commitID string) *RunData {
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: course.GetScmOrganizationName()}
	return &RunData{
		Course:        course,
		Assignment:    assignment,
		SubmissionDir: filepath.Join(course.CloneDir(), qf.AssignmentsRepo),
		Repo: &qf.Repository{
			HTMLURL:  repo.StudentRepoURL(skeletonOwner),
			RepoType: qf.Repository_USER,
		},
		JobOwner: skeletonOwner,
		CommitID: commitID,
	}
}

// SkeletonProblem returns the empty string if the given results describe a
// healthy test environment: the run completed, and the skeleton code scored no
// more than maxScore. Otherwise it returns a one-line explanation of the
// problem, addressed to the teaching staff.
func SkeletonProblem(results *score.Results, maxScore uint32) string {
	if results.Failed() {
		return fmt.Sprintf("the skeleton run failed with status %s; the tests never produced a usable result",
			results.GetBuildInfo().GetStatus())
	}
	sum := results.Sum()
	if sum <= maxScore {
		return ""
	}
	problem := fmt.Sprintf("skeleton code scored %d%%, expected at most %d%%", sum, maxScore)
	if passing := passingTests(results); len(passing) > 0 {
		problem += "; tests passing on the skeleton code: " + strings.Join(passing, ", ")
	}
	return problem
}

// passingTests returns the tests that awarded the skeleton code a non-zero
// score, formatted as "TestName (score/max)".
func passingTests(results *score.Results) []string {
	var passing []string
	for _, sc := range results.Scores {
		if sc.GetScore() > 0 {
			passing = append(passing, fmt.Sprintf("%s (%d/%d)", sc.GetTestName(), sc.GetScore(), sc.GetMaxScore()))
		}
	}
	return passing
}
