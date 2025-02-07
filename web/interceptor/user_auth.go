package interceptor

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	"github.com/quickfeed/quickfeed/web/auth"
)

type UserInterceptor struct {
	tm     *auth.TokenManager
	logger *zap.SugaredLogger
}

func NewUserInterceptor(logger *zap.SugaredLogger, tm *auth.TokenManager) *UserInterceptor {
	return &UserInterceptor{
		tm:     tm,
		logger: logger,
	}
}

func (u *UserInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		claims, updatedCookie, err := u.processHeader(conn.RequestHeader())
		if err != nil {
			return err
		}
		if updatedCookie != nil {
			conn.ResponseHeader().Set(auth.SetCookie, updatedCookie.String())
		}
		return next(claims.Context(ctx), conn)
	})
}

func (*UserInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

// WrapUnary returns a unary server interceptor verifying that the user is authenticated.
// The request's session cookie is verified that it contains a valid JWT claim.
// If a valid claim is found, the interceptor injects the user ID as metadata in the incoming context
// for service methods that come after this interceptor.
// The interceptor also updates the session cookie if needed.
func (u *UserInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		claims, updatedCookie, err := u.processHeader(request.Header())
		if err != nil {
			return nil, err
		}
		response, err := next(claims.Context(ctx), request)
		if err != nil {
			return nil, err
		}
		if updatedCookie != nil {
			response.Header().Set(auth.SetCookie, updatedCookie.String())
		}
		return response, nil
	})
}

// processHeader returns claims extracted from the given http.Header's cookie
// and an updated cookie if needed. An error is returned if the cookie is invalid
// or could not be updated.
func (u *UserInterceptor) processHeader(header http.Header) (*auth.Claims, *http.Cookie, error) {
	cookie := header.Get(auth.Cookie)
	claims, err := u.tm.GetClaims(cookie)
	if err != nil {
		return nil, nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("failed to extract JWT claims from session cookie: %w", err))
	}
	updatedCookie, err := u.tm.UpdateCookie(claims)
	if err != nil {
		return claims, nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("failed to update session cookie: %w", err))
	}
	if updatedCookie == nil {
		return claims, nil, nil
	}
	return claims, updatedCookie, nil
}
