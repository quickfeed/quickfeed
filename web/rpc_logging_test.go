package web_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestRPCLoggingStreamNames(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	teacher := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "teacher"})
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithLogger(logger))

	stream, err := client.CourseLogStream(client.Context(t, teacher), bounded(&qf.CourseLogRequest{CourseID: course.GetID()}))
	qtest.CheckError(t, err, nil)
	defer func() { _ = stream.Close() }()
	for stream.Receive() { //revive:disable-line:empty-block
	}
	qtest.CheckError(t, stream.Err(), nil)
	var completion string
	for record := range strings.SplitSeq(output.String(), "\n") {
		if strings.Contains(record, `"msg":"RPC completed"`) {
			completion = record
		}
	}
	for _, want := range []string{
		`"rpc_method":` + strconv.Quote(qfconnect.QuickFeedServiceCourseLogStreamProcedure),
		`"user":"teacher"`,
		`"course_id":` + strconv.FormatUint(course.GetID(), 10),
		`"course_code":` + strconv.Quote(course.GetCode()),
	} {
		if !strings.Contains(completion, want) {
			t.Errorf("stream completion %q does not contain %s", completion, want)
		}
	}
}

func TestRPCLoggingCourseScope(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithLogger(logger))
	teacher := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "teacher"})
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)

	if _, err := client.GetAssignments(client.Context(t, teacher), &qf.CourseRequest{CourseID: course.GetID()}); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	for _, want := range []string{
		`"rpc_method":"/qf.QuickFeedService/GetAssignments"`,
		`"user_id":` + strconv.FormatUint(teacher.GetID(), 10),
		`"user":` + strconv.Quote(teacher.GetLogin()),
		`"course_id":` + strconv.FormatUint(course.GetID(), 10),
		`"course_code":` + strconv.Quote(course.GetCode()),
		`"level":"DEBUG"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("authorized RPC log output %q does not contain %q", got, want)
		}
	}

	student := qtest.CreateFakeUser(t, db)
	output.Reset()
	_, err := client.GetAssignments(client.Context(t, student), &qf.CourseRequest{CourseID: course.GetID()})
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("GetAssignments() code = %v, want permission denied", connect.CodeOf(err))
	}
	if got := output.String(); strings.Contains(got, `"course_id"`) || strings.Contains(got, `"course_code"`) {
		t.Errorf("unauthorized RPC created course-scoped log output: %q", got)
	}

	output.Reset()
	_, err = client.GetAssignments(context.Background(), &qf.CourseRequest{CourseID: course.GetID()})
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("GetAssignments() code = %v, want unauthenticated", connect.CodeOf(err))
	}
	if got := output.String(); strings.Contains(got, `"course_id"`) || strings.Contains(got, `"course_code"`) || strings.Contains(got, `"user"`) {
		t.Errorf("unauthenticated RPC created course-scoped log output: %q", got)
	}
}

// TestRPCLoggingNoDuplicateScope guards against a handler repeating an attribute
// that an enclosing scope already carries, either from the logging interceptors
// or from a scope the handler itself derived. slog keeps duplicate keys, so a
// repeated attribute shows up twice in the same record.
func TestRPCLoggingNoDuplicateScope(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithLogger(logger))
	teacher := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "teacher"})
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)
	ctx := client.Context(t, teacher)
	target := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "student"})

	// Cover the handlers that derive their own scope, since those are the ones
	// that can duplicate an attribute. Each call must reach the handler's own
	// logging, not just the interceptor's completion record.
	calls := []struct {
		name string
		call func()
	}{
		{"GetSubmissions", func() {
			_, _ = client.GetSubmissions(ctx, &qf.SubmissionRequest{
				CourseID:  course.GetID(),
				FetchMode: &qf.SubmissionRequest_UserID{UserID: teacher.GetID()},
			})
		}},
		{"GetSubmissionsByCourse", func() {
			_, _ = client.GetSubmissionsByCourse(ctx, &qf.SubmissionRequest{
				CourseID:  course.GetID(),
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_ALL},
			})
		}},
		{"GetGroup", func() {
			_, _ = client.GetGroup(ctx, &qf.GroupRequest{CourseID: course.GetID(), GroupID: 1234})
		}},
		{"UpdateCourse", func() {
			// An unknown SCM organization makes the handler log under its scope.
			_, _ = client.UpdateCourse(ctx, &qf.Course{
				ID: course.GetID(), Name: course.GetName(), Code: course.GetCode(),
				Year: course.GetYear(), Tag: course.GetTag(),
				ScmOrganizationID: 1234, ScmOrganizationName: "unknown-org",
			})
		}},
		{"UpdateGroup", func() {
			_, _ = client.UpdateGroup(ctx, &qf.Group{ID: 1234, CourseID: course.GetID(), Name: "unknown", Users: []*qf.User{teacher}})
		}},
		{"DeleteGroup", func() {
			_, _ = client.DeleteGroup(ctx, &qf.GroupRequest{CourseID: course.GetID(), GroupID: 1234})
		}},
		{"UpdateAssignments", func() {
			_, _ = client.UpdateAssignments(ctx, &qf.CourseRequest{CourseID: course.GetID()})
		}},
		{"UpdateUser", func() {
			_, err := client.UpdateUser(ctx, &qf.User{ID: target.GetID(), IsAdmin: true})
			qtest.CheckError(t, err, nil)
		}},
		{"IsEmptyRepo", func() {
			_, _ = client.IsEmptyRepo(ctx, &qf.RepositoryRequest{CourseID: course.GetID(), GroupID: 1234})
		}},
		{"CourseLogStream", func() {
			// Bounded, so the stream closes rather than tailing; the handler's
			// records are written either way.
			stream, err := client.CourseLogStream(ctx, (&qf.CourseLogRequest{CourseID: course.GetID(), Limit: 1}).WithTo(time.Now()))
			if err != nil {
				return
			}
			for stream.Receive() { //revive:disable-line:empty-block
			}
			_ = stream.Close()
		}},
	}
	// Every attribute that some enclosing scope may already carry.
	scopedKeys := []string{
		label.RPCMethod, label.UserID, label.User, label.CourseID, label.CourseCode,
		label.Organization, label.Repository, label.RepositoryType,
		label.Assignment, label.Group, label.GroupID, label.SubmissionID,
		label.TargetUserID, label.Commit,
	}
	for _, c := range calls {
		t.Run(c.name, func(t *testing.T) {
			output.Reset()
			c.call() // the outcome does not matter; the emitted records do
			handlerRecords := 0
			for record := range strings.SplitSeq(strings.TrimSpace(output.String()), "\n") {
				if record == "" {
					continue
				}
				if !strings.Contains(record, `"msg":"RPC completed"`) {
					handlerRecords++
				}
				for _, key := range scopedKeys {
					if count := strings.Count(record, `"`+key+`":`); count > 1 {
						t.Errorf("log record repeats %q %d times, want at most once: %s", key, count, record)
					}
				}
			}
			// Without a record from the handler itself, this call would check nothing.
			if handlerRecords == 0 {
				t.Errorf("no log record from the handler; the call no longer exercises the handler's own scope")
			}
		})
	}
}

func TestRPCLoggingEnrollmentNames(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	teacher := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "teacher", ScmRemoteID: 1})
	student := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "student", ScmRemoteID: 2})
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)
	qtest.EnrollStudent(t, db, student, course)
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := web.NewMockClient(t, db, scm.WithMockOptions(scm.WithMockCourses(), scm.WithMockOrgs("teacher", "student")), web.WithInterceptors(), web.WithLogger(logger))
	ctx := client.Context(t, teacher)
	_, err := client.UpdateEnrollments(ctx, &qf.Enrollments{Enrollments: []*qf.Enrollment{{
		CourseID: course.GetID(), UserID: student.GetID(), Status: qf.Enrollment_TEACHER,
	}}})
	qtest.CheckError(t, err, nil)
	for record := range strings.SplitSeq(strings.TrimSpace(output.String()), "\n") {
		for _, want := range []string{`"user":"teacher"`, `"course_code":` + strconv.Quote(course.GetCode())} {
			if strings.Count(record, want) != 1 {
				t.Errorf("log record %q must contain %s exactly once", record, want)
			}
		}
	}
	if !strings.Contains(output.String(), `"target_user":"student"`) {
		t.Errorf("enrollment log lacks the target user: %s", output.String())
	}
	output.Reset()
	_, err = client.UpdateEnrollments(ctx, &qf.Enrollments{Enrollments: []*qf.Enrollment{{
		CourseID: course.GetID(), UserID: teacher.GetID(), Status: qf.Enrollment_STUDENT,
	}}})
	qtest.CheckError(t, err, connect.NewError(connect.CodePermissionDenied, errors.New("course creator cannot be demoted")))
	for record := range strings.SplitSeq(strings.TrimSpace(output.String()), "\n") {
		if strings.Count(record, `"user":"teacher"`) != 1 {
			t.Errorf("rejected demotion lacks unique caller name: %s", record)
		}
	}
}
