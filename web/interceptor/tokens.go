package interceptor

import (
	"context"
	"sort"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
// The list is updated only if the called method is in the list and has been successful (no error has been returned by the server).
// This method only logs errors to avoid overwriting the gRPC error messages returned by the server.
func TokenInterceptor(logger *zap.SugaredLogger, tm *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		idx := sort.SearchStrings(methods, method)
		if idx < len(methods) && methods[idx] == method {
			logger.Debugf("Token manager found valid method: %s", method) // tmp
			// Pass the request to the server and inspect the response.
			resp, grpcErr := handler(ctx, req)
			if grpcErr == nil {
				switch method {
				// User has been promoted to admin or demoted.
				case "UpdateUser":
					// Add ID of the user from the request payload to the list.
					if err := tm.Add(req.(*qf.User).GetID()); err != nil {
						logger.Error(err)
						return resp, grpcErr
					}
				// The signed in user gets the teacher role in the new course.
				case "CreateCourse":
					meta, ok := metadata.FromIncomingContext(ctx)
					if !ok {
						logger.Error("TokenInterceptor: failed to extract metadata")
						return resp, grpcErr
					}
					token, err := extractToken(meta)
					if err != nil {
						logger.Errorf("TokenInterceptor: failed to extract authentication token: %v", err)
						return resp, grpcErr
					}
					claims, err := tm.GetClaims(token)
					if err != nil {
						logger.Error(err)
						return resp, grpcErr
					}
					if err := tm.Add(claims.UserID); err != nil {
						logger.Error(err)
						return resp, grpcErr
					}
				// User enrolled into a new course or promoted to TA.
				case "UpdateEnrollments":
					for _, enrol := range req.(*qf.Enrollments).GetEnrollments() {
						if err := tm.Add(enrol.GetUserID()); err != nil {
							logger.Error(err)
							return resp, grpcErr
						}
					}
				// Users added to a group or removed from a group.
				case "UpdateGroup":
					for _, user := range req.(*qf.Group).GetUsers() {
						if err := tm.Add(user.GetID()); err != nil {
							logger.Error(err)
							return resp, grpcErr
						}
					}
				case "DeleteGroup":
					group, err := tm.Database().GetGroup(req.(*qf.GroupRequest).GetGroupID())
					if err != nil {
						logger.Errorf("TokenInterceptor: failed to get group from database: %v", err)
						return resp, grpcErr
					}
					for _, user := range group.GetUsers() {
						if err := tm.Add(user.GetID()); err != nil {
							logger.Error(err)
							return resp, grpcErr
						}
					}
				}
			}
		}
		return handler(ctx, req)
	}
}
