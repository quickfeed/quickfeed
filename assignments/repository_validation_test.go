package assignments

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
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

	// Test code pushed without assignment.json and without the counterpart
	// folder in the assignments repository; QuickFeed cannot tell this apart
	// from shared course code, so it reports the folder once.
	writeRepoFile(t, testsDir, "lab6", "lab6_test.go")

	// A shared package is reported the same way until it is listed in the
	// ignore file; see TestCourseRepositoryIssuesIgnoreFile.
	writeRepoFile(t, testsDir, "internal/pkg", testsFile)
	// Repository metadata is never an assignment folder.
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
		{Assignment: "internal", File: "internal", Problem: `no assignment configuration found; add "internal/assignment.json" if this is an assignment, or list the folder in .quickfeedignore`},
		{Assignment: "lab2", File: "lab2", Problem: `assignment folder is missing from "assignments" repository`, Transient: true},
		{Assignment: "lab3", File: "lab3", Problem: `assignment folder is missing from "tests" repository; add it or list the folder in .quickfeedignore`, Transient: true},
		{Assignment: "lab4", File: "lab4", Problem: `no assignment configuration found; add "lab4/assignment.json" if this is an assignment, or list the folder in .quickfeedignore`},
		{Assignment: "lab5", File: "lab5", Problem: "missing tests.json or criteria.json"},
		{Assignment: "lab6", File: "lab6", Problem: `no assignment configuration found; add "lab6/assignment.json" if this is an assignment, or list the folder in .quickfeedignore`},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("courseRepositoryIssues() mismatch (-want +got):\n%s", diff)
	}
}

