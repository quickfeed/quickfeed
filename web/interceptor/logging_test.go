package interceptor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/web/auth"
)

type logNameDB struct {
	database.Database
	userCalls, courseCalls []uint64
	err                    error
}

func (db *logNameDB) GetUser(id uint64) (*qf.User, error) {
	db.userCalls = append(db.userCalls, id)
	return &qf.User{ID: id, Login: "teacher"}, db.err
}

func (db *logNameDB) GetCourse(id uint64) (*qf.Course, error) {
	db.courseCalls = append(db.courseCalls, id)
	return &qf.Course{ID: id, Code: "DAT520"}, db.err
}

func TestEnrichRequestLogger(t *testing.T) {
	tests := []struct {
		name       string
		courses    map[uint64]qf.Enrollment_UserStatus
		wantCourse bool
	}{
		{name: "trusted enrollment", courses: map[uint64]qf.Enrollment_UserStatus{7: qf.Enrollment_TEACHER}, wantCourse: true},
		{name: "untrusted course", courses: map[uint64]qf.Enrollment_UserStatus{8: qf.Enrollment_TEACHER}},
		{name: "no enrollment", courses: map[uint64]qf.Enrollment_UserStatus{7: qf.Enrollment_NONE}},
		{name: "no courses"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &logNameDB{}
			interceptor := NewContextLoggingInterceptor(db)
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, nil))
			state := &requestLog{logger: logger}
			ctx := qlog.NewContext(context.WithValue(context.Background(), requestLogKey{}, state), logger)
			ctx = (&auth.Claims{UserID: 12, Courses: test.courses}).Context(ctx)
			// The request's user is the target, never the calling user.
			ctx = interceptor.enrichRequestLogger(ctx, &qf.Enrollment{CourseID: 7, UserID: 99})
			qlog.FromContext(ctx).Info("request")
			state.logger.Info("completion")
			got := output.String()
			if !strings.Contains(got, `"user_id":12`) {
				t.Fatalf("log output %q lacks trusted user ID", got)
			}
			if strings.Contains(got, `"course_id":7`) != test.wantCourse {
				t.Errorf("course scope in %q = %t, want %t", got, !test.wantCourse, test.wantCourse)
			}
			if strings.Count(got, `"user":"teacher"`) != 2 {
				t.Errorf("user name missing from handler or completion log: %s", got)
			}
			if strings.Contains(got, `"course_code":"DAT520"`) != test.wantCourse {
				t.Errorf("unexpected course name scope: %s", got)
			}
			if len(db.userCalls) != 1 || db.userCalls[0] != 12 {
				t.Errorf("looked up users %v, want only caller 12", db.userCalls)
			}
			if (len(db.courseCalls) == 1) != test.wantCourse {
				t.Errorf("looked up courses %v, want course lookup = %t", db.courseCalls, test.wantCourse)
			}
		})
	}
}

func TestEnrichRequestLoggerWithoutClaims(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	ctx := qlog.NewContext(context.Background(), logger)
	db := &logNameDB{}
	ctx = NewContextLoggingInterceptor(db).enrichRequestLogger(ctx, &qf.CourseRequest{CourseID: 7})
	qlog.FromContext(ctx).Info("request")
	if got := output.String(); strings.Contains(got, "course_id") || strings.Contains(got, "user_id") {
		t.Errorf("unauthenticated log output contains trusted scope: %q", got)
	}
	if len(db.userCalls)+len(db.courseCalls) != 0 {
		t.Error("unauthenticated request looked up names")
	}
}

func TestContextLoggingLookupFailure(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	db := &logNameDB{err: errors.New("database unavailable")}
	interceptor := NewContextLoggingInterceptor(db)
	ctx := (&auth.Claims{UserID: 12, Courses: map[uint64]qf.Enrollment_UserStatus{7: qf.Enrollment_TEACHER}}).Context(qlog.NewContext(t.Context(), logger))
	wantErr := errors.New("handler error")
	next := interceptor.WrapUnary(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		qlog.FromContext(ctx).Info("handler")
		return nil, wantErr
	})
	if _, err := next(ctx, connect.NewRequest(&qf.CourseRequest{CourseID: 7})); !errors.Is(err, wantErr) {
		t.Errorf("handler error = %v, want %v", err, wantErr)
	}
	got := output.String()
	for _, want := range []string{`"user_id":12`, `"course_id":7`} {
		if !strings.Contains(got, want) {
			t.Errorf("log %q lacks %s after lookup failure", got, want)
		}
	}
	if strings.Contains(got, `"user":`) || strings.Contains(got, `"course_code":`) {
		t.Errorf("lookup failure attached a name: %s", got)
	}
}

