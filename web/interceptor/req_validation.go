package interceptor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
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

type ValidationInterceptor struct{}

func NewValidationInterceptor() *ValidationInterceptor {
	return &ValidationInterceptor{}
}

func (*ValidationInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (*ValidationInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

// WrapUnary returns a new unary server interceptor that validates requests
// that implements the validator interface.
// Invalid requests are rejected without logging and before it reaches any
// user-level code and returns an illegal argument to the client.
// Further, the response values are cleaned of any remote IDs.
// In addition, the interceptor also implements a cancellation mechanism.
func (*ValidationInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if request.Any() != nil {
			if err := validate(ctx, request.Any()); err != nil {
				// Reject the request if it is invalid.
				return nil, err
			}
		}
		resp, err := next(ctx, request)
		if err != nil {
			return nil, err
		}
		clean(resp.Any())
		return resp, err
	})
}

func validate(ctx context.Context, req any) error {
	if v, ok := req.(validator); ok {
		if !v.IsValid() {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload"))
		}
	} else {
		// just logging, but still handling the call
		qlog.FromContext(ctx).Debug("message does not implement validator interface", "type", fmt.Sprintf("%T", req))
	}
	return nil
}

func clean(resp any) {
	if resp != nil {
		if v, ok := resp.(idCleaner); ok {
			v.RemoveRemoteID()
		}
	}
}
