package interceptors

import (
	"context"

	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ValidateToken(logger *zap.SugaredLogger, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
			updatedClaims, err := tokens.NewClaims(claims.UserID)
			if err != nil {
				logger.Errorf("Failed to generate new user claims %v", err)
				return nil, ErrAccessDenied
			}
			updatedToken := tokens.NewToken(updatedClaims)
			logger.Debugf("Old token: %s", token)        // tmp
			logger.Debugf("New token: %v", updatedToken) // tmp
			tokenCookie, err := tokens.NewTokenCookie(ctx, updatedToken)
			if err != nil {
				logger.Errorf("Failed to make token cookie: %v", err)
				return nil, ErrAccessDenied
			}
			//
			if err := tokens.Remove(claims.UserID); err != nil {
				logger.Errorf("Failed to update token list: %s", err)
				return nil, ErrAccessDenied
			}
			if err := setCookie(ctx, tokenCookie.String()); err != nil {
				logger.Errorf("Failed to set auth cookie: %s", err)
			}
		}

		ctx, err = setToMetadata(ctx, "token", token)
		if err != nil {
			logger.Error(err)
			return nil, ErrAccessDenied
		}
		return handler(ctx, req)
	}
}
