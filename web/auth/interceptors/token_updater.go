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

var methods []string = []string{
	"UpdateUser",
	"CreateCourse",
	"UpdateEnrollments",
}

func UpdateTokens(logger *zap.Logger, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug("TOKEN UPDATE INTERCEPTOR")
		sort.Strings(methods)
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		if sort.SearchStrings(methods, method) < len(methods) {
			resp, err := handler(ctx, req)
			// We only want to add IDs to the list of tokens that need update if the request was successful
			if err == nil {
				switch method {
				case "UpdateUser":
					tokens.Add(req.(*pb.User).GetID())
				case "CreateCourse":
					// TODO(vera): needs to extract JWT to get current ID, actually can be updated right here
				case "UpdateEnrollments":
					for _, enrol := range req.(*pb.Enrollments).GetEnrollments() {
						userID := enrol.GetUserID()
						// If a group enrollment is updated user ID will be 0, ignore
						if userID > 0 {
							tokens.Add(userID)
						}
					}
				}
			}
			return resp, err
		}
		return handler(ctx, req)
	}
}
