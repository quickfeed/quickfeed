package interceptors

import (
	"context"

	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func AccessControl(logger *zap.Logger, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// get claims from jwt

		// check if expire

		// check if need update

		// update if needed: 1) new claims 2) set in cookie

		// check if the user allowed to call the method
		return handler(ctx, req)
	}
}
