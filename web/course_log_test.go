package web_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// seedCourseLog writes n records for course to a fresh store, alternating
// between two repositories and INFO/ERROR levels, through the same public
// path production code uses (a *slog.Logger scoped by qlog.CourseAttrs).
func seedCourseLog(t *testing.T, dir string, course *qf.Course, n int) *courselog.Store {
	t.Helper()
	store, err := courselog.NewStore(dir, qtest.Logger(t))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	logger := slog.New(courselog.NewHandler(store)).With(qlog.CourseAttrs(course)...)
	for i := range n {
		repo, level := "student-a", slog.LevelInfo
		if i%2 == 1 {
			repo, level = "student-b", slog.LevelError
		}
		logger.Log(context.Background(), level, "test record", label.Repository, repo)
	}
	return store
}

func TestGetCourseLog(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)
	student := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student, course)
	admin := qtest.CreateFakeUser(t, db)
	qtest.UpdateUser(t, db, &qf.User{ID: admin.GetID(), IsAdmin: true})

	otherCourse := qtest.MockCourses[1]
	qtest.CreateCourse(t, db, teacher, otherCourse)

	store := seedCourseLog(t, t.TempDir(), course, 6) // 3 student-a/INFO, 3 student-b/ERROR
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithCourseLogStore(store))

	t.Run("teacher success", func(t *testing.T) {
		got, err := client.GetCourseLog(client.Context(t, teacher), &qf.CourseLogRequest{CourseID: course.GetID(), Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.GetEntries()) != 6 {
			t.Errorf("len(Entries) = %d, want 6", len(got.GetEntries()))
		}
		wantRepos := []string{"student-a", "student-b"}
		gotRepos := got.GetRepositories()
		if len(gotRepos) != len(wantRepos) || gotRepos[0] != wantRepos[0] || gotRepos[1] != wantRepos[1] {
			t.Errorf("Repositories = %v, want %v", gotRepos, wantRepos)
		}
		if got.GetTruncated() {
			t.Error("Truncated = true, want false: the limit was not reached")
		}
	})

	t.Run("student denied", func(t *testing.T) {
		_, err := client.GetCourseLog(client.Context(t, student), &qf.CourseLogRequest{CourseID: course.GetID()})
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Errorf("code = %v, want PermissionDenied", connect.CodeOf(err))
		}
	})

	t.Run("non-enrolled admin denied", func(t *testing.T) {
		_, err := client.GetCourseLog(client.Context(t, admin), &qf.CourseLogRequest{CourseID: course.GetID()})
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Errorf("code = %v, want PermissionDenied: a site admin who is not a teacher of the course", connect.CodeOf(err))
		}
	})

	t.Run("inverted interval rejected", func(t *testing.T) {
		now := time.Now()
		_, err := client.GetCourseLog(client.Context(t, teacher), &qf.CourseLogRequest{
			CourseID: course.GetID(),
			From:     timestamppb.New(now),
			To:       timestamppb.New(now.Add(-time.Hour)),
		})
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want InvalidArgument for From after To", connect.CodeOf(err))
		}
	})

	t.Run("repository filter", func(t *testing.T) {
		got, err := client.GetCourseLog(client.Context(t, teacher), &qf.CourseLogRequest{
			CourseID: course.GetID(), Repository: "student-a", Limit: 100,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.GetEntries()) != 3 {
			t.Fatalf("len(Entries) = %d, want 3", len(got.GetEntries()))
		}
		for _, e := range got.GetEntries() {
			if e.GetRepository() != "student-a" {
				t.Errorf("entry repository = %q, want %q", e.GetRepository(), "student-a")
			}
		}
		// The repository dropdown must not shrink to just the filter applied.
		if len(got.GetRepositories()) != 2 {
			t.Errorf("Repositories = %v, want both repositories regardless of the filter", got.GetRepositories())
		}
	})

	t.Run("level filter", func(t *testing.T) {
		got, err := client.GetCourseLog(client.Context(t, teacher), &qf.CourseLogRequest{
			CourseID: course.GetID(), Level: qf.CourseLogEntry_ERROR, Limit: 100,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.GetEntries()) != 3 {
			t.Fatalf("len(Entries) = %d, want 3", len(got.GetEntries()))
		}
		for _, e := range got.GetEntries() {
			if e.GetLevel() != qf.CourseLogEntry_ERROR {
				t.Errorf("entry level = %v, want ERROR", e.GetLevel())
			}
		}
	})

	t.Run("truncation flag", func(t *testing.T) {
		got, err := client.GetCourseLog(client.Context(t, teacher), &qf.CourseLogRequest{CourseID: course.GetID(), Limit: 2})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.GetEntries()) != 2 {
			t.Fatalf("len(Entries) = %d, want 2", len(got.GetEntries()))
		}
		if !got.GetTruncated() {
			t.Error("Truncated = false, want true: only 2 of 6 matching entries were returned")
		}
	})

	t.Run("empty log", func(t *testing.T) {
		got, err := client.GetCourseLog(client.Context(t, teacher), &qf.CourseLogRequest{CourseID: otherCourse.GetID(), Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.GetEntries()) != 0 || len(got.GetRepositories()) != 0 || got.GetTruncated() {
			t.Errorf("GetCourseLog() = %+v, want an empty result for a course with no log activity", got)
		}
	})
}

// TestGetCourseLogHandlerErrors exercises GetCourseLog's own error paths
// directly, without the access-control interceptor: an unknown course ID
// would never reach checkTeacher's course anyway, since access is granted or
// denied from the caller's actual enrollments, not from whether the
// requested course exists.
func TestGetCourseLogHandlerErrors(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)

	dir := t.TempDir()
	store := seedCourseLog(t, dir, course, 0)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithCourseLogStore(store))

	t.Run("unknown course", func(t *testing.T) {
		_, err := client.GetCourseLog(t.Context(), &qf.CourseLogRequest{CourseID: 1337, Limit: 10})
		if connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("code = %v, want NotFound", connect.CodeOf(err))
		}
	})

	t.Run("storage failure", func(t *testing.T) {
		courseDir := filepath.Join(dir, course.GetScmOrganizationName())
		if err := os.MkdirAll(courseDir, 0o750); err != nil {
			t.Fatal(err)
		}
		date := time.Now().UTC().Format("2006-01-02")
		lines := "not json\n{\"time\":\"" + time.Now().UTC().Format(time.RFC3339) + "\",\"level\":\"INFO\",\"msg\":\"after the bad line\"}\n"
		if err := os.WriteFile(filepath.Join(courseDir, date+".jsonl"), []byte(lines), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := client.GetCourseLog(t.Context(), &qf.CourseLogRequest{CourseID: course.GetID(), Limit: 10})
		if connect.CodeOf(err) != connect.CodeInternal {
			t.Errorf("code = %v, want Internal for a malformed non-final line", connect.CodeOf(err))
		}
	})
}
