package interceptor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bufbuild/connect-go"
	"go.uber.org/zap"

	"github.com/quickfeed/quickfeed/web/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

// UnaryUserVerifier returns a unary server interceptor verifying that the user is authenticated.
// The request's session cookie is verified that it contains a valid JWT claim.
// If a valid claim is found, the interceptor injects the user ID as metadata in the incoming context
// for service methods that come after this interceptor.
// The interceptor also updates the session cookie if needed.
func UnaryUserVerifier(logger *zap.SugaredLogger, tm *auth.TokenManager) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			cookie := request.Header().Get(auth.Cookie)
			claims, err := tm.GetClaims(cookie)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("failed to extract JWT claims from session cookie: %w", err))
			}
			var updatedCookie *http.Cookie
			if tm.UpdateRequired(claims) {
				logger.Debug("Updating cookie for user ", claims.UserID)
				updatedCookie, err = tm.UpdateCookie(claims)
				if err != nil {
					return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("failed to update session cookie: %w", err))
				}
			}
			newCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(auth.UserKey, strconv.FormatUint(claims.UserID, 10)))
			response, err := next(newCtx, request)
			if err != nil {
				return nil, err
			}
			if updatedCookie != nil {
				response.Header().Set(auth.SetCookie, updatedCookie.String())
			}
			return response, nil
		})
	})
}
