package interceptor

import (
	"context"

	"github.com/quickfeed/quickfeed/web/auth/tokens"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidAuthCookie = status.Errorf(codes.Unauthenticated, "request does not contain a valid authentication cookie.")
	ErrContextMetadata   = status.Errorf(codes.Unauthenticated, "could not obtain metadata from context")
)

// StreamWrapper wraps a stream with a context.
// This is required because we cannot modify the context of a stream directly.
type StreamWrapper struct {
	grpc.ServerStream
	context context.Context
}

func (s *StreamWrapper) Context() context.Context {
	return s.context
}

// StreamUserVerifier returns a gRPC stream server interceptor that verifies
// the user is authenticated.
func StreamUserVerifier(logger *zap.SugaredLogger, tm *tokens.TokenManager) grpc.StreamServerInterceptor {
	return func(req interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		inCtx := stream.Context()
		context, err := getAuthenticatedContext(inCtx, logger, tm)
		if err != nil {
			return err
		}
		// Wrapping the stream in a StreamWrapper to allow us to use a modified context.
		// This way we can use the context to get the user ID in the handler.
		wrappedStream := &StreamWrapper{ServerStream: stream, context: context}
		return handler(req, wrappedStream)
	}
}

// UnaryUserVerifier returns a gRPC unary server interceptor that verifies
// the user is authenticated. This is done by checking the context metadata
// and verifying the session cookie. The context is modified to contain the
// the user ID if the session cookie is valid.
func UnaryUserVerifier(logger *zap.SugaredLogger, tm *tokens.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, err := getAuthenticatedContext(ctx, logger, tm)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}
