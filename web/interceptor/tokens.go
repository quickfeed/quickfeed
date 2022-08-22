package interceptor

import (
	"context"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
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
		claims, err := tm.GetClaims(ctx)
		if err != nil {
			return err
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
		return ErrAccessDenied
	},
}

// TokenRefresher updates list of users who need a new JWT next time they send a request to the server.
// This method only logs errors to avoid overwriting the gRPC error messages returned by the server.
func TokenRefresher(logger *zap.SugaredLogger, tm *auth.TokenManager) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := request.Spec().Procedure
			method := procedure[strings.LastIndex(procedure, "/")+1:]
			if tokenUpdateFn, ok := tokenUpdateMethods[method]; ok {
				if msg, ok := request.Any().(userIDs); ok {
					if err := tokenUpdateFn(ctx, tm, msg); err != nil {
						logger.Error(err)
						return nil, ErrAccessDenied
					}
				} else {
					logger.Errorf("%s's argument is missing 'userIDs' interface", method)
					return nil, ErrAccessDenied
				}
			}
			return next(ctx, request)
		})
	})
}
