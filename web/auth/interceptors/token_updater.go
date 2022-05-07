package interceptors

import (
	"context"
	"sort"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var methods = []string{
	"UpdateUser",
	"CreateCourse",
	"UpdateEnrollments",
	"UpdateGroup",
}

// UpdateTokens adds relevant user IDs to the list of users that need their token refreshed
// next time they sign in because their access roles might have changed
// This method only logs errors to avoid overwriting the gRPC response status.
func UpdateTokens(logger *zap.SugaredLogger, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sort.Strings(methods)
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		// There are only three methods
		if sort.SearchStrings(methods, method) < len(methods) {
			logger.Debugf("Interceptors: updating token list on method %s", method)
			resp, err := handler(ctx, req)
			// We only want to add the user ID to the list of tokens to update
			// if the request was successful
			if err == nil {
				switch method {
				// User has been promoted to admin or demoted.
				case "UpdateUser":
					if err := tokens.Add(req.(*pb.User).GetID()); err != nil {
						logger.Error(err)
					}
				// The signed in user gets a teacher role for the new course.
				case "CreateCourse":
					token, err := GetFromMetadata(ctx, "token", "")
					if err != nil {
						logger.Error(err)
					}
					claims, err := tokens.GetClaims(token)
					if err != nil {
						logger.Error(err)
					}
					if err := tokens.Add(claims.UserID); err != nil {
						logger.Error(err)
					}
				// Users get enrolled into a course.
				case "UpdateEnrollments":
					for _, enrol := range req.(*pb.Enrollments).GetEnrollments() {
						if err := tokens.Add(enrol.GetUserID()); err != nil {
							logger.Error(err)
						}
					}
				// Group is approved or modified.
				case "UpdateGroup":
					for _, user := range req.(*pb.Group).GetUsers() {
						if err := tokens.Add(user.GetID()); err != nil {
							logger.Error(err)
						}
					}
				}
			}
			return resp, err
		}
		return handler(ctx, req)
	}
}
