package interceptors

import (
	"context"
	"time"

	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ValidateToken(logger *zap.SugaredLogger, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		logger.Debug("TOKEN VALIDATE INTERCEPTOR")
		token, err := GetFromMetadata(ctx, "cookie", tokens.GetAuthCookieName())
		if err != nil {
			logger.Error(err)
			return nil, ErrAccessDenied
		}
		logger.Debugf("Extracted auth token: %s", token) // tmp
		claims, err := tokens.GetClaims(token)
		if err != nil {
			logger.Errorf("Failed to extract claims from JWT: %s", err)
			return nil, ErrAccessDenied
		}
		logger.Debugf("Claims from token: %+v", claims)
		// If user ID is in the update token list, generate and set new JWT
		if tokens.UpdateRequired(claims) {
			logger.Debugf("Token update required for user %d", claims.UserID)
			updatedToken, err := tokens.NewTokenCookie(claims.UserID)
			if err != nil {
				logger.Errorf("Failed to generate new user claims %v", err)
				return nil, ErrAccessDenied
			}
			logger.Debugf("Old token: %s", token)        // tmp
			logger.Debugf("New token: %v", updatedToken) // tmp
			// tokenCookie, err := tokens.NewTokenCookie(ctx, updatedToken)
			// if err != nil {
			// 	logger.Errorf("Failed to make token cookie: %v", err)
			// 	return nil, ErrAccessDenied
			// }
			// //
			if err := tokens.Remove(claims.UserID); err != nil {
				logger.Errorf("Failed to update token list: %s", err)
				return nil, ErrAccessDenied
			}
			if err := setCookie(ctx, updatedToken.String()); err != nil {
				logger.Errorf("Failed to set auth cookie: %s", err)
			}
			token = updatedToken.String()
		}

		ctx, err = setToMetadata(ctx, "token", token)
		if err != nil {
			logger.Error(err)
			return nil, ErrAccessDenied
		}
		logger.Debugf("Token validator interceptor took %v", time.Since(start))
		return handler(ctx, req)
	}
}
