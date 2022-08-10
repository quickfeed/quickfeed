package interceptor

import (
	"context"
	"strings"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	role      int
	roles     []role
	requestID interface {
		FetchID(string) uint64
	}
)

const (
	// user role implies that user attempts to access information about himself.
	user role = iota
	// group role implies that the user is a course student + a member of the given group.
	group
	// student role implies that the user is enrolled in the course with any role.
	student
	// teacher: user enrolled in the course with teacher status.
	teacher
	// courseAdmin: an admin user who is also enrolled into the course.
	courseAdmin
	// admin is the user with admin privileges.
	admin
)

// If there are several roles that can call a method, a role with the least privilege must come first.
// If method is not in the map, there is no restrictions to call it.
var access = map[string]roles{
	"GetEnrollmentsByCourse":  {student, teacher},
	"UpdateUser":              {user, admin},
	"GetEnrollmentsByUser":    {user, admin},
	"GetSubmissions":          {user, group, teacher, courseAdmin},
	"GetGroupByUserAndCourse": {group, teacher},
	"CreateGroup":             {group, teacher},
	"GetGroup":                {group, teacher},
	"UpdateGroup":             {teacher},
	"DeleteGroup":             {teacher},
	"GetGroupsByCourse":       {teacher},
	"UpdateCourse":            {teacher},
	"UpdateEnrollments":       {teacher},
	"UpdateAssignments":       {teacher},
	"UpdateSubmission":        {teacher},
	"UpdateSubmissions":       {teacher},
	"RebuildSubmissions":      {teacher},
	"CreateBenchmark":         {teacher},
	"UpdateBenchmark":         {teacher},
	"DeleteBenchmark":         {teacher},
	"CreateCriterion":         {teacher},
	"UpdateCriterion":         {teacher},
	"DeleteCriterion":         {teacher},
	"CreateReview":            {teacher},
	"UpdateReview":            {teacher},
	"GetReviewers":            {teacher},
	"IsEmptyRepo":             {teacher},
	"GetSubmissionsByCourse":  {teacher, courseAdmin},
	"GetUserByCourse":         {teacher, admin},
	"GetUsets":                {admin},
	"GetOrganization":         {admin},
	"CreateCourse":            {admin},
}

func AccessControl(logger *zap.SugaredLogger, db database.Database, tm *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		logger.Debugf("ACCESS CONTROL for method %s", method) // tmp
		roles, ok := access[method]
		if ok {
			logger.Debug("Got roles: ", roles) // tmp
			claims, err := tm.GetClaims(ctx)
			if err != nil {
				logger.Error("Access control: failed to get claims from request context: %v", err)
				return handler(ctx, req)
			}
			logger.Debug("Got user claims: ", claims) // tmp
		}

		return handler(ctx, req)
	}
}
