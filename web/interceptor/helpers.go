package interceptor

import (
	"context"
	"strconv"

	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// getAuthenticatedContext returns a new context with the user ID attached to it.
// If the context does not contain a valid session cookie, it returns an error.
func getAuthenticatedContext(ctx context.Context, logger *zap.SugaredLogger, tm *auth.TokenManager) (context.Context, error) {
	claims, err := tm.GetClaims(ctx)
	if err != nil {
		logger.Errorf("Failed to extract claims from JWT: %v", err)
		return nil, ErrInvalidAuthCookie
	}
	if tm.UpdateRequired(claims) {
		logger.Debug("Updating token for user ", claims.UserID)
		updatedToken, err := tm.NewAuthCookie(claims.UserID)
		if err != nil {
			logger.Errorf("Failed to update cookie: %v", err)
			return nil, ErrInvalidAuthCookie
		}
		if err := grpc.SendHeader(ctx, metadata.Pairs(auth.SetCookie, updatedToken.String())); err != nil {
			logger.Errorf("Failed to set grpc header: %v", err)
			return nil, ErrInvalidAuthCookie
		}
	}
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Error("Failed to extract metadata")
		return nil, ErrContextMetadata
	}
	meta.Set(auth.UserKey, strconv.FormatUint(claims.UserID, 10))
	return metadata.NewIncomingContext(ctx, meta), nil
}

// GetAccessTable returns the current access table for tests.
func GetAccessTable() map[string]roles {
	return access
}
