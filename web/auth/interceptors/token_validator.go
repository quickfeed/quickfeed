package interceptors

import (
	"context"
	"time"

	"github.com/autograde/quickfeed/web/auth/tokens"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ValidateToken validates the integrity of a JWT in each request. It will also create and set a new JWT
// if the current token is in the update list or about to expire.
func ValidateToken(logger *zap.SugaredLogger, tokens *tokens.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		token, err := GetFromMetadata(ctx, "cookie", tokens.GetAuthCookieName())
		if err != nil {
			logger.Error(err)
			return nil, ErrAccessDenied
		}
		claims, err := tokens.GetClaims(token)
		if err != nil {
			logger.Errorf("Failed to extract claims from JWT: %s", err)
			return nil, ErrAccessDenied
		}
		logger.Debugf("Claims from token: %+v", claims)
		// If the token is about to expire or the user ID
		// is in the update token list, generate and set a new JWT.
		if tokens.UpdateRequired(claims) {
			logger.Debugf("Token update required for user %d", claims.UserID)
			updatedToken, err := tokens.NewTokenCookie(claims.UserID)
			if err != nil {
				logger.Errorf("Failed to generate new user claims %v", err)
				return nil, ErrAccessDenied
			}
			if err := tokens.Remove(claims.UserID); err != nil {
				logger.Errorf("Failed to update token list: %s", err)
				return nil, ErrAccessDenied
			}
			if err := setCookie(ctx, updatedToken.String()); err != nil {
				logger.Errorf("Failed to set auth cookie: %s", err)
			}
			token = updatedToken.Value
		}

		ctx, err = setToMetadata(ctx, "token", token)
		if err != nil {
			logger.Error(err)
			return nil, ErrAccessDenied
		}
		logger.Debugf("Token validator interceptor (%s) took %v", info.FullMethod, time.Since(start))
		return handler(ctx, req)
	}
}
