package interceptor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/auth/tokens"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var ErrAccessDenied = status.Errorf(codes.Unauthenticated, "access denied")

// getAuthenticatedContext returns a new context with the user ID attached to it.
// If the context does not contain a valid session cookie, it returns an error.
func getAuthenticatedContext(ctx context.Context, logger *zap.SugaredLogger, tm *tokens.TokenManager) (context.Context, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrContextMetadata
	}
	newMeta, err := userValidation(ctx, logger, meta, tm)
	if err != nil {
		return nil, err
	}
	return metadata.NewIncomingContext(ctx, newMeta), nil
}

// userValidation returns modified metadata containing a valid user.
// An error is returned if the user is not authenticated.
func userValidation(ctx context.Context, logger *zap.SugaredLogger, meta metadata.MD, tm *tokens.TokenManager) (metadata.MD, error) {
	token, err := extractToken(meta)
	if err != nil {
		return nil, ErrInvalidAuthCookie
	}
	claims, err := tm.GetClaims(token)
	if err != nil {
		logger.Errorf("Failed to extract claims from JWT: %v", err)
		return nil, ErrInvalidAuthCookie
	}
	if tm.UpdateRequired(claims) {
		logger.Debugf("Updating token for user %d", claims.UserID)
		if err := refreshAuthCookie(ctx, tm, claims); err != nil {
			logger.Errorf("Failed to update authentication token: %v", err)
			return nil, ErrInvalidAuthCookie
		}
	}
	meta.Set(auth.UserKey, strconv.FormatUint(claims.UserID, 10))
	return meta, nil
}

// extractToken extracts a JWT authentication token from metadata.
func extractToken(meta metadata.MD) (string, error) {
	cookies := meta.Get(auth.Cookie)
	for _, cookie := range cookies {
		_, cookievalue, ok := strings.Cut(cookie, auth.CookieName)
		if ok {
			return strings.TrimSpace(cookievalue), nil
		}
	}
	return "", ErrInvalidAuthCookie
}

// refreshAuthCookie sets cookie with an updated JWT authentication token.
func refreshAuthCookie(ctx context.Context, tm *tokens.TokenManager, claims *tokens.Claims) error {
	updatedToken, err := tm.NewAuthCookie(claims.UserID)
	if err != nil {
		return err
	}
	if err := tm.Remove(claims.UserID); err != nil {
		return err
	}
	return setCookie(ctx, updatedToken.String())
}

// setCookie sets a "Set-Cookie" header with JWT token to the outgoing context.
func setCookie(ctx context.Context, cookie string) error {
	if cookie == "" {
		return fmt.Errorf("empty cookie")
	}
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("failed to read metadata")
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "Set-Cookie", cookie)
	if err := grpc.SetHeader(ctx, meta); err != nil {
		return fmt.Errorf("failed to set grpc header: %w", err)
	}
	return nil
}
