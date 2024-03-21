package interceptor

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
)

type ContextLoggingInterceptor struct {
	db     database.Database
	logger *zap.Logger
}

func NewContextLoggingInterceptor(logger *zap.Logger, db database.Database) *ContextLoggingInterceptor {
	return &ContextLoggingInterceptor{logger: logger, db: db}
}

func (*ContextLoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (*ContextLoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

// WrapUnary updates list of users who need a new JWT next time they send a request to the server.
// This method only logs errors to avoid overwriting the gRPC error messages returned by the server.
func (t *ContextLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := request.Spec().Procedure
		method := procedure[strings.LastIndex(procedure, "/")+1:]
		if claims, ok := auth.ClaimsFromContext(ctx); ok {
			user, err := t.db.GetUser(claims.UserID)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("cannot get user for %s: %w", method, err))
			}
			logger := t.logger.With(
				zap.String("method", method),
				zap.String("user", user.GetLogin()),
			)
			for courseID, status := range claims.Courses {
				course, err := t.db.GetCourse(courseID, false)
				if err != nil {
					return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("cannot get course for %s: %w", method, err))
				}
				logger = logger.With(
					zap.String("course", course.GetCode()),
					zap.String("status", qf.Enrollment_UserStatus_name[int32(status)]),
				)
			}
			ctx = context.WithValue(ctx, "logger", logger)
		} else {
			return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("cannot populate context for %s: message type %T does not implement 'userIDs' interface", method, request))
		}
		return next(ctx, request)
	})
}
