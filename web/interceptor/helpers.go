package interceptor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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
		logger.Debug("Updating cookie for user ", claims.UserID)
		updatedCookie, err := tm.UpdateCookie(claims)
		if err != nil {
			logger.Errorf("Failed to update cookie: %v", err)
			return nil, nil, ErrInvalidAuthCookie
		}
		return newCtx, updatedCookie, nil
	}
	return newCtx, nil, nil
}

func has(method string) bool {
	_, ok := accessRolesFor[method]
	return ok
}

func CheckAccessMethods(expectedMethodNames map[string]bool) error {
	missingMethods := []string{}
	superfluousMethods := []string{}
	for method := range expectedMethodNames {
		if !has(method) {
			missingMethods = append(missingMethods, method)
		}
	}
	for method := range accessRolesFor {
		if !expectedMethodNames[method] {
			superfluousMethods = append(superfluousMethods, method)
		}
	}
	if len(missingMethods) > 0 {
		return fmt.Errorf("missing required method(s) in access control table: %v", missingMethods)
	}
	if len(superfluousMethods) > 0 {
		return fmt.Errorf("superfluous method(s) in access control table: %v", superfluousMethods)
	}
	return nil
}
