package interceptors

import (
	"context"
	"sort"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/web/auth/tokens"
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
// This method only logs errors to avoid overwriting the gRPC responses.
func UpdateTokens(logger *zap.SugaredLogger, tokens *tokens.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		sort.Strings(methods)
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		idx := sort.SearchStrings(methods, method)
		if idx < len(methods) && methods[idx] == method {
			logger.Debugf("Token updater found method with index %d", sort.SearchStrings(methods, method)) // tmp
			logger.Debugf("Interceptors: updating token list on method %s", method)                        // tmp
			resp, err := handler(ctx, req)
			// We only want to add the user ID to the list of tokens to update
			// if the request was successful
			if err == nil {
				switch method {
				// User has been promoted to admin or demoted.
				case "UpdateUser":
					// Add id of the user whose info has been updated.
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
				// Users has been enrolled into a course or promoted to TA.
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
		logger.Debugf("Token update interceptor took %v", time.Since(start))
		return handler(ctx, req)
	}
}
