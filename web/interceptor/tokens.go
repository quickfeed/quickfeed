package interceptor

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/web/auth"
)

type (
	// The userIDs interface must be implemented by request types that may need to update the tokens.
	userIDs interface{ UserIDs() []uint64 }
	// Marker interface to detect the GroupRequest type needed for DeleteGroup.
	isGroup interface{ GetGroupID() uint64 }
)

var defaultTokenUpdater = func(_ context.Context, tm *auth.TokenManager, msg userIDs) error {
	for _, userID := range msg.UserIDs() {
		if err := tm.Add(userID); err != nil {
			return err
		}
	}
	return nil
}

// tokenUpdateMethods is a map of methods that require updating the list of users who need a new JWT.
var tokenUpdateMethods = map[string]func(context.Context, *auth.TokenManager, userIDs) error{
	"UpdateUser":        defaultTokenUpdater, // User has been promoted to admin or demoted.
	"UpdateGroup":       defaultTokenUpdater, // Users added to a group or removed from a group.
	"UpdateEnrollments": defaultTokenUpdater, // User enrolled into a new course or promoted to TA.

	"CreateCourse": // The signed in user gets the teacher role in the new course.
	func(ctx context.Context, tm *auth.TokenManager, _ userIDs) error {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok {
			return fmt.Errorf("CreateCourse: missing claims in context")
		}
		return tm.Add(claims.UserID)
	},

	"DeleteGroup": // Group members removed from the group.
	func(ctx context.Context, tm *auth.TokenManager, msg userIDs) error {
		if grp, ok := msg.(isGroup); ok {
			group, err := tm.Database().GetGroup(grp.GetGroupID())
			if err != nil {
				return err
			}
			return defaultTokenUpdater(ctx, tm, group)
		}
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot update token for %s: request does not contain a group", "DeleteGroup"))
	},
}

type TokenInterceptor struct {
	tokenManager *auth.TokenManager
}

func NewTokenInterceptor(tm *auth.TokenManager) *TokenInterceptor {
	return &TokenInterceptor{tokenManager: tm}
}

func (*TokenInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (*TokenInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return nil // not supported
	})
}

// WrapUnary updates list of users who need a new JWT next time they send a request to the server.
// This method only logs errors to avoid overwriting the gRPC error messages returned by the server.
func (t *TokenInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := request.Spec().Procedure
		method := procedure[strings.LastIndex(procedure, "/")+1:]
		if tokenUpdateFn, ok := tokenUpdateMethods[method]; ok {
			if msg, ok := request.Any().(userIDs); ok {
				if err := tokenUpdateFn(ctx, t.tokenManager, msg); err != nil {
					return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot update token for %s: %w", method, err))
				}
			} else {
				return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("cannot update token for %s: message type %T does not implement 'userIDs' interface", method, request))
			}
		}
		return next(ctx, request)
	})
}
