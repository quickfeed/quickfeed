package interceptor

import (
	"context"

	"github.com/bufbuild/connect-go"
	"go.uber.org/zap"

	"github.com/quickfeed/quickfeed/web/auth"
	"google.golang.org/grpc"
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

// TODO(jostein): Re-implement this for streaming.
// StreamUserVerifier returns a gRPC stream server interceptor that verifies
// the user is authenticated.
// func StreamUserVerifier(logger *zap.SugaredLogger, tm *auth.TokenManager) grpc.StreamServerInterceptor {
// 	return func(req interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
// 		inCtx := stream.Context()
// 		context, err := getAuthenticatedContext(inCtx, logger, tm)
// 		if err != nil {
// 			return err
// 		}
// 		// Wrapping the stream in a StreamWrapper to allow us to use a modified context.
// 		// This way we can use the context to get the user ID in the handler.
// 		wrappedStream := &StreamWrapper{ServerStream: stream, context: context}
// 		return handler(req, wrappedStream)
// 	}
// }

// UnaryUserVerifier returns a gRPC unary server interceptor that verifies
// the user is authenticated. This is done by checking the request headers
// and verifying the session cookie. The context is modified to contain the
// the user ID if the session cookie is valid.
func UnaryUserVerifier(logger *zap.SugaredLogger, tm *auth.TokenManager) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			newCtx, cookie, err := getAuthenticatedContext(ctx, request.Header(), logger, tm)
			if err != nil {
				return nil, err
			}
			response, err := next(newCtx, request)
			if err != nil {
				return nil, err
			}
			if cookie != nil {
				response.Header().Set(auth.SetCookie, cookie.String())
			}
			return response, err
		})
	})
}
