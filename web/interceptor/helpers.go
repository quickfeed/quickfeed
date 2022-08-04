package interceptor

import (
	"context"
	"strconv"
	"strings"

	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/auth/tokens"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// getAuthenticatedContext returns a new context with the user ID attached to it.
// If the context does not contain a valid session cookie, it returns an error.
func getAuthenticatedContext(ctx context.Context, logger *zap.SugaredLogger, tm *tokens.TokenManager) (context.Context, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Error("Failed to extract metadata")
		return nil, ErrContextMetadata
	}
	token, err := extractToken(meta)
	if err != nil {
		logger.Errorf("Failed to extract token: %v", err)
		return nil, ErrInvalidAuthCookie
	}
	claims, err := tm.GetClaims(token)
	if err != nil {
		logger.Errorf("Failed to extract claims from JWT: %v", err)
		return nil, ErrInvalidAuthCookie
	}
	if tm.UpdateRequired(claims) {
		logger.Debug("Updating token for user ", claims.UserID)
		updatedToken, err := tm.NewAuthCookie(claims.UserID)
		if err != nil {
			logger.Errorf("Failed to update cookie: %v", err)
		}
		if err := grpc.SendHeader(ctx, metadata.Pairs(tokens.SetCookie, updatedToken.String())); err != nil {
			logger.Errorf("Failed to set grpc header: %v", err)
			return nil, ErrInvalidAuthCookie
		}
	}
	meta.Set(auth.UserKey, strconv.FormatUint(claims.UserID, 10))
	return metadata.NewIncomingContext(ctx, meta), nil
}

// extractToken extracts a JWT authentication token from metadata.
func extractToken(meta metadata.MD) (string, error) {
	cookies := meta.Get(auth.Cookie)
	for _, cookie := range cookies {
		_, cookieValue, ok := strings.Cut(cookie, tokens.AuthCookieName+"=")
		if ok {
			return strings.TrimSpace(cookieValue), nil
		}
	}
	return "", ErrInvalidAuthCookie
}
