package assignments

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestCourseRepositoryIssues(t *testing.T) {
	testsDir := t.TempDir()
	assignmentsDir := t.TempDir()

	writeRepoFile(t, testsDir, "lab1", assignmentFile)
	writeRepoFile(t, testsDir, "lab1", testsFile)
	makeRepoDir(t, assignmentsDir, "lab1")

	writeRepoFile(t, testsDir, "lab2", assignmentFile)
	writeRepoFile(t, testsDir, "lab2", criteriaFile)

	makeRepoDir(t, assignmentsDir, "lab3")

	makeRepoDir(t, testsDir, "lab4")
	makeRepoDir(t, assignmentsDir, "lab4")

	writeRepoFile(t, testsDir, "lab5", assignmentFile)
	makeRepoDir(t, assignmentsDir, "lab5")

	// Shared packages and repository metadata are not assignment folders.
	writeRepoFile(t, testsDir, "internal/pkg", testsFile)
	makeRepoDir(t, testsDir, scriptsDir)
	makeRepoDir(t, testsDir, ".github")

	parsed := []*qf.Assignment{
		{Name: "lab1"},
		{Name: "lab2", Reviewers: 1},
		{Name: "lab5", Reviewers: 1},
	}
	got, err := courseRepositoryIssues(testsDir, assignmentsDir, parsed)
	if err != nil {
		t.Fatal(err)
	}
	want := []RepoIssue{
		{Assignment: "lab2", File: "lab2", Problem: `assignment folder is missing from "assignments" repository`},
		{Assignment: "lab3", File: "lab3", Problem: `assignment folder is missing from "tests" repository`},
		{Assignment: "lab4", File: "lab4/assignment.json", Problem: `missing "lab4/assignment.json"`},
		{Assignment: "lab4", File: "lab4", Problem: "missing tests.json or criteria.json"},
		{Assignment: "lab5", File: "lab5", Problem: "missing tests.json or criteria.json"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("courseRepositoryIssues() mismatch (-want +got):\n%s", diff)
	}
}

func TestValidateCourseRepositoriesCountsAllIssues(t *testing.T) {
	testsDir := t.TempDir()
	assignmentsDir := t.TempDir()
	writeFile(t, testsDir, "lab1", assignmentFile, j1)
	makeRepoDir(t, assignmentsDir, "lab1")

	ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
	got, err := ValidateCourseRepositories(ctx, testsDir, assignmentsDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	// The tests repository content check reports the missing tests.json. The
	// cross-repository check finds no additional problem for the aligned folders.
	if got != 1 {
		t.Errorf("ValidateCourseRepositories() = %d, want 1", got)
	}
}

func TestValidateCourseRepositoriesDoesNotDuplicateMissingAssignment(t *testing.T) {
	testsDir := t.TempDir()
	assignmentsDir := t.TempDir()
	writeFile(t, testsDir, "lab1", testsFile, testJson)
	makeRepoDir(t, assignmentsDir, "lab1")

	ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
	got, err := ValidateCourseRepositories(ctx, testsDir, assignmentsDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Errorf("ValidateCourseRepositories() = %d, want one missing-assignment issue", got)
	}
}

func TestCourseRepositoryIssuesMissingRoot(t *testing.T) {
	_, err := courseRepositoryIssues(filepath.Join(t.TempDir(), "missing"), t.TempDir(), nil)
	if err == nil {
		t.Fatal("courseRepositoryIssues() error = nil, want missing repository error")
	}
}

func TestUpdateFromTestsRepoReportsAlignmentIssues(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	testsDir := t.TempDir()
	assignmentsDir := t.TempDir()
	writeFile(t, testsDir, "lab1", assignmentFile, j1)
	writeFile(t, testsDir, "lab1", testsFile, testJson)
	scmClient := &cloneOnlySCM{directories: map[string]string{
		qf.TestsRepo:       testsDir,
		qf.AssignmentsRepo: assignmentsDir,
	}}

	ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
	got, err := UpdateFromTestsRepo(ctx, &ci.Local{}, db, scmClient, course)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Errorf("UpdateFromTestsRepo() issue count = %d, want 1", got)
	}
	stored, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 || stored[0].GetName() != "lab1" {
		t.Errorf("stored assignments = %+v, want lab1", stored)
	}
}

func TestUpdateFromTestsRepoPersistsBeforeAlignmentFailure(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	testsDir := t.TempDir()
	writeFile(t, testsDir, "lab1", assignmentFile, j1)
	writeFile(t, testsDir, "lab1", testsFile, testJson)
	scmClient := &cloneOnlySCM{
		directories: map[string]string{qf.TestsRepo: testsDir},
		errors:      map[string]error{qf.AssignmentsRepo: errors.New("clone failed")},
	}

	ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
	_, err := UpdateFromTestsRepo(ctx, &ci.Local{}, db, scmClient, course)
	if err == nil {
		t.Fatal("UpdateFromTestsRepo() error = nil, want assignments clone error")
	}
	stored, dbErr := db.GetAssignmentsByCourse(course.GetID())
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	if len(stored) != 1 || stored[0].GetName() != "lab1" {
		t.Errorf("stored assignments = %+v, want lab1 despite validation failure", stored)
	}
}

type cloneOnlySCM struct {
	scm.SCM
	directories map[string]string
	errors      map[string]error
}

func (s *cloneOnlySCM) Clone(_ context.Context, options *scm.CloneOptions) (string, error) {
	if err := s.errors[options.Repository]; err != nil {
		return "", err
	}
	return s.directories[options.Repository], nil
}

func makeRepoDir(t *testing.T, root string, elements ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(append([]string{root}, elements...)...), 0o750); err != nil {
		t.Fatal(err)
	}
}

func writeRepoFile(t *testing.T, root string, elements ...string) {
	t.Helper()
	path := filepath.Join(append([]string{root}, elements...)...)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
}
