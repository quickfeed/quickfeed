package interceptor

import (
	"context"
	"sort"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var methods = []string{
	"UpdateUser",
	"CreateCourse",
	"UpdateEnrollments",
	"UpdateGroup",
	"DeleteGroup",
}

func init() {
	sort.Strings(methods)
}

// TokenManager updates list of users who need a new JWT next time they send a request to the server.
// This method only logs errors to avoid overwriting the gRPC error messages returned by the server.
func TokenRefresher(logger *zap.SugaredLogger, tm *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		idx := sort.SearchStrings(methods, method)
		if idx < len(methods) && methods[idx] == method {
			switch method {
			// User has been promoted to admin or demoted.
			case "UpdateUser":
				// Add ID of the user from the request payload to the list.
				if err := tm.Add(req.(*qf.User).GetID()); err != nil {
					logger.Error(err)
				}
			// The signed in user gets the teacher role in the new course.
			case "CreateCourse":
				claims, err := tm.GetClaims(ctx)
				if err != nil {
					logger.Error(err)
				}
				if err := tm.Add(claims.UserID); err != nil {
					logger.Error(err)
				}
			// User enrolled into a new course or promoted to TA.
			case "UpdateEnrollments":
				for _, enrol := range req.(*qf.Enrollments).GetEnrollments() {
					if err := tm.Add(enrol.GetUserID()); err != nil {
						logger.Error(err)
					}
				}
			// Users added to a group or removed from a group.
			case "UpdateGroup":
				for _, user := range req.(*qf.Group).GetUsers() {
					if err := tm.Add(user.GetID()); err != nil {
						logger.Error(err)
					}
				}
			case "DeleteGroup":
				group, err := tm.Database().GetGroup(req.(*qf.GroupRequest).GetGroupID())
				if err != nil {
					logger.Errorf("TokenInterceptor: failed to get group from database: %v", err)
				}
				for _, user := range group.GetUsers() {
					if err := tm.Add(user.GetID()); err != nil {
						logger.Error(err)
					}
				}
			}
		}
		return handler(ctx, req)
	}
}
