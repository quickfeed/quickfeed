package ci_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
)

// localTestOrg is the course organization used by the local test runs below.
// The repositories it names are created in a temporary directory; no network
// or Docker daemon is involved.
const localTestOrg = "qf-local-2025"

// localRunScript prints a single valid score line for TestLocalRun. The Local
// runner replaces the process environment with the job's environment, so the
// script must only use bash builtins; see ci.Local.Run.
const localRunScript = `#image/dummy

echo
echo "{\"Secret\":\"$QUICKFEED_SESSION_SECRET\",\"TestName\":\"TestLocalRun\",\"Score\":5,\"MaxScore\":10,\"Weight\":1}"
echo "{\"Secret\":\"$QUICKFEED_SESSION_SECRET\",\"TestName\":\"TestNotExpected\",\"Score\":10,\"MaxScore\":10,\"Weight\":1}"
`

// setupLocalCourse creates the course's tests and assignments repositories for
// the given assignments below a fresh repository path, and returns the course.
func setupLocalCourse(t *testing.T, runScript string, labs ...string) *qf.Course {
	t.Helper()
	root := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", root)
	for _, repo := range []string{qf.TestsRepo, qf.AssignmentsRepo} {
		for _, lab := range labs {
			mkdirAll(t, filepath.Join(root, localTestOrg, repo, lab))
		}
	}
	if runScript != "" {
		scripts := filepath.Join(root, localTestOrg, qf.TestsRepo, "scripts")
		mkdirAll(t, scripts)
		writeFile(t, filepath.Join(scripts, "run.sh"), runScript)
	}
	return &qf.Course{Code: "QF101", ScmOrganizationName: localTestOrg}
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestRunTestsSubmissionDir runs the tests for a local submission directory
// with the Local runner and without an SCM client, which is the path qcm uses
// to check a course's test environment.
func TestRunTestsSubmissionDir(t *testing.T) {
	course := setupLocalCourse(t, localRunScript, "lab1")
	submissionDir := t.TempDir()
	mkdirAll(t, filepath.Join(submissionDir, "lab1"))
	writeFile(t, filepath.Join(submissionDir, "lab1", "lab1.go"), "package lab1\n")

	runData := &ci.RunData{
		Course: course,
		Assignment: &qf.Assignment{
			Name: "lab1",
			ExpectedTests: []*qf.TestInfo{
				{TestName: "TestLocalRun", MaxScore: 10, Weight: 1},
			},
		},
		Repo: &qf.Repository{
			HTMLURL:  qf.RepoURL{ProviderURL: "github.com", Organization: localTestOrg}.StudentRepoURL("local"),
			RepoType: qf.Repository_USER,
		},
		JobOwner:      "local",
		CommitID:      "deadbeef",
		SubmissionDir: submissionDir,
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	// The local path must not need an SCM client.
	results, err := runData.RunTests(ctx, nil, &ci.Local{})
	if err != nil {
		t.Fatal(err)
	}
	if got := results.GetBuildInfo().GetStatus(); got != score.RunStatus_SUCCESS {
		t.Errorf("RunTests() status = %s, want SUCCESS\nbuild log:\n%s", got, results.GetBuildInfo().GetBuildLog())
	}
	// TestNotExpected is not listed in the assignment's expected tests, so
	// ExtractResults discards it; only TestLocalRun's 5 of 10 counts.
	if got, want := results.Sum(), uint32(50); got != want {
		t.Errorf("RunTests() sum = %d, want %d", got, want)
	}
	if len(results.Scores) != 1 {
		t.Errorf("RunTests() recorded %d scores, want 1: %+v", len(results.Scores), results.Scores)
	}
}

// TestRunTestsSubmissionDirMissingCourseRepos checks that a missing course
// repository is reported instead of dereferencing the absent SCM client.
func TestRunTestsSubmissionDirMissingCourseRepos(t *testing.T) {
	// The assignments repository is missing, so cloning it would be required.
	course := setupLocalCourse(t, localRunScript, "lab1")
	if err := os.RemoveAll(filepath.Join(course.CloneDir(), qf.AssignmentsRepo)); err != nil {
		t.Fatal(err)
	}
	submissionDir := t.TempDir()
	mkdirAll(t, filepath.Join(submissionDir, "lab1"))

	runData := &ci.RunData{
		Course:     course,
		Assignment: &qf.Assignment{Name: "lab1"},
		Repo: &qf.Repository{
			HTMLURL:  qf.RepoURL{ProviderURL: "github.com", Organization: localTestOrg}.StudentRepoURL("local"),
			RepoType: qf.Repository_USER,
		},
		JobOwner:      "local",
		CommitID:      "deadbeef",
		SubmissionDir: submissionDir,
	}
	if _, err := runData.RunTests(t.Context(), nil, &ci.Local{}); err == nil {
		t.Fatal("RunTests() error = nil, want an error for the missing course repositories")
	}
}
