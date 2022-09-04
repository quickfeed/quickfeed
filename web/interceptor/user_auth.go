package interceptor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"go.uber.org/zap"

	"github.com/quickfeed/quickfeed/web/auth"
)

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
			newCtx := context.WithValue(ctx, auth.ContextKeyUserID, claims.UserID)
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
