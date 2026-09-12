package interceptor

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/web/auth"
)

func TestEnrichRequestLogger(t *testing.T) {
	tests := []struct {
		name       string
		courses    map[uint64]qf.Enrollment_UserStatus
		wantCourse bool
	}{
		{name: "trusted enrollment", courses: map[uint64]qf.Enrollment_UserStatus{7: qf.Enrollment_TEACHER}, wantCourse: true},
		{name: "untrusted course", courses: map[uint64]qf.Enrollment_UserStatus{8: qf.Enrollment_TEACHER}},
		{name: "no enrollment", courses: map[uint64]qf.Enrollment_UserStatus{7: qf.Enrollment_NONE}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, nil))
			state := &requestLog{logger: logger}
			ctx := qlog.NewContext(context.WithValue(context.Background(), requestLogKey{}, state), logger)
			ctx = (&auth.Claims{UserID: 12, Courses: test.courses}).Context(ctx)
			ctx = enrichRequestLogger(ctx, &qf.CourseRequest{CourseID: 7})
			qlog.FromContext(ctx).Info("request")
			got := output.String()
			if !strings.Contains(got, `"user_id":12`) {
				t.Fatalf("log output %q lacks trusted user ID", got)
			}
			if strings.Contains(got, `"course_id":7`) != test.wantCourse {
				t.Errorf("course scope in %q = %t, want %t", got, !test.wantCourse, test.wantCourse)
			}
		})
	}
}

func TestEnrichRequestLoggerWithoutClaims(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	ctx := qlog.NewContext(context.Background(), logger)
	ctx = enrichRequestLogger(ctx, &qf.CourseRequest{CourseID: 7})
	qlog.FromContext(ctx).Info("request")
	if got := output.String(); strings.Contains(got, "course_id") || strings.Contains(got, "user_id") {
		t.Errorf("unauthenticated log output contains trusted scope: %q", got)
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
