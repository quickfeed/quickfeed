package interceptor

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
)

// accessChecker defines the function type used for checking access for a specific method.
// Returns an empty string if access is granted, or a reason string if denied.
type accessChecker func(db database.Database, req any, claims *auth.Claims) string

// accessGranted is the constant string returned when access is granted.
const accessGranted = ""

// Below are access checker functions for common role checks.
// The main roles include: none, user, student, group, teacher, admin, and combinations thereof.
// These checker functions can be used for different RPC methods as needed.

// checkNone allows access to any authenticated user.
func checkNone(db database.Database, req any, claims *auth.Claims) string {
	return accessGranted
}

// checkUser checks if the user ID in the request matches the user ID in the claims.
// The [req] is expected to implement [userIDProvider].
func checkUser(db database.Database, req any, claims *auth.Claims) string {
	if claims.SameUser(req) { // user role
		return accessGranted
	}
	return "user ID mismatch"
}

// checkTeacher checks if the user is a teacher in the course specified in the request.
// The [req] is expected to implement [courseIDProvider].
func checkTeacher(db database.Database, req any, claims *auth.Claims) string {
	if claims.IsCourseTeacher(getCourseID(req)) { // teacher role in course
		return accessGranted
	}
	return "not teacher"
}

// checkAdmin checks if the user has admin privileges.
func checkAdmin(db database.Database, req any, claims *auth.Claims) string {
	if claims.Admin { // admin role
		return accessGranted
	}
	return "not admin"
}

// checkUserOrStudentOrTeacherOrAdmin checks if the user is the same as in the request,
// or is a student or teacher in the course specified in the request, or an admin.
// The [req] is expected to implement [userIDProvider] or [courseIDProvider].
func checkUserOrStudentOrTeacherOrAdmin(db database.Database, req any, claims *auth.Claims) string {
	if claims.SameUser(req) { // user role
		return accessGranted
	}
	if claims.IsCourseStudent(getCourseID(req)) { // student role in course
		return accessGranted
	}
	if claims.IsCourseTeacher(getCourseID(req)) { // teacher role in course
		return accessGranted
	}
	if claims.Admin { // admin role
		return accessGranted
	}
	return "not enrolled or not admin"
}

// checkStudentOrTeacher checks if the user is a student or teacher in the course specified in the request.
// The [req] is expected to implement [courseIDProvider].
func checkStudentOrTeacher(db database.Database, req any, claims *auth.Claims) string {
	if claims.IsCourseStudent(getCourseID(req)) { // student role in course
		return accessGranted
	}
	if claims.IsCourseTeacher(getCourseID(req)) { // teacher role in course
		return accessGranted
	}
	return "not student or teacher"
}

// checkGroupOrTeacher checks if the user is a member of the group specified in the request,
// or is a teacher in the course specified in the request.
// The [req] is expected to implement [groupIDProvider] or [courseIDProvider].
func checkGroupOrTeacher(db database.Database, req any, claims *auth.Claims) string {
	if claims.IsGroupMember(req) { // CreateGroup: claims user must be member of the group being created
		return accessGranted
	}
	if claims.IsInGroup(req) { // GetGroup: request's group ID must be in the claims' groups to allow access
		return accessGranted
	}
	if claims.IsCourseTeacher(getCourseID(req)) { // teacher role in course
		return accessGranted
	}
	return "not group member or teacher"
}

// checkUpdateUser checks if the user is updating their own information or if they are an admin.
// The [req] is expected to implement [userIDProvider].
func checkUpdateUser(db database.Database, req any, claims *auth.Claims) string {
	if claims.SameUser(req) { // user role
		if claims.UnauthorizedAdminChange(req) {
			return fmt.Sprintf("non-admin user %d attempted to grant admin privileges", claims.UserID)
		}
		return accessGranted
	}
	if claims.Admin { // admin role
		return accessGranted
	}
	return "user ID mismatch or not admin"
}

// checkGetSubmissions checks if the user is a student, group member, or teacher for accessing submissions.
// The [req] is expected to implement [userIDProvider] or [groupIDProvider] or [courseIDProvider].
func checkGetSubmissions(db database.Database, req any, claims *auth.Claims) string { // roles: student, group, teacher
	if !hasGroupID(req) { // student role
		if !claims.SameUser(req) {
			return fmt.Sprintf("ID mismatch in claims (%d) and request (%d)", claims.UserID, getUserID(req))
		}
		if claims.IsCourseStudent(getCourseID(req)) {
			return accessGranted
		}
	}
	if claims.IsInGroup(req) { // group role
		return accessGranted
	}
	if claims.IsCourseTeacher(getCourseID(req)) { // teacher role in course
		return accessGranted
	}
	return "not student, group member, or teacher"
}

// checkUpdateSubmission checks if the submission is valid and if the user is a teacher in the course specified in the request.
// The [req] is expected to implement [submissionIDProvider] and optionally [courseIDProvider].
// If the request does not provide a CourseID, the course is determined from the submission's assignment.
func checkUpdateSubmission(db database.Database, req any, claims *auth.Claims) string {
	if !isValidSubmission(db, req) {
		return "invalid submission"
	}
	// Get course ID from request, or fetch it from the submission's assignment
	courseID := cmp.Or(getCourseID(req), getCourseIDFromDB(req, db))
	if claims.IsCourseTeacher(courseID) { // teacher role in course
		return accessGranted
	}
	return "not teacher"
}

// getCourseIDFromDB retrieves the course ID associated with the submission in the request.
// This is a HACK. It does not report error if db lookup failed.
// This is only here temporarily since the UpdateSubmissionRequest has been replaced with Grade, which does not have CourseID.
func getCourseIDFromDB(req any, db database.Database) uint64 {
	submissionID := getSubmissionID(req)
	sbm, err := db.GetSubmission(&qf.Submission{ID: submissionID})
	if err != nil {
		return 0
	}
	assignment, err := db.GetAssignment(&qf.Assignment{ID: sbm.GetAssignmentID()})
	if err != nil {
		return 0
	}
	return assignment.GetCourseID()
}

// methodCheckers maps each method to its corresponding access checker function.
// Each checker returns an empty string if access is granted, or a reason string if denied.
var methodCheckers = map[string]accessChecker{
	"GetUser":                  checkNone,
	"GetCourse":                checkNone,
	"GetCourses":               checkNone,
	"SubmissionStream":         checkNone, // No role required as long as the user is authenticated, i.e. has a valid token.
	"CreateEnrollment":         checkUser,
	"UpdateCourseVisibility":   checkUser,
	"UpdateUser":               checkUpdateUser,
	"GetEnrollments":           checkUserOrStudentOrTeacherOrAdmin,
	"GetSubmissions":           checkGetSubmissions,
	"GetSubmission":            checkTeacher,
	"CreateGroup":              checkGroupOrTeacher,
	"GetGroup":                 checkGroupOrTeacher,
	"GetAssignments":           checkStudentOrTeacher,
	"GetRepositories":          checkStudentOrTeacher,
	"UpdateGroup":              checkTeacher,
	"DeleteGroup":              checkTeacher,
	"GetGroupsByCourse":        checkTeacher,
	"UpdateCourse":             checkTeacher,
	"UpdateEnrollments":        checkTeacher,
	"UpdateAssignments":        checkTeacher,
	"UpdateSubmission":         checkUpdateSubmission,
	"RebuildSubmissions":       checkTeacher,
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
