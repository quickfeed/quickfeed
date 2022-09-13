package interceptor

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
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
	"GetSubmissions":          {student, group, teacher},
	"GetSubmission":           {teacher},
	"GetGroupByUserAndCourse": {student, teacher},
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
	"GetSubmissionsByCourse":  {teacher},
	"GetUserByCourse":         {teacher, admin},
	"GetUsers":                {admin},
	"GetOrganization":         {admin},
	"CreateCourse":            {admin},
}

type accessControlInterceptor struct {
	tokenManager *auth.TokenManager
}

func NewAccessControlInterceptor(tm *auth.TokenManager) *accessControlInterceptor {
	return &accessControlInterceptor{tokenManager: tm}
}

func (a *accessControlInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (a *accessControlInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

// AccessControl checks user information stored in the JWT claims against the list of roles required to call the method.
func (a *accessControlInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := request.Spec().Procedure
		method := procedure[strings.LastIndex(procedure, "/")+1:]
		req, ok := request.Any().(requestID)
		if !ok {
			return nil, connect.NewError(connect.CodeUnimplemented,
				fmt.Errorf("%s failed: message type %T does not implement IDFor interface", method, request))
		}
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok {
			return nil, connect.NewError(connect.CodePermissionDenied,
				fmt.Errorf("AccessControl(%s): failed to get claims from request context", method))
		}
		for _, role := range accessRolesFor[method] {
			switch role {
			case none:
				return next(ctx, request)
			case user:
				if claims.SameUser(req) {
					// Make sure the user is not updating own admin status.
					if method == "UpdateUser" {
						if req.(*qf.User).GetIsAdmin() && !claims.Admin {
							return nil, connect.NewError(connect.CodePermissionDenied,
								fmt.Errorf("AccessControl(%s): user %d attempted to change admin status from %v to %v",
									method, claims.UserID, claims.Admin, req.(*qf.User).GetIsAdmin()))
						}
					}
					return next(ctx, request)
				}
			case student:
				// GetSubmissions is used to fetch individual and group submissions.
				// For individual submissions needs an extra check for user ID in request.
				if method == "GetSubmissions" && req.IDFor("group") == 0 {
					if !claims.SameUser(req) {
						return nil, connect.NewError(connect.CodePermissionDenied,
							fmt.Errorf("AccessControl(%s): ID mismatch in claims (%d) and request (%d)",
								method, claims.UserID, req.IDFor("user")))
					}
				}
				if claims.HasCourseStatus(req, qf.Enrollment_STUDENT) {
					return next(ctx, request)
				}
			case group:
				// Request for CreateGroup will not have ID yet, need to check
				// if the user is in the group (unless teacher).
				if method == "CreateGroup" {
					notMember := !req.(*qf.Group).Contains(&qf.User{ID: claims.UserID})
					notTeacher := !claims.HasCourseStatus(req, qf.Enrollment_TEACHER)
					if notMember && notTeacher {
						return nil, connect.NewError(connect.CodePermissionDenied,
							fmt.Errorf("AccessControl(%s): user %d tried to create group while not teacher or group member", method, claims.UserID))
					}
					// Otherwise, create the group.
					return next(ctx, request)
				}
				groupID := req.IDFor("group")
				for _, group := range claims.Groups {
					if group == groupID {
						return next(ctx, request)
					}
				}
			case teacher:
				if method == "GetUserByCourse" {
					if err := claims.IsCourseTeacher(a.tokenManager.Database(), request.Any().(*qf.CourseUserRequest)); err != nil {
						return nil, connect.NewError(connect.CodePermissionDenied,
							fmt.Errorf("AccessControl(%s): %w", method, err))
					}
					return next(ctx, request)
				}
				if claims.HasCourseStatus(req, qf.Enrollment_TEACHER) {
					return next(ctx, request)
				}
			case admin:
				if claims.Admin {
					return next(ctx, request)
				}
			}
		}
		return nil, connect.NewError(connect.CodePermissionDenied,
			fmt.Errorf("AccessDenied(%s): required roles %v not satisfied by claims: %s", method, accessRolesFor[method], claims))
	})
}
