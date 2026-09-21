package interceptor

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/web/auth"
)

// requestLog holds the logger for a single RPC. It is shared, via the request
// context, between the RPCLoggingInterceptor that emits the completion record
// and the ContextLoggingInterceptor that enriches the logger further down the
// chain. Both run in the handler's goroutine, and the enrichment happens-before
// the completion record, so the field needs no synchronization.
type requestLog struct {
	logger    *slog.Logger
	procedure string
}

type requestLogKey struct{}

// RPCLoggingInterceptor records the outcome and duration of every RPC.
type RPCLoggingInterceptor struct {
	logger *slog.Logger
}

func NewRPCLoggingInterceptor(logger *slog.Logger) *RPCLoggingInterceptor {
	return &RPCLoggingInterceptor{logger: logger}
}

func (i *RPCLoggingInterceptor) requestContext(ctx context.Context, procedure string) (context.Context, *requestLog) {
	state := &requestLog{logger: i.logger.With(label.RPCMethod, procedure), procedure: procedure}
	ctx = context.WithValue(ctx, requestLogKey{}, state)
	return qlog.NewContext(ctx, state.logger), state
}

func (*RPCLoggingInterceptor) logCompletion(state *requestLog, started time.Time, err error) {
	duration := time.Since(started)
	if err == nil {
		state.logger.Debug("RPC completed", label.Duration, duration)
		return
	}
	code := connect.CodeOf(err)
	if isAnonymousSessionCheck(state.procedure, code) {
		state.logger.Debug("RPC completed", label.Code, code.String(), label.Duration, duration, label.Error, err)
		return
	}
	state.logger.Error("RPC completed", label.Code, code.String(), label.Duration, duration, label.Error, err)
}

// isAnonymousSessionCheck reports whether the failed RPC is the frontend's
// GetUser call made without a session cookie. The frontend issues it on every
// page load to find out whether someone is logged in, so an unauthenticated
// answer is the expected outcome for a visitor and not a server error.
func isAnonymousSessionCheck(procedure string, code connect.Code) bool {
	return code == connect.CodeUnauthenticated && procedure == qfconnect.QuickFeedServiceGetUserProcedure
}

func (i *RPCLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, state := i.requestContext(ctx, request.Spec().Procedure)
		started := time.Now()
		response, err := next(ctx, request)
		i.logCompletion(state, started, err)
		return response, err
	}
}

func (i *RPCLoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, state := i.requestContext(ctx, conn.Spec().Procedure)
		started := time.Now()
		err := next(ctx, conn)
		i.logCompletion(state, started, err)
		return err
	}
}

func (*RPCLoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// ContextLoggingInterceptor enriches the request logger after authentication
// and access control have accepted the request.
type ContextLoggingInterceptor struct{}

func NewContextLoggingInterceptor() *ContextLoggingInterceptor {
	return &ContextLoggingInterceptor{}
}

// requestLogger returns an enriched logger with the calling user's ID and
// the request's course ID, if the caller's JWT claims indicate that the
// caller is a trusted member of the course. It returns false if the request
// carries no claims, leaving the logger unchanged.
func requestLogger(ctx context.Context, request any) (*slog.Logger, bool) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return nil, false
	}
	logger := qlog.FromContext(ctx).With(label.UserID, claims.UserID)
	courseID := getCourseID(request)
	if status, trusted := claims.Courses[courseID]; courseID > 0 && trusted && status != qf.Enrollment_NONE {
		logger = logger.With(label.CourseID, courseID)
	}
	return logger, true
}

// setCompletionLogger makes logger the one that records this RPC's completion.
func setCompletionLogger(ctx context.Context, logger *slog.Logger) {
	if state, ok := ctx.Value(requestLogKey{}).(*requestLog); ok {
		state.logger = logger
	}
}

// enrichRequestLogger adds the calling user, and the requested course when the
// caller is a trusted member of it, to the request logger. Handlers obtain the
// enriched logger with qlog.FromContext and must not repeat these attributes.
func enrichRequestLogger(ctx context.Context, request any) context.Context {
	logger, ok := requestLogger(ctx, request)
	if !ok {
		return ctx
	}
	setCompletionLogger(ctx, logger)
	return qlog.NewContext(ctx, logger)
}

func (*ContextLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		return next(enrichRequestLogger(ctx, request.Any()), request)
	}
}

// WrapStreamingHandler enriches the handler's logger with the calling user.
// The requested course is only known once the handler receives the request, by
// which point the handler holds its context; the course therefore reaches the
// completion record only, and a handler that needs a course scope derives its
// own with qlog.WithCourse. Both loggers derive from the unenriched ctx, so
// neither repeats the other's attributes.
func (*ContextLoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(enrichRequestLogger(ctx, nil), &checkedConn{
			StreamingHandlerConn: conn,
			check: func(req any) error {
				if logger, ok := requestLogger(ctx, req); ok {
					setCompletionLogger(ctx, logger)
				}
				return nil
			},
		})
	}
}

func (*ContextLoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}
