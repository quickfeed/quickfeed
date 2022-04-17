package interceptors

import (
	"context"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// "user" role implies that user attempts to access information about himself.
// "group" role implies that the user is a course student + a member of the given group.
// "student" role implies that the user is enrolled in the course with any role.
// "teacher": user enrolled in the course with teacher status.
// "owner" is the creator of the manual review.
var accessRoles = map[string][]string{
	"GetUserByCourse":         {"admin", "teacher"},
	"GetSubmissionsByCourse":  {"admin", "teacher"},
	"UpdateUser":              {"admin", "user"},
	"GetEnrollmentsByUser":    {"admin", "user"},
	"GetOrganization":         {"admin"}, // TODO(vera): not needed in case of admin service
	"GetEnrollmentsByCourse":  {"student"},
	"UpdateReview":            {"teacher", "owner"}, // TODO(vera): a new role for review owner or just a server-side check?
	"GetGroupByUserAndCourse": {"teacher", "group"},
	"CreateGroup":             {"teacher", "group"},
	"GetGroup":                {"teacher", "group"}, // TODO(vera): "group" needs a db call or a field in claims
	"UpdateGroup":             {"teacher"},
	"DeleteGroup":             {"teacher"},
	"GetSubmissions":          {"teacher"},
	"IsEmptyRepo":             {"teacher"},
	"GetGroupsByCourse":       {"teacher"},
	"UpdateCourse":            {"teacher"},
	"UpdateEnrollments":       {"teacher"},
	"UpdateSubmission":        {"teacher"},
	"RebuildSubmissions":      {"teacher"},
	"CreateBenchmark":         {"teacher"},
	"UpdateBenchmark":         {"teacher"},
	"DeleteBenchmark":         {"teacher"},
	"CreateCriterion":         {"teacher"},
	"UpdateCriterion":         {"teacher"},
	"DeleteCriterion":         {"teacher"},
	"CreateReview":            {"teacher"},
	"UpdateSubmissions":       {"teacher"},
	"GetReviewers":            {"teacher"},
	"UpdateAssignments":       {"teacher"},
}

func AccessControl(logger *zap.Logger, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// get JWT from context
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Sugar().Errorf("Access control failed: missing metadata")
			return nil, status.Errorf(codes.Unauthenticated, "access denied")
		}
		logger.Sugar().Debugf("Request %s has metadata: %+v", info.FullMethod, meta) // tmp
		// get claims from jwt
		for _, c := range meta.Get("cookie") {
			fields := strings.Fields(c)
			for _, field := range fields {
				logger.Sugar().Debugf("metadata field: %s", field) // tmp
				if strings.Contains(field, "auth") {
					token := strings.Split(field, "=")[1]
					logger.Sugar().Debugf("extracted token: %s", token) // tmp
					claims, err := tokens.GetClaims(token)
					if err != nil {
						logger.Sugar().Errorf("Failed to extract claims from JWT: %v", err)
						return nil, status.Errorf(codes.Unauthenticated, "access denied")
					}
					// TODO(vera): refactor this part
					method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
					// If method is not in the map, there is no restrictions to call it
					roles, ok := accessRoles[method]
					logger.Sugar().Debugf("Access control: User %d calls method %s. Expected roles: %v", claims.UserID, method, roles)
					if ok {
						for _, role := range roles {
							switch role {
							// TODO(vera): refactor case handlers
							case "user":
								// Methods that can be accessed by the owner of the UserID can also be accessed by admin.
								// Skip other checks if user has admin role.
								if !claims.Admin {
									switch method {
									case "UpdateUser":
										if claims.UserID != req.(*pb.User).GetID() && !claims.Admin {
											logger.Sugar().Errorf("Access control failed: user %d attempted to change info for user %d", claims.UserID, req.(*pb.User).GetID())
											return nil, status.Errorf(codes.PermissionDenied, "permission denied")
										}
									case "GetEnrollmentsByUser":
										if claims.UserID != req.(*pb.EnrollmentStatusRequest).GetUserID() && !claims.Admin {
											logger.Sugar().Errorf("Access control failed: user %d requested enrollments for user %d", claims.UserID, req.(*pb.EnrollmentStatusRequest).GetUserID())
											return nil, status.Errorf(codes.PermissionDenied, "permission denied")
										}
									}
								}
								return handler(ctx, req)
							case "group":
								return handler(ctx, req)
							case "teacher":
								return handler(ctx, req)
							case "admin":
								if !claims.Admin {
									logger.Sugar().Errorf("Access control failed (method: %s): user is not admin", method)
									return nil, status.Errorf(codes.PermissionDenied, "permission denied")
								}
								return handler(ctx, req)
							default:
								logger.Sugar().Debugf("Unknown access role: %s", role)
							}
						}
					}
				}
			}
		}
		// check if expire

		// check if need update

		// update if needed: 1) new claims 2) set in cookie

		// check if the user allowed to call the method

		return handler(ctx, req)
	}
}
