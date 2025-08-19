package interceptor

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
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
	"GetUser":                  {none},
	"GetCourse":                {none},
	"GetCourses":               {none},
	"SubmissionStream":         {none}, // No role required as long as the user is authenticated, i.e. has a valid token.
	"CreateEnrollment":         {user},
	"UpdateCourseVisibility":   {user},
	"UpdateUser":               {user, admin},
	"GetEnrollments":           {user, student, teacher, admin},
	"GetSubmissions":           {student, group, teacher},
	"GetSubmission":            {teacher},
	"CreateGroup":              {group, teacher},
	"GetGroup":                 {group, teacher},
	"GetAssignments":           {student, teacher},
	"GetRepositories":          {student, teacher},
	"UpdateGroup":              {teacher},
	"DeleteGroup":              {teacher},
	"GetGroupsByCourse":        {teacher},
	"UpdateCourse":             {teacher},
	"UpdateEnrollments":        {teacher},
	"UpdateAssignments":        {teacher},
	"UpdateSubmission":         {teacher},
	"UpdateSubmissions":        {teacher},
	"RebuildSubmissions":       {teacher},
	"CreateBenchmark":          {teacher},
	"UpdateBenchmark":          {teacher},
	"DeleteBenchmark":          {teacher},
	"CreateCriterion":          {teacher},
	"UpdateCriterion":          {teacher},
	"DeleteCriterion":          {teacher},
	"CreateReview":             {teacher},
	"UpdateReview":             {teacher},
	"CreateAssignmentFeedback": {student, teacher},
	"GetAssignmentFeedback":    {student, teacher},
	"IsEmptyRepo":              {teacher},
	"GetSubmissionsByCourse":   {teacher},
	"GetUsers":                 {admin},
}

type AccessControlInterceptor struct {
	tokenManager *auth.TokenManager
}

func NewAccessControlInterceptor(tm *auth.TokenManager) *AccessControlInterceptor {
	return &AccessControlInterceptor{tokenManager: tm}
}

func (*AccessControlInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (*AccessControlInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

// WrapUnary checks user information stored in the JWT claims against the list of roles required to call the method.
func (a *AccessControlInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := request.Spec().Procedure
		method := procedure[strings.LastIndex(procedure, "/")+1:]
		req, ok := request.Any().(requestID)
		if !ok {
			return nil, connect.NewError(connect.CodeUnimplemented,
				fmt.Errorf("access denied for %s: message type %T does not implement 'requestID' interface", method, request))
		}
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok {
			return nil, connect.NewError(connect.CodePermissionDenied,
				fmt.Errorf("access denied for %s: failed to get claims from request context", method))
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
								fmt.Errorf("access denied for %s: user %d attempted to change admin status from %v to %v",
									method, claims.UserID, claims.Admin, req.(*qf.User).GetIsAdmin()))
						}
					}
					return next(ctx, request)
				}
			case student:
				// GetSubmissions is used to fetch individual and group submissions.
				// For individual submissions needs an extra check for user ID in request.
				if method == "GetSubmissions" {
					if req.IDFor("group") != 0 {
						// Group submissions are handled by the group role.
						continue
					}
					if !claims.SameUser(req) {
						return nil, connect.NewError(connect.CodePermissionDenied,
							fmt.Errorf("access denied for %s: ID mismatch in claims (%d) and request (%d)",
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
							fmt.Errorf("access denied for %s: user %d tried to create group while not teacher or group member", method, claims.UserID))
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
				if method == "RebuildSubmissions" || method == "UpdateSubmission" {
					if !isValidSubmission(a.tokenManager.Database(), req) {
						return nil, connect.NewError(connect.CodePermissionDenied,
							fmt.Errorf("access denied for %s: %v", method, "invalid submission"))
					}
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
			fmt.Errorf("access denied for %s: required roles %v not satisfied by claims: %s", method, accessRolesFor[method], claims))
	})
}
