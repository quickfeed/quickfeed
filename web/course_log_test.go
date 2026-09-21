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

// courseLogger returns a logger whose records the sink routes to course's log,
// the same public path production code takes.
func courseLogger(store *courselog.Store, course *qf.Course) *slog.Logger {
	return slog.New(courselog.NewHandler(store)).With(qlog.CourseAttrs(course)...)
}

// seedCourseLog writes n records for course to a fresh store, alternating
// between two repositories and INFO/ERROR levels.
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
	logger := courseLogger(store, course)
	for i := range n {
		repo, level := "student-a", slog.LevelInfo
		if i%2 == 1 {
			repo, level = "student-b", slog.LevelError
		}
		logger.Log(context.Background(), level, "test record", label.Repository, repo)
	}
	return store
}

// bounded returns req with To set to now, which makes CourseLogStream answer
// with the backlog alone and close, rather than staying open to tail.
func bounded(req *qf.CourseLogRequest) *qf.CourseLogRequest {
	return req.WithTo(time.Now())
}

// backlog opens the stream for req and returns its first message, which holds
// the stored entries matching req.
func backlog(t *testing.T, client *web.MockClient, ctx context.Context, req *qf.CourseLogRequest) (*qf.CourseLog, error) {
	t.Helper()
	stream, err := client.CourseLogStream(ctx, req)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = stream.Close() })
	return next(stream)
}

// next returns the stream's next message, or the error that ended it. A
// streaming call reports rejection when the client reads, not when it opens.
func next(stream *connect.ServerStreamForClient[qf.CourseLog]) (*qf.CourseLog, error) {
	if !stream.Receive() {
		return nil, stream.Err()
	}
	return stream.Msg(), nil
}

func TestCourseLogStream(t *testing.T) {
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
		got, err := backlog(t, client, client.Context(t, teacher), bounded(&qf.CourseLogRequest{CourseID: course.GetID(), Limit: 100}))
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

	t.Run("bounded request ends the stream", func(t *testing.T) {
		stream, err := client.CourseLogStream(client.Context(t, teacher), bounded(&qf.CourseLogRequest{CourseID: course.GetID(), Limit: 100}))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = stream.Close() }()
		if _, err := next(stream); err != nil {
			t.Fatal(err)
		}
		if stream.Receive() {
			t.Error("stream sent a second message, want it closed after the backlog when To is set")
		}
		if err := stream.Err(); err != nil {
			t.Errorf("Err() = %v, want nil: a bounded stream ends cleanly", err)
		}
	})

	t.Run("student denied", func(t *testing.T) {
		_, err := backlog(t, client, client.Context(t, student), &qf.CourseLogRequest{CourseID: course.GetID()})
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Errorf("code = %v, want PermissionDenied", connect.CodeOf(err))
		}
	})

	t.Run("non-enrolled admin denied", func(t *testing.T) {
		_, err := backlog(t, client, client.Context(t, admin), &qf.CourseLogRequest{CourseID: course.GetID()})
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Errorf("code = %v, want PermissionDenied: a site admin who is not a teacher of the course", connect.CodeOf(err))
		}
	})

	t.Run("inverted interval rejected", func(t *testing.T) {
		now := time.Now()
		_, err := backlog(t, client, client.Context(t, teacher), &qf.CourseLogRequest{
			CourseID: course.GetID(),
			From:     timestamppb.New(now),
			To:       timestamppb.New(now.Add(-time.Hour)),
		})
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want InvalidArgument for From after To", connect.CodeOf(err))
		}
	})

	t.Run("repository filter", func(t *testing.T) {
		got, err := backlog(t, client, client.Context(t, teacher), bounded(&qf.CourseLogRequest{
			CourseID: course.GetID(), Repository: "student-a", Limit: 100,
		}))
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
		got, err := backlog(t, client, client.Context(t, teacher), bounded(&qf.CourseLogRequest{
			CourseID: course.GetID(), Level: qf.CourseLogEntry_ERROR, Limit: 100,
		}))
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
		got, err := backlog(t, client, client.Context(t, teacher), bounded(&qf.CourseLogRequest{CourseID: course.GetID(), Limit: 2}))
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
		got, err := backlog(t, client, client.Context(t, teacher), bounded(&qf.CourseLogRequest{CourseID: otherCourse.GetID(), Limit: 100}))
		if err != nil {
			t.Fatal(err)
		}
		if len(got.GetEntries()) != 0 || len(got.GetRepositories()) != 0 || got.GetTruncated() {
			t.Errorf("CourseLogStream() = %+v, want an empty result for a course with no log activity", got)
		}
	})
}

