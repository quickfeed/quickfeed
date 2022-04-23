package interceptors

import (
	"context"
	"strings"

	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ValidateToken(logger *zap.Logger, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug("TOKEN VALIDATE INTERCEPTOR")
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Sugar().Errorf("Token validation failed: missing metadata")
			return nil, status.Errorf(codes.Unauthenticated, "access denied")
		}
		logger.Sugar().Debugf("Validate: Request %s has metadata: %+v", info.FullMethod, meta) // tmp
		for _, c := range meta.Get("cookie") {
			fields := strings.Fields(c)
			for _, field := range fields {
				logger.Sugar().Debugf("metadata field: %s", field) // tmp
				if strings.Contains(field, "auth") {
					token := strings.Split(field, "=")[1]
					logger.Sugar().Debugf("extracted token: %s", token) // tmp
					claims, err := tokens.GetClaims(token)
					if err != nil {
						logger.Sugar().Errorf("Failed to extract claims from JWT: %s", err)
						return nil, status.Errorf(codes.Unauthenticated, "access denied")
					}
					// If user ID is in the update token list, generate and set new JWT
					if tokens.UpdateRequired(claims) {
						updatedClaims, err := tokens.NewClaims(claims.UserID)
						if err != nil {
							logger.Sugar().Errorf("Token update failed: cannot generate new claims %v", err)
							return nil, status.Errorf(codes.Unauthenticated, "access denied")
						}
						updatedToken := tokens.NewToken(updatedClaims)
						tokenCookie, err := tokens.NewTokenCookie(ctx, updatedToken)
						if err != nil {
							logger.Sugar().Errorf("Token update failed: cannot make token cookie: %v", err)
							return nil, status.Errorf(codes.Unauthenticated, "access denied")
						}
						ctx = metadata.AppendToOutgoingContext(ctx, "set-cookie", tokenCookie.String())
						if err := grpc.SetHeader(ctx, meta); err != nil {
							logger.Sugar().Errorf("Token update failed: cannot set header: %s", err)
							return nil, status.Errorf(codes.Unauthenticated, "access denied")
						}

					}
				}
			}
		}
		return handler(ctx, req)
	}
}
