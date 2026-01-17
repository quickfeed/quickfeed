package interceptor

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/web/auth"
)

// accessChecker defines the function type used for checking access for a specific method.
// Returns an empty string if access is granted, or a reason string if denied.
type accessChecker func(db database.Database, req any, claims *auth.Claims) string

// accessGranted is the constant string returned when access is granted.
const accessGranted = ""

// Helper functions for common role checks

func checkNone(db database.Database, req any, claims *auth.Claims) string {
	return accessGranted
}

func checkUser(db database.Database, req any, claims *auth.Claims) string {
	if claims.SameUser(req) {
		return accessGranted
	}
	return "user ID mismatch"
}

func checkTeacher(db database.Database, req any, claims *auth.Claims) string {
	if claims.IsCourseTeacher(getCourseID(req)) {
		return accessGranted
	}
	return "not teacher"
}

func checkAdmin(db database.Database, req any, claims *auth.Claims) string {
	if claims.Admin {
		return accessGranted
	}
	return "not admin"
}

func checkUserOrStudentOrTeacherOrAdmin(db database.Database, req any, claims *auth.Claims) string {
	if claims.SameUser(req) {
		return accessGranted
	}
	if claims.IsCourseStudent(getCourseID(req)) {
		return accessGranted
	}
	if claims.IsCourseTeacher(getCourseID(req)) {
		return accessGranted
	}
	if claims.Admin {
		return accessGranted
	}
	return "not enrolled or not admin"
}

func checkStudentOrTeacher(db database.Database, req any, claims *auth.Claims) string {
	if claims.IsCourseStudent(getCourseID(req)) {
		return accessGranted
	}
	if claims.IsCourseTeacher(getCourseID(req)) {
		return accessGranted
	}
	return "not student or teacher"
}

func checkGroupOrTeacher(db database.Database, req any, claims *auth.Claims) string {
	if claims.IsInGroup(req) {
		return accessGranted
	}
	if claims.IsCourseTeacher(getCourseID(req)) {
		return accessGranted
	}
	return "not group member or teacher"
}

// methodCheckers maps each method to its corresponding access checker function.
// Each checker returns an empty string if access is granted, or a reason string if denied.
var methodCheckers = map[string]accessChecker{
	"GetUser":                checkNone,
	"GetCourse":              checkNone,
	"GetCourses":             checkNone,
	"SubmissionStream":       checkNone, // No role required as long as the user is authenticated, i.e. has a valid token.
	"CreateEnrollment":       checkUser,
	"UpdateCourseVisibility": checkUser,
	"UpdateUser": func(db database.Database, req any, claims *auth.Claims) string { // roles: {user, admin},
		if claims.SameUser(req) {
			if claims.UnauthorizedAdminChange(req) {
				return fmt.Sprintf("non-admin user %d attempted to grant admin privileges", claims.UserID)
			}
			return accessGranted
		}
		if claims.Admin {
			return accessGranted
		}
		return "user ID mismatch or not admin"
	},
	"GetEnrollments": checkUserOrStudentOrTeacherOrAdmin,
	"GetSubmissions": func(db database.Database, req any, claims *auth.Claims) string { // roles: {student, group, teacher},
		// student role
		if !hasGroupID(req) {
			if !claims.SameUser(req) {
				return fmt.Sprintf("ID mismatch in claims (%d) and request (%d)", claims.UserID, getUserID(req))
			}
		}
		if claims.IsCourseStudent(getCourseID(req)) {
			return accessGranted
		}
		// group role
		if claims.IsInGroup(req) {
			return accessGranted
		}
		// teacher role
		if claims.IsCourseTeacher(getCourseID(req)) {
			return accessGranted
		}
		return "not student, group member, or teacher"
	},
	"GetSubmission": checkTeacher,
	"CreateGroup": func(db database.Database, req any, claims *auth.Claims) string { // roles: {group, teacher},
		// group role
		notMember := !claims.IsGroupMember(req)
		notTeacher := !claims.IsCourseTeacher(getCourseID(req))
		if notMember && notTeacher {
			return fmt.Sprintf("user %d tried to create group while not teacher or group member", claims.UserID)
		}
		return accessGranted
	},
	"GetGroup":          checkGroupOrTeacher,
	"GetAssignments":    checkStudentOrTeacher,
	"GetRepositories":   checkStudentOrTeacher,
	"UpdateGroup":       checkTeacher,
	"DeleteGroup":       checkTeacher,
	"GetGroupsByCourse": checkTeacher,
	"UpdateCourse":      checkTeacher,
	"UpdateEnrollments": checkTeacher,
	"UpdateAssignments": checkTeacher,
	"UpdateSubmission": func(db database.Database, req any, claims *auth.Claims) string { // roles: {teacher},
		if !isValidSubmission(db, req) {
			return "invalid submission"
		}
		if claims.IsCourseTeacher(getCourseID(req)) {
			return accessGranted
		}
		return "not teacher"
	},
	"UpdateSubmissions": checkTeacher,
	"RebuildSubmissions": func(db database.Database, req any, claims *auth.Claims) string { // roles: {teacher},
		if getSubmissionID(req) != 0 && !isValidSubmission(db, req) {
			return "invalid submission"
		}
		if claims.IsCourseTeacher(getCourseID(req)) {
			return accessGranted
		}
		return "not teacher"
	},
	"CreateBenchmark":          checkTeacher,
	"UpdateBenchmark":          checkTeacher,
	"DeleteBenchmark":          checkTeacher,
	"CreateCriterion":          checkTeacher,
	"UpdateCriterion":          checkTeacher,
	"DeleteCriterion":          checkTeacher,
	"CreateReview":             checkTeacher,
	"UpdateReview":             checkTeacher,
	"CreateAssignmentFeedback": checkStudentOrTeacher,
	"GetAssignmentFeedback":    checkTeacher,
	"IsEmptyRepo":              checkTeacher,
	"GetSubmissionsByCourse":   checkTeacher,
	"GetUsers":                 checkAdmin,
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

// WrapUnary checks user information stored in the JWT claims against the access checker for the method.
func (a *AccessControlInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := request.Spec().Procedure
		method := procedure[strings.LastIndex(procedure, "/")+1:]
		req := request.Any()
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok {
			return nil, accessDeniedError(method, "failed to get claims from request context")
		}
		checker, ok := methodCheckers[method]
		if !ok {
			return nil, accessDeniedError(method, "unknown method")
		}
		if reason := checker(a.db, req, claims); reason != "" {
			return nil, accessDeniedError(method, reason)
		}
		return next(ctx, request)
	})
}

// accessDeniedError creates a standardized access denied error for the given method and reason.
func accessDeniedError(method, reason string) error {
	return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("access denied for %s: %s", method, reason))
}