// TestCourseLogStreamTail covers the live half: a request without To keeps the
// stream open and delivers entries as they are logged, which is the whole
// point of streaming the log rather than polling it.
func TestCourseLogStreamTail(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)

	store := seedCourseLog(t, t.TempDir(), course, 2)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithCourseLogStore(store))
	logger := courseLogger(store, course)

	stream, err := client.CourseLogStream(client.Context(t, teacher), &qf.CourseLogRequest{CourseID: course.GetID(), Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stream.Close() }()

	first, err := next(stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.GetEntries()) != 2 {
		t.Fatalf("len(Entries) = %d, want the 2 seeded entries in the backlog", len(first.GetEntries()))
	}

	logger.Warn("pushed while watching", label.Repository, "student-c")
	live, err := next(stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(live.GetEntries()) != 1 {
		t.Fatalf("len(Entries) = %d, want 1 entry per live message", len(live.GetEntries()))
	}
	if got := live.GetEntries()[0].GetMessage(); got != "pushed while watching" {
		t.Errorf("Message = %q, want %q", got, "pushed while watching")
	}
	// A repository first seen live must be offered as a filter option.
	if repos := live.GetRepositories(); len(repos) != 1 || repos[0] != "student-c" {
		t.Errorf("Repositories = %v, want [student-c]: a newly seen repository is named once", repos)
	}
	// ...and only once.
	logger.Warn("second push", label.Repository, "student-c")
	again, err := next(stream)
	if err != nil {
		t.Fatal(err)
	}
	if repos := again.GetRepositories(); len(repos) != 0 {
		t.Errorf("Repositories = %v, want none: the client already has student-c", repos)
	}
}

// TestCourseLogStreamTailFilters checks that the live half applies the same
// repository and level filters as the backlog, so a filtered view does not
// fill up with entries the teacher asked to hide.
func TestCourseLogStreamTailFilters(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)

	store := seedCourseLog(t, t.TempDir(), course, 0)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithCourseLogStore(store))
	logger := courseLogger(store, course)

	stream, err := client.CourseLogStream(client.Context(t, teacher), &qf.CourseLogRequest{
		CourseID: course.GetID(), Repository: "student-a", Level: qf.CourseLogEntry_WARN, Limit: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stream.Close() }()
	if _, err := next(stream); err != nil {
		t.Fatal(err)
	}

	logger.Info("below the level filter", label.Repository, "student-a")
	logger.Warn("wrong repository", label.Repository, "student-b")
	logger.Warn("kept", label.Repository, "student-a")

	live, err := next(stream)
	if err != nil {
		t.Fatal(err)
	}
	if got := live.GetEntries()[0].GetMessage(); got != "kept" {
		t.Errorf("Message = %q, want %q: the filtered-out entries must not be sent", got, "kept")
	}
}

// From may be set without To, and the tail has to honour it: a window that
// begins after the subscription did leaves a stretch the backlog query
// excluded, which the live path must not hand over instead.
func TestCourseLogStreamTailAppliesFrom(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)

	store := seedCourseLog(t, t.TempDir(), course, 0)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithCourseLogStore(store))
	logger := courseLogger(store, course)

	ctx, cancel := context.WithTimeout(client.Context(t, teacher), 500*time.Millisecond)
	defer cancel()
	stream, err := client.CourseLogStream(ctx, &qf.CourseLogRequest{
		CourseID: course.GetID(), From: timestamppb.New(time.Now().Add(time.Hour)), Limit: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stream.Close() }()
	first, err := next(stream)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(first.GetEntries()); n != 0 {
		t.Fatalf("backlog holds %d entries, want 0 for a window that begins in an hour", n)
	}

	logger.Warn("logged before the window begins", label.Repository, "student-a")

	// The stream has nothing to say until the deadline: the record is older
	// than the window asked for, however new it is to the subscription.
	if msg, err := next(stream); err == nil {
		t.Fatalf("stream sent %q, want an entry before From withheld", msg.GetEntries()[0].GetMessage())
	} else if connect.CodeOf(err) != connect.CodeDeadlineExceeded {
		t.Errorf("code = %v, want DeadlineExceeded", connect.CodeOf(err))
	}
}

// TestCourseLogStreamHandlerErrors exercises the handler's own error paths
// directly, without the access-control interceptor: an unknown course ID
// would never reach checkTeacher's course anyway, since access is granted or
// denied from the caller's actual enrollments, not from whether the
// requested course exists.
func TestCourseLogStreamHandlerErrors(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)

	dir := t.TempDir()
	store := seedCourseLog(t, dir, course, 0)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithCourseLogStore(store))

	t.Run("unknown course", func(t *testing.T) {
		_, err := backlog(t, client, t.Context(), &qf.CourseLogRequest{CourseID: 1337, Limit: 10})
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

		_, err := backlog(t, client, t.Context(), &qf.CourseLogRequest{CourseID: course.GetID(), Limit: 10})
		if connect.CodeOf(err) != connect.CodeInternal {
			t.Errorf("code = %v, want Internal for a malformed non-final line", connect.CodeOf(err))
		}
	})
}
