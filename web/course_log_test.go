package web_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
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
	req.To = qf.TimePosition(time.Now())
	return req
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
		for _, e := range got.GetEntries() {
			if !e.GetCursor().IsValid() {
				t.Errorf("entry %q has cursor %v, want a valid one to resume from", e.GetMessage(), e.GetCursor())
			}
		}
	})

	t.Run("bounded request honors a from cursor", func(t *testing.T) {
		all, err := backlog(t, client, client.Context(t, teacher), bounded(&qf.CourseLogRequest{CourseID: course.GetID(), Limit: 100}))
		if err != nil {
			t.Fatal(err)
		}
		got, err := backlog(t, client, client.Context(t, teacher), bounded(&qf.CourseLogRequest{
			CourseID: course.GetID(), From: qf.CursorPosition(all.GetEntries()[2].GetCursor()), Limit: 100,
		}))
		if err != nil {
			t.Fatal(err)
		}
		if n := len(got.GetEntries()); n != 3 {
			t.Errorf("len(Entries) = %d, want 3, the entries after the third", n)
		}
	})

	// The validation interceptor rejects a cursor that is no position at all
	// before the handler runs; only the store knows where its records end.
	const byInterceptor, byStore = "invalid payload", "invalid course log cursor"
	today := qf.LogDay(time.Now())
	for name, tt := range map[string]struct {
		cursor     *qf.LogCursor
		rejectedBy string
	}{
		"no date":         {&qf.LogCursor{Offset: 1}, byInterceptor},
		"off midnight":    {&qf.LogCursor{Date: timestamppb.New(today.Add(time.Hour))}, byInterceptor},
		"inside a record": {qf.NewLogCursor(today, 1), byStore},
		"beyond the end":  {qf.NewLogCursor(today, 1<<40), byStore},
	} {
		requests := map[string]*qf.CourseLogRequest{
			"from, following": {From: qf.CursorPosition(tt.cursor)},
			"from, bounded":   {From: qf.CursorPosition(tt.cursor), To: qf.TimePosition(time.Now())},
			"to":              {To: qf.CursorPosition(tt.cursor)},
		}
		for kind, req := range requests {
			t.Run(fmt.Sprintf("%s cursor rejected, %s", name, kind), func(t *testing.T) {
				req.CourseID, req.Limit = course.GetID(), 100
				_, err := backlog(t, client, client.Context(t, teacher), req)
				var connectErr *connect.Error
				if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeInvalidArgument || connectErr.Message() != tt.rejectedBy {
					t.Errorf("error = %v, want InvalidArgument %q for cursor %v", err, tt.rejectedBy, tt.cursor)
				}
			})
		}
	}

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
			From:     qf.TimePosition(now),
			To:       qf.TimePosition(now.Add(-time.Hour)),
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
		CourseID: course.GetID(), From: qf.TimePosition(time.Now().Add(time.Hour)), Limit: 100,
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

// TestCourseLogStreamPaging walks the requests the Course Logs page sends, in
// the order a teacher might: a preset showing the newest entries, Load older,
// a picked date showing the oldest, and Load newer until the view catches up.
// The stream follows new entries only once its backlog reaches the present.
func TestCourseLogStreamPaging(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)

	store := seedCourseLog(t, t.TempDir(), course, 0)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithCourseLogStore(store))
	logger := courseLogger(store, course)
	for i := range 5 {
		logger.Info(fmt.Sprintf("record %d", i))
	}
	from := qf.TimePosition(time.Now().Add(-time.Hour))
	ctx := client.Context(t, teacher)

	// open sends req and checks its backlog; it returns the stream, still
	// open if the server kept it so, and the backlog's cursors.
	open := func(t *testing.T, req *qf.CourseLogRequest, want []string, wantTruncated bool) (*connect.ServerStreamForClient[qf.CourseLog], []*qf.LogCursor) {
		t.Helper()
		req.CourseID, req.Limit = course.GetID(), 2
		stream, err := client.CourseLogStream(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = stream.Close() })
		got, err := next(stream)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(messages(got), want) || got.GetTruncated() != wantTruncated {
			t.Fatalf("backlog = %v, truncated %v; want %v, truncated %v", messages(got), got.GetTruncated(), want, wantTruncated)
		}
		cursors := make([]*qf.LogCursor, len(got.GetEntries()))
		for i, e := range got.GetEntries() {
			cursors[i] = e.GetCursor()
		}
		return stream, cursors
	}
	closed := func(t *testing.T, stream *connect.ServerStreamForClient[qf.CourseLog]) {
		t.Helper()
		if stream.Receive() {
			t.Errorf("stream sent %v, want it closed after the backlog", messages(stream.Msg()))
		}
		if err := stream.Err(); err != nil {
			t.Errorf("Err() = %v, want nil: the stream ends cleanly", err)
		}
	}
	follows := func(t *testing.T, stream *connect.ServerStreamForClient[qf.CourseLog], msg string) {
		t.Helper()
		logger.Info(msg)
		live, err := next(stream)
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{msg}; !slices.Equal(messages(live), want) {
			t.Errorf("live = %v, want %v", messages(live), want)
		}
	}

	// A preset: the newest entries, then whatever is logged next.
	stream, cursors := open(t, &qf.CourseLogRequest{From: from, Newest: true}, []string{"record 3", "record 4"}, true)
	follows(t, stream, "record 5")
	oldest := cursors[0]

	t.Run("load older", func(t *testing.T) {
		stream, _ := open(t, &qf.CourseLogRequest{From: from, To: qf.CursorPosition(oldest), Newest: true}, []string{"record 1", "record 2"}, true)
		closed(t, stream)
	})

	t.Run("picked date, then load newer until live", func(t *testing.T) {
		stream, cursors := open(t, &qf.CourseLogRequest{From: from}, []string{"record 0", "record 1"}, true)
		closed(t, stream)
		stream, cursors = open(t, &qf.CourseLogRequest{From: qf.CursorPosition(cursors[1])}, []string{"record 2", "record 3"}, true)
		closed(t, stream)
		stream, _ = open(t, &qf.CourseLogRequest{From: qf.CursorPosition(cursors[1])}, []string{"record 4", "record 5"}, false)
		follows(t, stream, "record 6")
	})

	t.Run("picked range", func(t *testing.T) {
		stream, _ := open(t, &qf.CourseLogRequest{From: from, To: qf.TimePosition(time.Now())}, []string{"record 0", "record 1"}, true)
		closed(t, stream)
	})
}

