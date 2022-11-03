package interceptor

import (
	"context"

	"github.com/bufbuild/connect-go"
)

// NewTokenAuthClientInterceptor returns a client interceptor that will add the given token in the Authorization header.
func NewTokenAuthClientInterceptor(token string) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if req.Spec().IsClient {
				// Send a token with client requests.
				req.Header().Set(tokenHeader, token)
			}
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
