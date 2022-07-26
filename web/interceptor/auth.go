package interceptor

import (
	"context"
	"strconv"

	"github.com/quickfeed/quickfeed/web/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidSessionCookie = status.Errorf(codes.Unauthenticated, "request does not contain a valid session cookie.")
	ErrContextMetadata      = status.Errorf(codes.Unauthenticated, "could not obtain metadata from context")
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
func StreamUserVerifier() grpc.StreamServerInterceptor {
	return func(req interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		inCtx := stream.Context()
		context, err := getAuthenticatedContext(inCtx)
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
func UnaryUserVerifier() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, err := getAuthenticatedContext(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// getAuthenticatedContext returns a new context with the user ID attached to it.
// If the context does not contain a valid session cookie, it returns an error.
func getAuthenticatedContext(ctx context.Context) (context.Context, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrContextMetadata
	}
	newMeta, err := userValidation(meta)
	if err != nil {
		return nil, err
	}
	return metadata.NewIncomingContext(ctx, newMeta), nil
}

// userValidation returns modified metadata containing a valid user.
// An error is returned if the user is not authenticated.
func userValidation(meta metadata.MD) (metadata.MD, error) {
	for _, cookie := range meta.Get(auth.Cookie) {
		if user := auth.Get(cookie); user > 0 {
			meta.Set(auth.UserKey, strconv.FormatUint(user, 10))
			return meta, nil
		}
	}
	return nil, ErrInvalidSessionCookie
}
