package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/web/auth"
)

type (
	role  int
	roles []role
)

//go:generate stringer -type=role
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
	"GetAssignmentFeedback":    {teacher},
	"IsEmptyRepo":              {teacher},
	"GetSubmissionsByCourse":   {teacher},
	"GetUsers":                 {admin},
}

// accessChecker defines the function type used for checking access for a specific role.
type accessChecker func(db database.Database, method string, req any, claims *auth.Claims) error

var (
	accessGranted         error = nil
	errContinueToNextRole       = errors.New("continue to next role checker")
)

// roleCheckers maps each role to its corresponding access checker function.
// Each checker returns nil if access is granted, errContinueToNextRole to try the next role,
// or any other error indicates access is denied.
var roleCheckers = map[role]accessChecker{
	none: func(db database.Database, method string, req any, claims *auth.Claims) error {
		return accessGranted
	},
	user: func(db database.Database, method string, req any, claims *auth.Claims) error {
		if claims.SameUser(req) {
			if method == "UpdateUser" && claims.UnauthorizedAdminChange(req) {
				return accessDeniedError(method, "non-admin user %d attempted to grant admin privileges", claims.UserID)
			}
			return accessGranted
		}
		return errContinueToNextRole
	},
	student: func(db database.Database, method string, req any, claims *auth.Claims) error {
		if method == "GetSubmissions" {
			if claims.IsGroupRequest(req) {
				return errContinueToNextRole // handled by group role
			}
			if !claims.SameUser(req) {
				return accessDeniedError(method, "ID mismatch in claims (%d) and request (%d)", claims.UserID, getUserID(req))
			}
		}
		if claims.IsCourseStudent(getCourseID(req)) {
			return accessGranted
		}
		return errContinueToNextRole
	},
	group: func(db database.Database, method string, req any, claims *auth.Claims) error {
		if method == "CreateGroup" {
			// Allow group creation if the user is either a teacher or a member of the group.
			notMember := !claims.IsGroupMember(req)
			notTeacher := !claims.IsCourseTeacher(getCourseID(req))
			if notMember && notTeacher {
				return accessDeniedError(method, "user %d tried to create group while not teacher or group member", claims.UserID)
			}
			return accessGranted
		}
		if claims.IsInGroup(req) {
			return accessGranted
		}
		return errContinueToNextRole
	},
	teacher: func(db database.Database, method string, req any, claims *auth.Claims) error {
		// Valid submission check is not needed for rebuilding all submissions (submissionID == 0).
		shouldValidate := method == "UpdateSubmission" || (method == "RebuildSubmissions" && getSubmissionID(req) != 0)
		if shouldValidate && !isValidSubmission(db, req) {
			return accessDeniedError(method, "invalid submission")
		}
		if claims.IsCourseTeacher(getCourseID(req)) {
			return accessGranted
		}
		return errContinueToNextRole
	},
	admin: func(db database.Database, method string, req any, claims *auth.Claims) error {
		if claims.Admin {
			return accessGranted
		}
		return errContinueToNextRole
	},
}

type AccessControlInterceptor struct {
	db database.Database
}

func NewAccessControlInterceptor(db database.Database) *AccessControlInterceptor {
	return &AccessControlInterceptor{db: db}
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
		req := request.Any()
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok {
			return nil, accessDeniedError(method, "failed to get claims from request context")
		}
		for _, role := range accessRolesFor[method] {
			if err := roleCheckers[role](a.db, method, req, claims); err == nil {
				return next(ctx, request)
			} else if err != errContinueToNextRole {
				return nil, err
			}
		}
		return nil, accessDeniedError(method, "required roles %v not satisfied by claims: %s", accessRolesFor[method], claims)
	})
}

// accessDeniedError creates a standardized access denied error for the given method and reason.
func accessDeniedError(method, reason string, args ...any) error {
	message := fmt.Sprintf("access denied for %s: %s", method, fmt.Sprintf(reason, args...))
	return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("%s", message))
}
