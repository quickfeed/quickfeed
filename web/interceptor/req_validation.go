package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
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
func Validation(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validate(logger, req); err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(ctx, MaxWait)
		defer cancel()
		// if response has information on remote ID, it will be removed
		resp, err := handler(ctx, req)
		clean(resp)
		return resp, err
	}
}

func validate(logger *zap.Logger, req interface{}) error {
	if v, ok := req.(validator); ok {
		if !v.IsValid() {
			return status.Errorf(codes.InvalidArgument, "invalid payload")
		}
	} else {
		// just logging, but still handling the call
		logger.Sugar().Debugf("message type %T does not implement validator interface", req)
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
