package interceptor

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

// getAuthenticatedContext returns a new context with the user ID attached to it.
// If the context does not contain a valid session cookie, it returns an error.
func getAuthenticatedContext(ctx context.Context, header http.Header, logger *zap.SugaredLogger, tm *auth.TokenManager) (context.Context, *http.Cookie, error) {
	cookies := header.Get(auth.Cookie)
	token, err := extractToken(cookies)
	if err != nil {
		logger.Errorf("Failed to extract token: %v", err)
		return nil, nil, err
	}
	claims, err := tm.GetClaims(token)
	if err != nil {
		logger.Errorf("Failed to extract claims from JWT: %v", err)
		return nil, nil, ErrInvalidAuthCookie
	}
	newCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(auth.UserKey, strconv.FormatUint(claims.UserID, 10)))
	if tm.UpdateRequired(claims) {
		logger.Debug("Updating token for user ", claims.UserID)
		updatedCookie, err := tm.NewAuthCookie(claims.UserID)
		if err != nil {
			logger.Errorf("Failed to update cookie: %v", err)
			return nil, nil, ErrInvalidAuthCookie
		}
		return newCtx, updatedCookie, nil
	}
	return newCtx, nil, nil
}

// extractToken extracts a JWT authentication token from metadata.
func extractToken(cookieString string) (string, error) {
	cookies := strings.Split(cookieString, ";")
	for _, cookie := range cookies {
		_, cookieValue, ok := strings.Cut(cookie, auth.CookieName+"=")
		if ok {
			return strings.TrimSpace(cookieValue), nil
		}
	}
	return "", ErrInvalidAuthCookie
}
