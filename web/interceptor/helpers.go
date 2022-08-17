package interceptor

import (
	"context"
	"fmt"
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
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Error("Failed to extract metadata")
		return nil, ErrContextMetadata
	}
	if tm.UpdateRequired(claims) {
		logger.Debug("Updating token for user ", claims.UserID)
		updatedToken, err := tm.NewAuthCookie(claims.UserID)
		if err != nil {
			logger.Errorf("Failed to update cookie: %v", err)
			return nil, ErrInvalidAuthCookie
		}
		if err := tm.Remove(claims.UserID); err != nil {
			logger.Error(err)
			return nil, ErrInvalidAuthCookie
		}
		if err := grpc.SendHeader(ctx, metadata.Pairs(auth.SetCookie, updatedToken.String())); err != nil {
			logger.Errorf("Failed to set grpc header: %v", err)
			return nil, ErrInvalidAuthCookie
		}
		meta = metadata.New(map[string]string{auth.Cookie: AuthTokenString(updatedToken.Value)})
	}
	meta.Set(auth.UserKey, strconv.FormatUint(claims.UserID, 10))
	return metadata.NewIncomingContext(ctx, meta), nil
}

func has(method string) bool {
	_, ok := access[method]
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
	for method := range access {
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

// AuthTokenString returns a string with JWT with correct format ("auth=JWT").
func AuthTokenString(token string) string {
	return auth.CookieName + "=" + token
}
