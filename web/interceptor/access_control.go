package interceptor

import (
	"context"
	"reflect"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	role      int
	roles     []role
	requestID interface {
		IDFor(string) uint64
	}
)

const (
	none role = iota
	// user role implies that user attempts to access information about himself.
	user
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
var accessRolesFor = map[string]roles{
	"GetUser":                 {none},
	"GetCourse":               {none},
	"GetCourses":              {none},
	"CreateEnrollment":        {user},
	"UpdateCourseVisibility":  {user},
	"GetCoursesByUser":        {user},
	"UpdateUser":              {user, admin},
	"GetEnrollmentsByUser":    {user, admin},
	"GetSubmissions":          {student, group, teacher, courseAdmin},
	"GetGroupByUserAndCourse": {group, teacher},
	"CreateGroup":             {group, teacher},
	"GetGroup":                {group, teacher},
	"GetAssignments":          {student, teacher},
	"GetEnrollmentsByCourse":  {student, teacher},
	"GetRepositories":         {student, teacher},
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
	"GetSubmissionsByCourse":  {courseAdmin},
	"GetUserByCourse":         {teacher, admin},
	"GetUsers":                {admin},
	"GetOrganization":         {admin},
	"CreateCourse":            {admin},
}

// AccessControl checks user information stored in the JWT claims against the list of roles required to call the method.
func AccessControl(logger *zap.SugaredLogger, tm *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		req, ok := request.(requestID)
		// The GetUserByCourse method sends a CourseUserRequest which has no IDs and needs a database query.
		if !ok {
			logger.Errorf("%s failed: message type '%s' does not implement IDFor interface",
				method, reflect.TypeOf(request).String())
			return nil, ErrAccessDenied
		}
		claims, err := tm.GetClaims(ctx)
		if err != nil {
			logger.Errorf("AccessControl(%s): failed to get claims from request context: %v", method, err)
			return handler(ctx, request)
		}
		for _, role := range accessRolesFor[method] {
			switch role {
			case none:
				return handler(ctx, request)
			case user:
				if claims.SameUser(req) {
					// Make sure the user is not updating own admin status.
					if method == "UpdateUser" {
						if req.(*qf.User).GetIsAdmin() && !claims.Admin {
							logger.Errorf("AccessControl(%s): user %d attempted to change admin status from %v to %v",
								method, claims.UserID, claims.Admin, req.(*qf.User).GetIsAdmin())
							return nil, ErrAccessDenied
						}
					}
					return handler(ctx, request)
				}
			case student:
				// GetSubmissions is used to fetch individual and group submissions.
				// For individual submissions needs an extra check for user ID in request.
				if method == "GetSubmissions" && req.IDFor("group") == 0 {
					if !claims.SameUser(req) {
						logger.Errorf("AccessControl(%s): ID mismatch in claims (%s) and request (%s)",
							method, claims.UserID, req.IDFor("user"))
						return nil, ErrAccessDenied
					}
				}
				if claims.HasCourseStatus(req, qf.Enrollment_STUDENT) {
					return handler(ctx, request)
				}
			case group:
				// Request for CreateGroup will not have ID yet, need to check
				// if the user is in the group (unless teacher).
				if method == "CreateGroup" {
					notMember := !req.(*qf.Group).Contains(&qf.User{ID: claims.UserID})
					notTeacher := !claims.HasCourseStatus(req, qf.Enrollment_TEACHER)
					if notMember && notTeacher {
						logger.Errorf("AccessControl(%s): user %d tried to create group while not teacher or group member", method, claims.UserID)
						return nil, ErrAccessDenied
					}
					// Otherwise, create the group.
					return handler(ctx, request)
				}
				groupID := req.IDFor("group")
				for _, group := range claims.Groups {
					if group == groupID {
						return handler(ctx, request)
					}
				}
			case teacher:
				if method == "GetUserByCourse" {
					if err := claims.IsCourseTeacher(tm.Database(), request.(*qf.CourseUserRequest)); err != nil {
						logger.Errorf("AccessControl(%s): %v", method, err)
						return nil, ErrAccessDenied
					}
					return handler(ctx, request)
				}
				if claims.HasCourseStatus(req, qf.Enrollment_TEACHER) {
					return handler(ctx, request)
				}
			case courseAdmin:
				if claims.Admin {
					if method == "GetUserByCourse" {
						if err := claims.IsCourseTeacher(tm.Database(), request.(*qf.CourseUserRequest)); err != nil {
							logger.Errorf("AccessControl(%s): %v", method, err)
							return nil, ErrAccessDenied
						}
						return handler(ctx, request)
					}
					if claims.HasCourseStatus(req, qf.Enrollment_TEACHER) {
						return handler(ctx, request)
					}
				}
			case admin:
				if claims.Admin {
					return handler(ctx, request)
				}
			}
		}
		logger.Errorf("AccessDenied(%s): required roles %v not satisfied by claims: %s", method, accessRolesFor[method], claims)
		return nil, ErrAccessDenied
	}
}