func TestStreamingContextLogging(t *testing.T) {
	for _, receiveErr := range []error{nil, io.EOF} {
		t.Run(strconv.FormatBool(receiveErr != nil), func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
			db := &logNameDB{}
			contextLogging := NewContextLoggingInterceptor(db)
			rpcLogging := NewRPCLoggingInterceptor(logger)
			handlerErr := errors.New("stream handler error")
			handler := rpcLogging.WrapStreamingHandler(contextLogging.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
				qlog.FromContext(ctx).Info("handler")
				if err := conn.Receive(&qf.CourseRequest{CourseID: 7}); err != nil {
					return err
				}
				return handlerErr
			}))
			ctx := (&auth.Claims{UserID: 12, Courses: map[uint64]qf.Enrollment_UserStatus{7: qf.Enrollment_TEACHER}}).Context(t.Context())
			wantErr := handlerErr
			if receiveErr != nil {
				wantErr = receiveErr
			}
			if err := handler(ctx, &fakeConn{receives: []error{receiveErr}}); !errors.Is(err, wantErr) {
				t.Errorf("stream error = %v, want %v", err, wantErr)
			}
			got := strings.Split(strings.TrimSpace(output.String()), "\n")
			if len(got) != 2 {
				t.Fatalf("log records = %v, want handler and completion", got)
			}
			for _, record := range got {
				if strings.Count(record, `"user":"teacher"`) != 1 || strings.Count(record, `"user_id":12`) != 1 {
					t.Errorf("log lacks unique caller scope: %s", record)
				}
			}
			if strings.Contains(got[0], `"course_`) {
				t.Errorf("stream handler has premature course scope: %s", got[0])
			}
			if strings.Contains(got[1], `"course_code":"DAT520"`) != (receiveErr == nil) {
				t.Errorf("unexpected completion course scope: %s", got[1])
			}
			if len(db.userCalls) != 1 || (len(db.courseCalls) == 1) != (receiveErr == nil) {
				t.Errorf("unexpected lookups: users %v, courses %v", db.userCalls, db.courseCalls)
			}
		})
	}
}

func TestRPCCompletionLogging(t *testing.T) {
	tests := []struct {
		name      string
		procedure string
		err       error
		wantLevel string
		wantCode  string
	}{
		{
			name:      "permission denied is an error",
			procedure: qfconnect.QuickFeedServiceGetCourseProcedure,
			err:       connect.NewError(connect.CodePermissionDenied, context.Canceled),
			wantLevel: "ERROR",
			wantCode:  "permission_denied",
		},
		{
			name:      "unauthenticated GetUser is the anonymous session check",
			procedure: qfconnect.QuickFeedServiceGetUserProcedure,
			err:       connect.NewError(connect.CodeUnauthenticated, errors.New("failed to extract authentication cookie from request header")),
			wantLevel: "DEBUG",
			wantCode:  "unauthenticated",
		},
		{
			name:      "unauthenticated other method is still an error",
			procedure: qfconnect.QuickFeedServiceGetCourseProcedure,
			err:       connect.NewError(connect.CodeUnauthenticated, errors.New("failed to extract authentication cookie from request header")),
			wantLevel: "ERROR",
			wantCode:  "unauthenticated",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
			interceptor := NewRPCLoggingInterceptor(logger)
			state := &requestLog{logger: logger.With(label.RPCMethod, test.procedure), procedure: test.procedure}
			interceptor.logCompletion(state, time.Now(), test.err)

			got := output.String()
			for _, want := range []string{`"level":"` + test.wantLevel + `"`, `"rpc_method":"` + test.procedure + `"`, `"code":"` + test.wantCode + `"`, `"duration":`} {
				if !strings.Contains(got, want) {
					t.Errorf("log output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestRPCSuccessLogging(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	interceptor := NewRPCLoggingInterceptor(logger)
	state := &requestLog{logger: logger.With(label.RPCMethod, "/qf.QuickFeedService/GetCourse")}
	interceptor.logCompletion(state, time.Now(), nil)

	got := output.String()
	if !strings.Contains(got, `"level":"DEBUG"`) || !strings.Contains(got, `"duration":`) {
		t.Errorf("successful RPC log output lacks debug completion fields: %q", got)
	}
}