func TestCourseRepositoryIssuesIgnoreFile(t *testing.T) {
	testsDir := t.TempDir()
	assignmentsDir := t.TempDir()

	writeRepoFile(t, testsDir, "lab1", assignmentFile)
	writeRepoFile(t, testsDir, "lab1", testsFile)
	makeRepoDir(t, assignmentsDir, "lab1")

	// Shared course code in the tests repository, and handout material that
	// students receive but that has no tests; neither is an assignment.
	writeRepoFile(t, testsDir, "internal", "helper.go")
	makeRepoDir(t, assignmentsDir, "internal")
	makeRepoDir(t, assignmentsDir, "resources")

	writeIgnoreFile(
		t, testsDir,
		"# folders that are not assignments",
		"internal",
		"",
		"resources",
	)

	got, err := courseRepositoryIssues(testsDir, assignmentsDir, []*qf.Assignment{{Name: "lab1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("courseRepositoryIssues() = %+v, want no issues", got)
	}
}

func TestCourseRepositoryIssuesInvalidIgnoreEntry(t *testing.T) {
	testsDir := t.TempDir()
	assignmentsDir := t.TempDir()

	writeRepoFile(t, testsDir, "internal", "helper.go")
	writeIgnoreFile(t, testsDir, "internal/*")

	got, err := courseRepositoryIssues(testsDir, assignmentsDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []RepoIssue{
		{File: ".quickfeedignore", Problem: `invalid entry "internal/*": only top-level folder names are supported`},
		// The invalid entry does not hide the folder it was meant to exclude.
		{Assignment: "internal", File: "internal", Problem: `no assignment configuration found; add "internal/assignment.json" if this is an assignment, or list the folder in .quickfeedignore`},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("courseRepositoryIssues() mismatch (-want +got):\n%s", diff)
	}
}

func TestCourseRepositoryIssuesMissingRoot(t *testing.T) {
	_, err := courseRepositoryIssues(filepath.Join(t.TempDir(), "missing"), t.TempDir(), nil)
	if err == nil {
		t.Fatal("courseRepositoryIssues() error = nil, want missing repository error")
	}
}

// newCourse returns a course backed by a fresh test database, ready to be
// updated from a pair of local directories standing in for the two course
// repositories.
func newCourse(t *testing.T) (database.Database, *qf.Course) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	t.Cleanup(cleanup)
	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)
	return db, course
}

// TestUpdateFromCourseRepositoriesIssueCount checks that the returned count
// covers both the tests repository's content problems and the alignment of the
// two repositories, without reporting the same folder twice.
func TestUpdateFromCourseRepositoriesIssueCount(t *testing.T) {
	tests := []struct {
		name            string
		writeRepo       func(t *testing.T, testsDir, assignmentsDir string)
		wantCount       int
		wantAssignments []string
	}{
		{
			// The content check reports the missing tests.json; the aligned
			// folders add nothing to it.
			name: "MissingTestsFile",
			writeRepo: func(t *testing.T, testsDir, assignmentsDir string) {
				writeFile(t, testsDir, "lab1", assignmentFile, j1)
				makeRepoDir(t, assignmentsDir, "lab1")
			},
			wantCount:       1,
			wantAssignments: []string{"lab1"},
		},
		{
			// A folder configured by tests.json alone is missing its
			// assignment.json; it must be reported once, not once per check.
			name: "MissingAssignmentFile",
			writeRepo: func(t *testing.T, testsDir, assignmentsDir string) {
				writeFile(t, testsDir, "lab1", testsFile, testJson)
				makeRepoDir(t, assignmentsDir, "lab1")
			},
			// The folder holds no parsable assignment, so nothing is stored.
			wantCount: 1,
		},
		{
			// An assignment folder that has not reached the assignments
			// repository yet: one alignment issue, no content issue.
			name: "MissingFromAssignmentsRepository",
			writeRepo: func(t *testing.T, testsDir, _ string) {
				writeFile(t, testsDir, "lab1", assignmentFile, j1)
				writeFile(t, testsDir, "lab1", testsFile, testJson)
			},
			wantCount:       1,
			wantAssignments: []string{"lab1"},
		},
		{
			name: "AlignedRepositories",
			writeRepo: func(t *testing.T, testsDir, assignmentsDir string) {
				writeFile(t, testsDir, "lab1", assignmentFile, j1)
				writeFile(t, testsDir, "lab1", testsFile, testJson)
				makeRepoDir(t, assignmentsDir, "lab1")
			},
			wantCount:       0,
			wantAssignments: []string{"lab1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, course := newCourse(t)
			testsDir, assignmentsDir := t.TempDir(), t.TempDir()
			tt.writeRepo(t, testsDir, assignmentsDir)
			scmClient := &cloneOnlySCM{directories: map[string]string{
				qf.TestsRepo:       testsDir,
				qf.AssignmentsRepo: assignmentsDir,
			}}

			ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
			got, err := UpdateFromCourseRepositories(ctx, &ci.Local{}, db, scmClient, course)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.wantCount {
				t.Errorf("UpdateFromCourseRepositories() issue count = %d, want %d", got, tt.wantCount)
			}
			assertStoredAssignments(t, db, course, tt.wantAssignments)
		})
	}
}

// TestUpdateFromCourseRepositoriesAssignmentsCloneFailure checks the failure
// policy for the assignments repository: the update itself still completes, and
// only the cross-repository comparison is skipped.
func TestUpdateFromCourseRepositoriesAssignmentsCloneFailure(t *testing.T) {
	db, course := newCourse(t)
	testsDir := t.TempDir()
	writeFile(t, testsDir, "lab1", assignmentFile, j1)
	scmClient := &cloneOnlySCM{
		directories: map[string]string{qf.TestsRepo: testsDir},
		errors:      map[string]error{qf.AssignmentsRepo: errors.New("clone failed")},
	}

	ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
	got, err := UpdateFromCourseRepositories(ctx, &ci.Local{}, db, scmClient, course)
	if err != nil {
		t.Fatalf("UpdateFromCourseRepositories() error = %v, want nil despite the assignments clone failure", err)
	}
	// The missing tests.json is still reported; the alignment issue that the
	// empty assignments repository would have produced is not.
	if got != 1 {
		t.Errorf("UpdateFromCourseRepositories() issue count = %d, want 1", got)
	}
	assertStoredAssignments(t, db, course, []string{"lab1"})
}

// TestUpdateFromCourseRepositoriesTestsCloneFailure checks that a failure to
// clone the tests repository aborts the update: nothing can be done without it.
func TestUpdateFromCourseRepositoriesTestsCloneFailure(t *testing.T) {
	db, course := newCourse(t)
	scmClient := &cloneOnlySCM{errors: map[string]error{qf.TestsRepo: errors.New("clone failed")}}

	ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
	if _, err := UpdateFromCourseRepositories(ctx, &ci.Local{}, db, scmClient, course); err == nil {
		t.Fatal("UpdateFromCourseRepositories() error = nil, want tests clone error")
	}
	if calls := scmClient.calls; !slices.Equal(calls, []string{qf.TestsRepo}) {
		t.Errorf("Clone() calls = %v, want only %q", calls, qf.TestsRepo)
	}
}

// TestUpdateFromCourseRepositoriesClonesBothRepositories checks that the update
// refreshes both local clones, whichever repository the push came from.
func TestUpdateFromCourseRepositoriesClonesBothRepositories(t *testing.T) {
	db, course := newCourse(t)
	testsDir, assignmentsDir := t.TempDir(), t.TempDir()
	writeFile(t, testsDir, "lab1", assignmentFile, j1)
	writeFile(t, testsDir, "lab1", testsFile, testJson)
	makeRepoDir(t, assignmentsDir, "lab1")
	scmClient := &cloneOnlySCM{directories: map[string]string{
		qf.TestsRepo:       testsDir,
		qf.AssignmentsRepo: assignmentsDir,
	}}

	ctx := qlog.NewContext(t.Context(), qtest.Logger(t))
	if _, err := UpdateFromCourseRepositories(ctx, &ci.Local{}, db, scmClient, course); err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{qf.TestsRepo, qf.AssignmentsRepo}
	if diff := cmp.Diff(wantCalls, scmClient.calls); diff != "" {
		t.Errorf("Clone() calls mismatch (-want +got):\n%s", diff)
	}
}

func assertStoredAssignments(t *testing.T, db database.Database, course *qf.Course, want []string) {
	t.Helper()
	stored, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(stored))
	for _, assignment := range stored {
		got = append(got, assignment.GetName())
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("stored assignments mismatch (-want +got):\n%s", diff)
	}
}

type cloneOnlySCM struct {
	scm.SCM
	directories map[string]string
	errors      map[string]error
	calls       []string // repositories passed to Clone, in order
}

func (s *cloneOnlySCM) Clone(_ context.Context, options *scm.CloneOptions) (string, error) {
	s.calls = append(s.calls, options.Repository)
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

func writeIgnoreFile(t *testing.T, root string, lines ...string) {
	t.Helper()
	contents := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(root, ignoreFile), []byte(contents), 0o600); err != nil {
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