// logAt writes a record stamped at, whenever it is written. A record is
// stamped before it reaches the course file's lock, so this is what lets a
// test write records in an order their timestamps disagree with, as
// concurrent logging does.
func logAt(t *testing.T, logger *slog.Logger, at time.Time, msg string) {
	t.Helper()
	if err := logger.Handler().Handle(context.Background(), slog.NewRecord(at, slog.LevelInfo, msg, 0)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
}

func messages(log *qf.CourseLog) []string {
	msgs := make([]string, len(log.GetEntries()))
	for i, e := range log.GetEntries() {
		msgs[i] = e.GetMessage()
	}
	return msgs
}

// TestCourseLogStreamResume checks that a client reconnecting from its newest
// entry's cursor receives exactly the entries it missed, including entries
// sharing a timestamp and one written out of timestamp order.
func TestCourseLogStreamResume(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := qtest.MockCourses[0]
	teacher := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, teacher, course)

	// Stamped in the recent past, so a reconnect's window and the query's
	// clock both cover them however the test is scheduled.
	stamp := time.Now().Add(-time.Minute)
	from := qf.TimePosition(stamp.Add(-time.Hour))

	tests := map[string]struct {
		// before is logged before the client connects, live while it is
		// connected, and missed while it is disconnected.
		before, live, missed []record
	}{
		"record written out of timestamp order": {
			live:   []record{{stamp.Add(time.Nanosecond), "stamped later, written first"}},
			missed: []record{{stamp, "stamped earlier, written while disconnected"}},
		},
		"records sharing a timestamp": {
			before: []record{{stamp, "tie 1"}, {stamp, "tie 2"}},
			live:   []record{{stamp, "tie 3"}},
			missed: []record{{stamp, "tie 4"}, {stamp, "tie 5"}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			store := seedCourseLog(t, t.TempDir(), course, 0)
			client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithCourseLogStore(store))
			logger := courseLogger(store, course)
			for _, r := range tt.before {
				logAt(t, logger, r.at, r.msg)
			}

			ctx, disconnect := context.WithCancel(client.Context(t, teacher))
			stream, err := client.CourseLogStream(ctx, &qf.CourseLogRequest{CourseID: course.GetID(), From: from, Limit: 100})
			if err != nil {
				t.Fatal(err)
			}
			first, err := next(stream)
			if err != nil {
				t.Fatal(err)
			}
			seen := messages(first)
			last := lastCursor(first)
			for _, r := range tt.live {
				logAt(t, logger, r.at, r.msg)
				msg, err := next(stream)
				if err != nil {
					t.Fatal(err)
				}
				seen = append(seen, messages(msg)...)
				last = lastCursor(msg)
			}
			if want := append(recordMessages(tt.before), recordMessages(tt.live)...); !slices.Equal(seen, want) {
				t.Fatalf("delivered %v before disconnecting, want %v", seen, want)
			}
			disconnect()
			_ = stream.Close()

			for _, r := range tt.missed {
				logAt(t, logger, r.at, r.msg)
			}
			resumed, err := client.CourseLogStream(client.Context(t, teacher), &qf.CourseLogRequest{
				CourseID: course.GetID(), From: qf.CursorPosition(last), Limit: 100,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = resumed.Close() }()
			got, err := next(resumed)
			if err != nil {
				t.Fatal(err)
			}
			if want := recordMessages(tt.missed); !slices.Equal(messages(got), want) {
				t.Errorf("resumed backlog = %v, want %v: every record missed while disconnected, and none already delivered", messages(got), want)
			}

			// The resumed stream carries on live from where its backlog ends.
			logAt(t, logger, stamp, "after resuming")
			live, err := next(resumed)
			if err != nil {
				t.Fatal(err)
			}
			if want := []string{"after resuming"}; !slices.Equal(messages(live), want) {
				t.Errorf("live after resuming = %v, want %v", messages(live), want)
			}
		})
	}
}

type record struct {
	at  time.Time
	msg string
}

func recordMessages(records []record) []string {
	msgs := make([]string, len(records))
	for i, r := range records {
		msgs[i] = r.msg
	}
	return msgs
}

// lastCursor returns the cursor of log's newest entry, or nil if it has none.
func lastCursor(log *qf.CourseLog) *qf.LogCursor {
	entries := log.GetEntries()
	if len(entries) == 0 {
		return nil
	}
	return entries[len(entries)-1].GetCursor()
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
