package interceptor

import (
	"context"
	"errors"
	"time"

	"github.com/bufbuild/connect-go"
	"go.uber.org/zap"
)

// MaxWait is the maximum time a request is allowed to stay open before aborting.
const MaxWait = 2 * time.Minute

// validator should be implemented by request types to validate its content.
type validator interface {
	IsValid() bool
}

// idCleaner should be implemented by response types that have a remote ID that should be removed.
type idCleaner interface {
	RemoveRemoteID()
}

type ValidationInterceptor struct {
	logger *zap.SugaredLogger
}

func NewValidationInterceptor(logger *zap.SugaredLogger) *ValidationInterceptor {
	return &ValidationInterceptor{logger: logger}
}

func (*ValidationInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (*ValidationInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return nil // not supported
	})
}

// TokenRefresher updates list of users who need a new JWT next time they send a request to the server.
// This method only logs errors to avoid overwriting the gRPC error messages returned by the server.

// Validation returns a new unary server interceptor that validates requests
// that implements the validator interface.
// Invalid requests are rejected without logging and before it reaches any
// user-level code and returns an illegal argument to the client.
// Further, the response values are cleaned of any remote IDs.
// In addition, the interceptor also implements a cancellation mechanism.
func (v *ValidationInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if request.Any() != nil {
			if err := validate(v.logger, request.Any()); err != nil {
				// Reject the request if it is invalid.
				return nil, err
			}
		}
		resp, err := next(ctx, request)
		if err != nil {
			// Do not return the message to the client if an error occurs.
			// We log the error and return an empty response.
			v.logger.Errorf("Method '%s' failed: %v", request.Spec().Procedure, err)
			v.logger.Errorf("Request Message: %T: %v", request.Any(), request.Any())
			return nil, err
		}
		clean(resp.Any())
		return resp, err
	})
}

func validate(logger *zap.SugaredLogger, req interface{}) error {
	if v, ok := req.(validator); ok {
		if !v.IsValid() {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload"))
		}
	} else {
		// just logging, but still handling the call
		logger.Debugf("message type %T does not implement validator interface", req)
	}
	return nil
}

func clean(resp interface{}) {
	if resp != nil {
		if v, ok := resp.(idCleaner); ok {
			v.RemoveRemoteID()
		}
	}
}
