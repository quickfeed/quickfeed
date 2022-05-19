package interceptors

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	role      int
	roles     []role
	idRequest interface {
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

// If there are several roles that can call a method, a role with least privilege must come first
// If method is not in the map, there is no restrictions to call it
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
	"IsEmptyRepo":             {teacher},
	"GetGroupsByCourse":       {teacher},
	"UpdateCourse":            {teacher},
	"UpdateEnrollments":       {teacher},
	"UpdateSubmission":        {teacher},
	"RebuildSubmissions":      {teacher},
	"CreateBenchmark":         {teacher},
	"UpdateBenchmark":         {teacher},
	"DeleteBenchmark":         {teacher},
	"CreateCriterion":         {teacher},
	"UpdateCriterion":         {teacher},
	"DeleteCriterion":         {teacher},
	"CreateReview":            {teacher},
	"UpdateReview":            {teacher}, // TODO(vera): also had "owner" role, but looks excessive?
	"UpdateSubmissions":       {teacher},
	"GetReviewers":            {teacher},
	"UpdateAssignments":       {teacher},
	"GetSubmissionsByCourse":  {teacher, courseAdmin},
	"GetUserByCourse":         {teacher, admin},
	"GetOrganization":         {admin},
	"CreateCourse":            {admin},
}

func logError(logger *zap.SugaredLogger, format string, a ...interface{}) {
	logger.Error("AccessControl: " + fmt.Sprintf(format, a...))
}

func AccessControl(logger *zap.SugaredLogger, db database.Database, tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		token, err := GetFromMetadata(ctx, "token", "")
		if err != nil {
			logError(logger, "missing token in metadata: %s", err)
			return nil, ErrAccessDenied
		}
		claims, err := tokens.GetClaims(token)
		if err != nil {
			logError(logger, "failed to get user claims: %s", err)
			return nil, ErrAccessDenied
		}
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		logger.Debugf("ACCESS CONTROL: user %d calls method %s", claims.UserID, method) // tmp
		roles, ok := access[method]
		if ok {
			for _, role := range roles {
				switch role {
				case user:
					if m, ok := req.(idRequest); ok {
						id := m.FetchID("user")
						if id == claims.UserID {
							return handler(ctx, req)
						}
					} else {
						logger.Debugf("Method %s does not implement FetchID method", method)
					}
				case group:
					switch method {
					// Group memebers can access own group.
					case "GetGroup":
						groupID := req.(*pb.GetGroupRequest).GetGroupID()
						group, err := db.GetGroup(groupID)
						if err != nil {
							logError(logger, "no group with ID %d: %s", groupID, err)
							return nil, ErrAccessDenied
						}
						if group.Contains(&pb.User{ID: claims.UserID}) {
							return handler(ctx, req)
						}
					case "GetGroupByUserAndCourse":
						if claims.UserID == req.(*pb.GroupRequest).UserID {
							return handler(ctx, req)
						}
					// User creating a new group must be a group member
					// and enrolled in the course.
					case "CreateGroup":
						group := req.(*pb.Group)
						enrolled := hasCourseAccess(db, group.GetCourseID(), claims.UserID, pb.Enrollment_STUDENT)
						groupMember := group.Contains(&pb.User{ID: claims.UserID})
						if enrolled && groupMember {
							return handler(ctx, req)
						}
					// Group members can access group submissions.
					case "GetSubmissions":
						groupID := req.(*pb.SubmissionRequest).GetGroupID()
						// If groupID is zero, the user is not in a group or
						// only individual submissions are requested.
						if groupID > 0 {
							group, err := db.GetGroup(groupID)
							if err != nil {
								logError(logger, "no group with ID %d: %s", groupID, err)
								return nil, ErrAccessDenied
							}
							if group.Contains(&pb.User{ID: claims.UserID}) {
								return handler(ctx, req)
							}
						}
					}
				case student:
					if m, ok := req.(idRequest); ok {
						courseID := m.FetchID("course")
						if hasCourseAccess(db, courseID, claims.UserID, pb.Enrollment_STUDENT) {
							return handler(ctx, req)
						}
					} else {
						logger.Debugf("Method %s does not implement FetchID method", method)
					}
				case courseAdmin:
					if claims.Admin {
						if m, ok := req.(idRequest); ok {
							courseID := m.FetchID("course")
							if hasCourseAccess(db, courseID, claims.UserID, pb.Enrollment_TEACHER) {
								return handler(ctx, req)
							}
						} else {
							logger.Debugf("Method %s does not implement FetchID method", method)
						}
					}
				case teacher:
					var courseID uint64
					if m, ok := req.(idRequest); ok {
						courseID = m.FetchID("course")
					} else {
						logger.Debugf("Method %s does not implement FetchID method", method)
					}
					switch method {
					// Request here does not have a course ID
					// TODO(vera): add one?
					case "GetGroup":
						groupID := req.(*pb.GetGroupRequest).GetGroupID()
						group, err := db.GetGroup(groupID)
						if err != nil {
							logError(logger, "no group with ID %d: %s", groupID, err)
							return nil, ErrAccessDenied
						}
						courseID = group.GetCourseID()
					case "GetUserByCourse":
						courseCode := req.(*pb.CourseUserRequest).GetCourseCode()
						courseYear := req.(*pb.CourseUserRequest).GetCourseYear()
						query := &pb.Course{
							Code: courseCode,
							Year: courseYear,
						}
						course, err := db.GetCourse(query, false)
						if err != nil {
							logError(logger, "no course with code %s and year %d: %s", courseCode, courseYear, err)
							return nil, ErrAccessDenied
						}
						courseID = course.GetID()
					}
					if hasCourseAccess(db, courseID, claims.UserID, pb.Enrollment_TEACHER) {
						return handler(ctx, req)
					}
				case admin:
					if claims.Admin {
						return handler(ctx, req)
					}
				default: // tmp
					logger.Debugf("Unknown access role: %s", role)
				}
			}
			logError(logger, "user %d unauthorized to call method %s", claims.UserID, method)
			return nil, ErrAccessDenied
		}
		logger.Debugf("Access control (%s) took %v", method, time.Since(start))
		return handler(ctx, req)
	}
}
