package web_test

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

// TestCourseLogSinkNotOptedIn guards that an ordinary authenticated,
// authorized RPC that never calls qlog.WithCourse writes nothing to the
// course log store: the course_id the interceptors attach on their own must
// not, by itself, opt a scope into a course's log.
func TestCourseLogSinkNotOptedIn(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	dir := t.TempDir()
	store, err := courselog.NewStore(dir, qtest.Logger(t))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()
	sinkLogger := qlog.WithSink(qtest.Logger(t), courselog.NewHandler(store))

	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithLogger(sinkLogger))
	teacher := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)

	if _, err := client.GetAssignments(client.Context(t, teacher), &qf.CourseRequest{CourseID: course.GetID()}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("course log directory has entries %v, want none: GetAssignments does not opt into the course log", entries)
	}
}
