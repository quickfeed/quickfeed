package interceptor

import (
	"context"
	"reflect"
	"time"

	"github.com/bufbuild/connect-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// Validation returns a new unary server interceptor that validates requests
// that implements the validator interface.
// Invalid requests are rejected without logging and before it reaches any
// user-level code and returns an illegal argument to the client.
// Further, the response values are cleaned of any remote IDs.
// In addition, the interceptor also implements a cancellation mechanism.
func Validation(logger *zap.SugaredLogger) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			if request.Any() != nil {
				validate(logger, request.Any())
			}
			resp, err := next(ctx, request)
			if err != nil {
				// Do not return the message to the client if an error occurs.
				// We log the error and return an empty response.
				logger.Errorf("Method '%s' failed: %v", request.Spec().Procedure, err)
				logger.Errorf("Request Message: %v", request.Any())
				logger.Errorf("Request Headers: %v", request.Header())
				logger.Errorf("Request Spec: %v", request.Spec())
				return nil, err
			}
			clean(resp.Any())
			return resp, err
		})
	})
}

func validate(logger *zap.SugaredLogger, req interface{}) error {
	if v, ok := req.(validator); ok {
		if !v.IsValid() {
			return status.Errorf(codes.InvalidArgument, "invalid payload")
		}
	} else {
		// just logging, but still handling the call
		logger.Debugf("message type '%s' does not implement validator interface",
			reflect.TypeOf(req).String())
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
