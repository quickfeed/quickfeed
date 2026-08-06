package web_test

import (
	"bytes"
	"context"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestRPCLoggingCourseScope(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors(), web.WithLogger(logger))
	teacher := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)

	if _, err := client.GetAssignments(client.Context(t, teacher), &qf.CourseRequest{CourseID: course.GetID()}); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	for _, want := range []string{`"rpc_method":"/qf.QuickFeedService/GetAssignments"`, `"course_id":` + strconv.FormatUint(course.GetID(), 10), `"level":"DEBUG"`} {
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
	if got := output.String(); strings.Contains(got, `"course_id"`) {
		t.Errorf("unauthorized RPC created course-scoped log output: %q", got)
	}

	output.Reset()
	_, err = client.GetAssignments(context.Background(), &qf.CourseRequest{CourseID: course.GetID()})
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("GetAssignments() code = %v, want unauthenticated", connect.CodeOf(err))
	}
	if got := output.String(); strings.Contains(got, `"course_id"`) {
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
	teacher := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)
	ctx := client.Context(t, teacher)

	// Cover the handlers that derive their own scope, since those are the ones
	// that can duplicate an attribute. Each call must reach the handler's own
	// logging, not just the interceptor's completion record.
	calls := []struct {
		name string
		call func()
	}{
		{"GetSubmissions", func() {
			client.GetSubmissions(ctx, &qf.SubmissionRequest{
				CourseID:  course.GetID(),
				FetchMode: &qf.SubmissionRequest_UserID{UserID: teacher.GetID()},
			})
		}},
		{"GetSubmissionsByCourse", func() {
			client.GetSubmissionsByCourse(ctx, &qf.SubmissionRequest{
				CourseID:  course.GetID(),
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_ALL},
			})
		}},
		{"GetGroup", func() {
			client.GetGroup(ctx, &qf.GroupRequest{CourseID: course.GetID(), GroupID: 1234})
		}},
		{"UpdateCourse", func() {
			// An unknown SCM organization makes the handler log under its scope.
			client.UpdateCourse(ctx, &qf.Course{
				ID: course.GetID(), Name: course.GetName(), Code: course.GetCode(),
				Year: course.GetYear(), Tag: course.GetTag(),
				ScmOrganizationID: 1234, ScmOrganizationName: "unknown-org",
			})
		}},
		{"UpdateGroup", func() {
			client.UpdateGroup(ctx, &qf.Group{ID: 1234, CourseID: course.GetID(), Name: "unknown", Users: []*qf.User{teacher}})
		}},
		{"DeleteGroup", func() {
			client.DeleteGroup(ctx, &qf.GroupRequest{CourseID: course.GetID(), GroupID: 1234})
		}},
		{"UpdateAssignments", func() {
			client.UpdateAssignments(ctx, &qf.CourseRequest{CourseID: course.GetID()})
		}},
		{"IsEmptyRepo", func() {
			client.IsEmptyRepo(ctx, &qf.RepositoryRequest{CourseID: course.GetID(), GroupID: 1234})
		}},
	}
	// Every attribute that some enclosing scope may already carry.
	scopedKeys := []string{
		label.RPCMethod, label.UserID, label.CourseID, label.CourseCode,
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
