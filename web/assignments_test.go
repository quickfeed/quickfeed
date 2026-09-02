package web_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestUpdateAssignments(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())
	course := qtest.MockCourses[0]
	user := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, user, course)

	tests := []struct {
		name    string
		request *qf.CourseRequest
		wantErr error
	}{
		{
			name: "Invalid course ID (permission denied)",
			request: &qf.CourseRequest{
				CourseID: 111,
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for UpdateAssignments: not teacher")),
		},
		{
			// The mock's tests repository is empty, so the clone inside
			// UpdateFromCourseRepositories fails; previously this failure was
			// silently ignored and the RPC continued to the assignments repository.
			name: "Valid course ID but failed to clone tests repository",
			request: &qf.CourseRequest{
				CourseID: course.GetID(),
			},
			wantErr: connect.NewError(connect.CodeInternal, errors.New("failed to update assignments from tests repository")),
		},
	}

	ctx := client.Context(t, user)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.UpdateAssignments(ctx, test.request)
			if qtest.CheckCode(t, err, test.wantErr) {
				return // cannot continue since resp is invalid
			}
		})
	}
}

// TestUpdateAssignmentsIssueCount checks that a successful UpdateAssignments
// reports the number of content issues found in the course's tests repository.
//
// Each test gets its own database, mock client and repository path, since the
// failure cases above leave a partially cloned course directory behind.
func TestUpdateAssignmentsIssueCount(t *testing.T) {
	// lab1 is committed by qtest.PrepareGitRepo below, so all fixture content
	// for a test case must live under lab1, however many issues it should produce.
	tests := []struct {
		name      string
		lab1Files map[string]string
		wantCount uint32
	}{
		{
			name: "Tests repository without issues",
			lab1Files: map[string]string{
				"lab1/assignment.json": `{"order":1,"deadline":"24-01-2019T14:00"}`,
				"lab1/tests.json":      `[{"TestName":"TestOne","MaxScore":10,"Weight":1}]`,
			},
			wantCount: 0,
		},
		{
			name: "Tests repository with issues",
			lab1Files: map[string]string{
				// Missing tests.json is reported as one issue, since an
				// auto-graded assignment without expected tests always scores zero.
				"lab1/assignment.json": `{"order":1,"deadline":"24-01-2019T14:00"}`,
				// A benchmark without a heading is reported as a second issue.
				"lab1/criteria.json": `[{"heading":"","criteria":[{"description":"must pass"}]}]`,
			},
			wantCount: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The mock SCM clones with go-git from file://$QUICKFEED_REPOSITORY_PATH,
			// which it reads when the client below is created. The repositories are
			// therefore prepared under srcPath while that path is in effect.
			srcPath := t.TempDir()
			t.Setenv("QUICKFEED_REPOSITORY_PATH", srcPath)

			// qtest.PrepareGitRepo copies src/repo into dst/repo and commits
			// its lab1 folder, so the raw fixture content is written to a
			// scratch directory first, following the layout it expects.
			raw := t.TempDir()
			writeFiles(t, filepath.Join(raw, qf.TestsRepo), test.lab1Files)
			// The assignments repository is cloned as well, but its content is not read.
			writeFiles(t, filepath.Join(raw, qf.AssignmentsRepo), map[string]string{
				"lab1/lab1.go": "package lab1\n",
			})
			orgPath := filepath.Join(srcPath, qtest.MockOrg)
			qtest.PrepareGitRepo(t, raw, orgPath, qf.TestsRepo)
			qtest.PrepareGitRepo(t, raw, orgPath, qf.AssignmentsRepo)

			db, cleanup := qtest.TestDB(t)
			defer cleanup()
			client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())
			// Point course.CloneDir() at an empty path, so that the repositories
			// prepared above are cloned into a destination distinct from the source.
			t.Setenv("QUICKFEED_REPOSITORY_PATH", t.TempDir())

			course := qtest.MockCourses[0]
			user := qtest.CreateFakeUser(t, db)
			qtest.CreateCourse(t, db, user, course)

			ctx := client.Context(t, user)
			issues, err := client.UpdateAssignments(ctx, &qf.CourseRequest{CourseID: course.GetID()})
			if err != nil {
				t.Fatal(err)
			}
			if issues.GetCount() != test.wantCount {
				t.Errorf("UpdateAssignments() count = %d, want %d", issues.GetCount(), test.wantCount)
			}
		})
	}
}

// writeFiles writes the given files, keyed by their repository-relative path,
// under dir. It only creates plain files; qtest.PrepareGitRepo turns the
// result into the git repository the mock SCM clones from.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
