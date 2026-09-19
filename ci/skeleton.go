package ci

import (
	"fmt"
	"strings"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
)

// skeletonOwner is the job owner of a skeleton run, and thereby the name of the
// directory the skeleton code is copied to inside the job's temporary
// directory; it must not collide with the "tests" and "assignments" read-only
// mounts; see RunData.parseTestRunnerScript.
const skeletonOwner = "skeleton"

// MaxSkeletonScore is the highest total score the skeleton (handout) code in
// the course's assignments repository may achieve before the test environment
// is suspected of being broken.
//
// Students start from the skeleton code, so it should fail nearly every test; a
// higher score means the tests do not exercise what the students are asked to
// write, or that the solution leaked into the handout. The threshold is not
// zero because a test may award points for something the skeleton happens to
// satisfy, such as compiling or a zero-valued expected result.
const MaxSkeletonScore uint32 = 5

// NewSkeletonRun returns the run data for testing the course's skeleton
// (handout) code, taken from the local clone of the assignments repository,
// against the course's own tests; see NewLocalRun for the commit ID.
func NewSkeletonRun(course *qf.Course, assignment *qf.Assignment, commitID string) *RunData {
	return newSkeletonRun(course, assignment, commitID, "")
}

// NewSkeletonRunFromSnapshot returns the run data for a skeleton run that reads
// the course's tests and skeleton code from snapshotDir, a private copy of the
// two course repositories, instead of from the course's clone directory; see
// RunData.snapshotDir.
func NewSkeletonRunFromSnapshot(course *qf.Course, assignment *qf.Assignment, commitID, snapshotDir string) *RunData {
	return newSkeletonRun(course, assignment, commitID, snapshotDir)
}

// newSkeletonRun returns the run data for a skeleton run against the course
// repositories in snapshotDir, or against the course's clone directory when
// snapshotDir is empty.
//
// The skeleton code is the assignments repository that the run also mounts, so
// the submission directory is derived from the run's own course repositories
// and cannot name a different copy of that repository.
func newSkeletonRun(course *qf.Course, assignment *qf.Assignment, commitID, snapshotDir string) *RunData {
	runData := NewLocalRun(course, assignment, skeletonOwner, "", commitID)
	runData.snapshotDir = snapshotDir
	runData.SubmissionDir = runData.assignmentsDir()
	return runData
}

// SkeletonReport is the verdict on a skeleton run; see CheckSkeleton.
type SkeletonReport struct {
	// Problem explains on one line why the test environment is suspect. It is
	// empty for a healthy environment.
	Problem string
	// PassingTests names the tests that awarded the skeleton code a non-zero
	// score, each as "TestName (score/max)". A few may do so on a healthy
	// skeleton; when there is a Problem, they are the tests to look at.
	PassingTests []string
}

// CheckSkeleton judges the results of a skeleton run: the environment is
// healthy if the run completed and the skeleton code scored no more than
// maxScore.
func CheckSkeleton(results *score.Results, maxScore uint32) SkeletonReport {
	report := SkeletonReport{PassingTests: passingTests(results)}
	switch sum := results.Sum(); {
	case results.Failed():
		report.Problem = fmt.Sprintf("the skeleton run failed with status %s; the tests never produced a usable result",
			results.GetBuildInfo().GetStatus())
	case sum > maxScore:
		report.Problem = fmt.Sprintf("skeleton code scored %d%%, expected at most %d%%", sum, maxScore)
	}
	return report
}

// Healthy reports whether the run found nothing wrong with the test environment.
func (r SkeletonReport) Healthy() bool {
	return r.Problem == ""
}

// String returns the report on a single line: the problem, followed by the
// passing tests that explain it. A healthy report is the empty string.
func (r SkeletonReport) String() string {
	if r.Problem == "" || len(r.PassingTests) == 0 {
		return r.Problem
	}
	return r.Problem + "; tests passing on the skeleton code: " + strings.Join(r.PassingTests, ", ")
}

// SkeletonProblem returns a one-line explanation of the problem found by
// CheckSkeleton, or the empty string for a healthy test environment.
func SkeletonProblem(results *score.Results, maxScore uint32) string {
	return CheckSkeleton(results, maxScore).String()
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
