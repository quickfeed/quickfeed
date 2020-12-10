package web

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	scms "github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web/auth"
)

// AutograderService holds references to the database and
// other shared data structures.
type AutograderService struct {
	logger *zap.SugaredLogger
	db     *database.GormDB
	scms   *auth.Scms
	bh     BaseHookOptions
	runner ci.Runner
}

// NewAutograderService returns an AutograderService object.
func NewAutograderService(logger *zap.Logger, db *database.GormDB, scms *auth.Scms, bh BaseHookOptions, runner ci.Runner) *AutograderService {
	return &AutograderService{
		logger: logger.Sugar(),
		db:     db,
		scms:   scms,
		bh:     bh,
		runner: runner,
	}
}

// GetUser will return current user with active course enrollments
// to use in separating teacher and admin roles
// Access policy: everyone
func (s *AutograderService) GetUser(ctx context.Context, in *pb.Void) (*pb.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUser failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	userInfo, err := s.db.GetUserWithEnrollments(usr.GetID())
	if err != nil {
		s.logger.Errorf("GetUser failed to get user with enrollments: %w ", err)
	}
	return userInfo, nil

}

// GetUsers returns a list of all users.
// Access policy: Admin.
// Frontend note: This method is called from AdminPage.
func (s *AutograderService) GetUsers(ctx context.Context, in *pb.Void) (*pb.Users, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUsers failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsAdmin {
		s.logger.Error("GetUsers failed: user is not admin")
		return nil, status.Errorf(codes.PermissionDenied, "only admin can access other users")
	}
	users, err := s.getUsers()
	if err != nil {
		s.logger.Errorf("GetUsers failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "failed to get users")
	}
	return users, nil
}

// GetUserByCourse returns the user matching the given course name and GitHub login
// specified in CourseUserRequest.
// Access policy: Admins or course teachers
func (s *AutograderService) GetUserByCourse(ctx context.Context, in *pb.CourseUserRequest) (*pb.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUserByCourse failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	userInfo, err := s.getUserByCourse(in, usr)
	if err != nil {
		s.logger.Errorf("GetUserByCourse failed: %+v", err)
		return nil, status.Errorf(codes.FailedPrecondition, "failed to get student information")
	}
	return userInfo, nil
}

// UpdateUser updates the current users's information and returns the updated user.
// This function can also promote a user to admin or demote a user.
// Access policy: Admin can update other users's information and promote to Admin;
// Current User if Owner can update its own information.
func (s *AutograderService) UpdateUser(ctx context.Context, in *pb.User) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateUser failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !(usr.IsAdmin || usr.IsOwner(in.GetID())) {
		s.logger.Errorf("UpdateUser failed to update user %d: user is not admin or course creator", in.GetID())
		return nil, status.Errorf(codes.PermissionDenied, "only admin can update another user's information")
	}
	if _, err = s.updateUser(usr, in); err != nil {
		s.logger.Errorf("UpdateUser failed to update user %d: %w", in.GetID(), err)
		err = status.Errorf(codes.InvalidArgument, "failed to update user")
	}
	return &pb.Void{}, err
}

// IsAuthorizedTeacher checks whether current user has teacher scopes.
// Access policy: Any User.
func (s *AutograderService) IsAuthorizedTeacher(ctx context.Context, in *pb.Void) (*pb.AuthorizationResponse, error) {
	// Currently hardcoded for github only
	_, scm, err := s.getUserAndSCM(ctx, "github")
	if err != nil {
		s.logger.Errorf("IsAuthorizedTeacher failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	return &pb.AuthorizationResponse{
		IsAuthorized: hasTeacherScopes(ctx, scm),
	}, nil
}

// CreateCourse creates a new course.
// Access policy: Admin.
func (s *AutograderService) CreateCourse(ctx context.Context, in *pb.Course) (*pb.Course, error) {
	usr, scm, err := s.getUserAndSCM(ctx, in.Provider)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsAdmin {
		s.logger.Error("CreateCourse failed: user is not admin")
		return nil, status.Error(codes.PermissionDenied, "user must be admin to create course")
	}

	// make sure that the current user is set as course creator
	in.CourseCreatorID = usr.GetID()
	course, err := s.createCourse(ctx, scm, in)
	if err != nil {
		s.logger.Error("CreateCourse failed: ", err.Error())
		// errors informing about requested organization state will have code 9: FailedPrecondition
		// error message will be displayed to the user
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == ErrAlreadyExists || err == ErrFreePlan {
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Errorf(codes.InvalidArgument, "failed to create course")
	}
	return course, nil
}

// UpdateCourse changes the course information details.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateCourse(ctx context.Context, in *pb.Course) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCM(ctx, in.Provider)
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("UpdateCourse failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update course")
	}

	if err = s.updateCourse(ctx, scm, in); err != nil {
		s.logger.Errorf("UpdateCourse failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Errorf(codes.InvalidArgument, "failed to update course")
	}
	return &pb.Void{}, nil
}

// GetCourse returns course information for the given course.
// Access policy: Any User.
func (s *AutograderService) GetCourse(ctx context.Context, in *pb.CourseRequest) (*pb.Course, error) {
	courseID := in.GetCourseID()
	course, err := s.getCourse(courseID)
	if err != nil {
		s.logger.Errorf("GetCourse failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "course not found")
	}
	return course, nil
}

// GetCourses returns a list of all courses.
// Access policy: Any User.
func (s *AutograderService) GetCourses(ctx context.Context, in *pb.Void) (*pb.Courses, error) {
	courses, err := s.getCourses()
	if err != nil {
		s.logger.Errorf("GetCourses failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "no courses found")
	}
	return courses, nil
}

// UpdateCourseVisibility allows to edit what courses are visible in the sidebar.
// Access policy: Any User.
func (s *AutograderService) UpdateCourseVisibility(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("ChangeCourseVisibility failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsOwner(in.GetUserID()) {
		s.logger.Errorf("ChangeCourseVisibility failed: user %d attempts to update enrollment for user %s", usr.GetID(), in.GetUserID())
		return nil, status.Errorf(codes.PermissionDenied, "users cannot set course visibility for another users")
	}
	err = s.changeCourseVisibility(in)
	if err != nil {
		s.logger.Errorf("ChangeCourseVisibility failed: %w", err)
		err = status.Errorf(codes.InvalidArgument, "failed to update course visibility")
	}
	return &pb.Void{}, err
}

// CreateEnrollment enrolls a new student for the course specified in the request.
// Access policy: Any User.
func (s *AutograderService) CreateEnrollment(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	err := s.createEnrollment(in)
	if err != nil {
		s.logger.Errorf("CreateEnrollment failed: %w", err)
		err = status.Error(codes.InvalidArgument, "failed to create enrollment")
	}
	return &pb.Void{}, err
}

// UpdateEnrollment updates the enrollment status of a student as specified in the request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateEnrollment(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollment failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("UpdateEnrollment failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update enrollment status")
	}
	if s.isCourseCreator(in.CourseID, in.UserID) {
		s.logger.Errorf("UpdateEnrollment failed: user %s attempted to demote course creator", usr.GetName())
		return nil, status.Errorf(codes.PermissionDenied, "course creator cannot be demoted")
	}
	err = s.updateEnrollment(ctx, scm, usr.Login, in)
	if err != nil {
		s.logger.Errorf("UpdateEnrollment failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to update enrollment")
	}
	return &pb.Void{}, nil
}

// UpdateEnrollments changes status of all pending enrollments for the given course to approved
// Access policy: Teacher of CourseID
func (s *AutograderService) UpdateEnrollments(ctx context.Context, in *pb.CourseRequest) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollments failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("UpdateEnrollments failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update enrollment status")
	}
	err = s.updateEnrollments(ctx, scm, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollments failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		err = status.Error(codes.InvalidArgument, "failed to update pending enrollments")
	}
	return &pb.Void{}, err
}

// GetCoursesByUser returns all courses the given user is enrolled into with the given status.
// Access policy: Any User.
func (s *AutograderService) GetCoursesByUser(ctx context.Context, in *pb.EnrollmentStatusRequest) (*pb.Courses, error) {
	courses, err := s.getCoursesByUser(in)
	if err != nil {
		s.logger.Errorf("GetCoursesWithEnrollment failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "no courses with enrollment found")
	}
	return courses, nil
}

// GetEnrollmentsByUser returns all enrollments for the given user and enrollment status with preloaded courses and groups.
// Access policy: user with userID or admin
func (s *AutograderService) GetEnrollmentsByUser(ctx context.Context, in *pb.EnrollmentStatusRequest) (*pb.Enrollments, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByUser failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if usr.GetID() != in.GetUserID() && !usr.IsAdmin {
		s.logger.Errorf("GetEnrollmentsByUser failed: current user ID: %d, but requested user ID is %d", usr.ID, in.UserID)
		return nil, status.Errorf(codes.PermissionDenied, "only admins can request enrollments for other users")
	}

	// get all enrollments from the db (no scm)
	enrols, err := s.getEnrollmentsByUser(in)
	if err != nil {
		s.logger.Errorf("Get enrollments for user %d failed: %s", in.GetUserID(), err)
	}
	return enrols, nil
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
// Access policy: Teacher or student of CourseID.
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isEnrolled(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("GetEnrollmentsByCourse failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can get course enrollments")
	}

	enrolls, err := s.getEnrollmentsByCourse(in)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: %w", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to get enrollments for given course")
	}
	return enrolls, nil
}

// GetGroup returns information about a group.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroup(ctx context.Context, in *pb.GetGroupRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroup failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	group, err := s.getGroup(in)
	if err != nil {
		s.logger.Errorf("GetGroup failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "failed to get group")
	}
	if !(group.Contains(usr) || s.isTeacher(usr.GetID(), group.GetCourseID())) {
		s.logger.Error("GetGroup failed: user is not group member or teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only group members and teachers can access a group")
	}
	return group, nil
}

// GetGroupsByCourse returns a list of groups created for the course id in the record request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) GetGroupsByCourse(ctx context.Context, in *pb.CourseRequest) (*pb.Groups, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroups failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	courseID := in.GetCourseID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("GetGroups failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can access other groups")
	}
	groups, err := s.getGroups(in)
	if err != nil {
		s.logger.Errorf("GetGroups failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "failed to get groups")
	}
	return groups, nil
}

// GetGroupByUserAndCourse returns the group of the given student for a given course.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroupByUserAndCourse(ctx context.Context, in *pb.GroupRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroupByUserAndCourse failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	group, err := s.getGroupByUserAndCourse(in)
	if err != nil {
		if err != ErrUserNotInGroup {
			s.logger.Errorf("GetGroupByUserAndCourse failed: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "failed to get group for given user and course")
	}
	if !(group.Contains(usr) || s.isTeacher(usr.GetID(), group.GetCourseID())) {
		s.logger.Error("GetGroupByUserAndCourse failed: user is not group member or teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only group members and teachers can access another group")
	}
	return group, nil
}

// CreateGroup creates a new group in the database.
// Access policy: Any User enrolled in course and specified as member of the group or a course teacher.
func (s *AutograderService) CreateGroup(ctx context.Context, in *pb.Group) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("CreateGroup failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isEnrolled(usr.GetID(), in.GetCourseID()) {
		s.logger.Errorf("CreateGroup failed: user %s not enrolled in course %d", usr.GetLogin(), in.GetCourseID())
		return nil, status.Errorf(codes.PermissionDenied, "user not enrolled in given course")
	}
	if !(in.Contains(usr) || s.isTeacher(usr.GetID(), in.GetCourseID())) {
		s.logger.Error("CreateGroup failed: user is not group member or teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only group member or teacher can create group")
	}
	group, err := s.createGroup(in)
	if err != nil {
		if err == ErrGroupNameDuplicate {
			return nil, err
		}
		s.logger.Errorf("CreateGroup failed: %w", err)
		return nil, status.Error(codes.InvalidArgument, "failed to create group")
	}
	return group, nil
}

// UpdateGroup updates group information.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("UpdateGroup failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update groups")
	}
	err = s.updateGroup(ctx, scm, in)
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to update group")
	}
	return &pb.Void{}, nil
}

// DeleteGroup removes group record from the database.
// Access policy: Teacher of CourseID.
func (s *AutograderService) DeleteGroup(ctx context.Context, in *pb.GroupRequest) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	grp, err := s.getGroup(&pb.GetGroupRequest{GroupID: in.GetGroupID()})
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "failed to get group")
	}
	if !s.isTeacher(usr.GetID(), grp.GetCourseID()) {
		s.logger.Error("DeleteGroup failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can delete groups")
	}
	if err = s.deleteGroup(ctx, scm, in); err != nil {
		s.logger.Errorf("DeleteGroup failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(errors.Unwrap(err)); ok {
			return nil, parsedErr
		}
		return nil, status.Errorf(codes.InvalidArgument, "failed to delete group")
	}
	return &pb.Void{}, nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
// Access policy:
// Admin enrolled in CourseID,
// Current User if Owner of submission,
// Current User if member of group for group submission,
// Teacher of CourseID.
func (s *AutograderService) GetSubmissions(ctx context.Context, in *pb.SubmissionRequest) (*pb.Submissions, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}

	// grp may be nil if there is no group ID in request; this is fine, since the grp.Contains() returns false in this case.
	grp, _ := s.getGroup(&pb.GetGroupRequest{GroupID: in.GetGroupID()})

	// ensure that current user is teacher, enrolled admin, or the current user is owner of the submission request
	if !s.hasCourseAccess(usr.GetID(), in.GetCourseID(), func(e *pb.Enrollment) bool {
		return e.Status == pb.Enrollment_TEACHER || (usr.GetIsAdmin() && e.Status == pb.Enrollment_STUDENT) ||
			(e.Status == pb.Enrollment_STUDENT && (usr.IsOwner(in.GetUserID()) || grp.Contains(usr)))
	}) {
		s.logger.Error("GetSubmissions failed: user is not teacher or submission author")
		return nil, status.Errorf(codes.PermissionDenied, "only owner and teachers can get submissions")
	}
	submissions, err := s.getSubmissions(in)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "no submissions found")
	}
	return submissions, nil
}

// GetSubmissionsByCourse returns all the latest submissions
// for every individual or group course assignment for all course students/groups.
// Access policy: Admin enrolled in CourseID, Teacher of CourseID.
func (s *AutograderService) GetSubmissionsByCourse(ctx context.Context, in *pb.SubmissionsForCourseRequest) (*pb.CourseSubmissions, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetCourseLabSubmissions failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !(s.isTeacher(usr.GetID(), in.GetCourseID()) || usr.IsAdmin && s.isEnrolled(usr.GetID(), in.GetCourseID())) {
		s.logger.Errorf("GetCourseLabSubmissions failed: user %s is not teacher or submission author", usr.GetLogin())
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can get all lab submissions")
	}
	s.logger.Debugf("GetCourseLabSubmissions: %v", in)

	courseLinks, err := s.getAllCourseSubmissions(in)
	if err != nil {
		s.logger.Errorf("GetCourseLabSubmissions failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "no submissions found")
	}
	return courseLinks, nil
}

// UpdateSubmission is called to approve the given submission or to undo approval.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateSubmission(ctx context.Context, in *pb.UpdateSubmissionRequest) (*pb.Void, error) {
	if !s.isValidSubmission(in.SubmissionID) {
		s.logger.Errorf("UpdateSubmission failed: submission author has no access to the course")
		return nil, status.Errorf(codes.PermissionDenied, "submission author has no course access")
	}
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateSubmission failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("UpdateSubmission failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can approve submissions")
	}
	err = s.updateSubmission(in.GetCourseID(), in.GetSubmissionID(), in.GetStatus(), in.GetReleased(), in.GetScore())
	if err != nil {
		s.logger.Errorf("UpdateSubmission failed: %w", err)
		err = status.Errorf(codes.InvalidArgument, "failed to approve submission")
	}
	return &pb.Void{}, err
}

// RebuildSubmission rebuilds the submission with the given ID
func (s *AutograderService) RebuildSubmission(ctx context.Context, in *pb.RebuildRequest) (*pb.Submission, error) {
	if !s.isValidSubmission(in.GetSubmissionID()) {
		s.logger.Errorf("ApproveSubmission failed: submitter has no access to the course")
		return nil, status.Errorf(codes.PermissionDenied, "submitter has no course access")
	}
	submission, err := s.rebuildSubmission(ctx, in)
	if err != nil {
		return nil, err
	}
	return submission, nil
}

// CreateBenchmark adds a new grading benchmark for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) CreateBenchmark(ctx context.Context, in *pb.GradingBenchmark) (*pb.GradingBenchmark, error) {
	bm, err := s.createBenchmark(in)
	if err != nil {
		s.logger.Errorf("CreateBenchmark failed for %+v: %s", in, err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to add benchmark")
	}
	return bm, nil
}

// UpdateBenchmark edits a grading benchmark for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) UpdateBenchmark(ctx context.Context, in *pb.GradingBenchmark) (*pb.Void, error) {
	err := s.updateBenchmark(in)
	if err != nil {
		s.logger.Errorf("UpdateBenchmark failed for %+v: %s", in, err)
		err = status.Errorf(codes.InvalidArgument, "failed to update benchmark")
	}
	return &pb.Void{}, err
}

// DeleteBenchmark removes a grading benchmark
// Access policy: Teacher of CourseID
func (s *AutograderService) DeleteBenchmark(ctx context.Context, in *pb.GradingBenchmark) (*pb.Void, error) {
	err := s.deleteBenchmark(in)
	if err != nil {
		s.logger.Errorf("DeleteBenchmark failed for %+v: %s", in, err)
		err = status.Errorf(codes.InvalidArgument, "failed to delete benchmark")
	}
	return &pb.Void{}, err
}

// CreateCriterion adds a new grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) CreateCriterion(ctx context.Context, in *pb.GradingCriterion) (*pb.GradingCriterion, error) {
	c, err := s.createCriterion(in)
	if err != nil {
		s.logger.Errorf("CreateCriterion failed for %+v: %s", in, err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to add criterion")
	}
	return c, nil
}

// UpdateCriterion edits a grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) UpdateCriterion(ctx context.Context, in *pb.GradingCriterion) (*pb.Void, error) {
	err := s.updateCriterion(in)
	if err != nil {
		s.logger.Errorf("UpdateCriterion failed for %+v: %s", in, err)
		err = status.Errorf(codes.InvalidArgument, "failed to update criterion")
	}
	return &pb.Void{}, err
}

// DeleteCriterion removes a grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) DeleteCriterion(ctx context.Context, in *pb.GradingCriterion) (*pb.Void, error) {
	err := s.deleteCriterion(in)
	if err != nil {
		s.logger.Errorf("DeleteCriterion failed for %+v: %s", in, err)
		err = status.Errorf(codes.InvalidArgument, "failed to delete criterion")
	}
	return &pb.Void{}, err
}

// LoadCriteria loads grading criteria for an assignment from a json file
// Access policy: Teacher of CourseID
func (s *AutograderService) LoadCriteria(ctx context.Context, in *pb.LoadCriteriaRequest) (*pb.Benchmarks, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("LoadCriteria failed: scm authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("LoadCriteria failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can load grading criteria")
	}

	benchmarks, err := s.loadCriteria(ctx, scm, in)
	if err != nil {
		s.logger.Errorf("LoadCriteria failed for course %d and assignment %d: %s", in.CourseID, in.AssignmentID, err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to load grading criteria for assignment")
	}
	return &pb.Benchmarks{Benchmarks: benchmarks}, nil
}

// CreateReview adds a new submission review
// Access policy: Teacher of CourseID
func (s *AutograderService) CreateReview(ctx context.Context, in *pb.ReviewRequest) (*pb.Review, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("CreateReview failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("CreateReview failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can add reviews")
	}
	if !usr.IsOwner(in.Review.GetReviewerID()) {
		s.logger.Errorf("CreateReview failed: current user's ID: %d, when the reviewer's ID is %d ", usr.ID, in.Review.ReviewerID)
		return nil, status.Errorf(codes.PermissionDenied, "failed to create review: reviewers' IDs don't match")
	}
	if err := in.Review.MakeReviewString(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create review: parsing error")
	}
	review, err := s.createReview(in.Review)
	if err != nil {
		s.logger.Errorf("CreateReview failed for review %+v: %s", in, err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create review")
	}
	return review, nil
}

// UpdateReview updates a submission review
// Access policy: Teacher of CourseID, Author of the given Review
func (s *AutograderService) UpdateReview(ctx context.Context, in *pb.ReviewRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateReview failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("UpdateReview failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update reviews")
	}
	if !usr.IsOwner(in.Review.GetReviewerID()) {
		s.logger.Errorf("UpdateReview failed: current user's ID: %d, when the original reviewer's ID is %d ", usr.ID, in.Review.ReviewerID)
		return nil, status.Errorf(codes.PermissionDenied, "reviews can only be updated by original authors")
	}
	if err := in.Review.MakeReviewString(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create review: parsing error")
	}
	if err = s.updateReview(in.Review); err != nil {
		s.logger.Errorf("UpdateReview failed for review %+v: %s", in, err)
		err = status.Errorf(codes.InvalidArgument, "failed to update review")
	}
	return &pb.Void{}, err
}

// UpdateSubmissions approves and/or releases manual reviews for student submission for the given assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) UpdateSubmissions(ctx context.Context, in *pb.UpdateSubmissionsRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("ReleaseAll failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("ReleaseAll failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update reviews")
	}

	if err = s.updateSubmissions(in); err != nil {
		s.logger.Errorf("UpdateSubmissions failed for request %+v", in)
		err = status.Errorf(codes.InvalidArgument, "failed to update submissions")
	}
	return &pb.Void{}, err
}

// GetReviewers returns names of all active reviewers for a student submission
// Access policy: Teacher of CourseID
func (s *AutograderService) GetReviewers(ctx context.Context, in *pb.SubmissionReviewersRequest) (*pb.Reviewers, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetReviewers failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("GetReviewers failed: user is not course creator")
		return nil, status.Errorf(codes.PermissionDenied, "only course creator teacher can request information about reviewers")
	}
	reviewers, err := s.getReviewers(in.SubmissionID)
	if err != nil {
		s.logger.Errorf("GetReviewers failed: error fetching from database: %s", err.Error)
		return nil, status.Errorf(codes.InvalidArgument, "failed to get reviewers")
	}
	return &pb.Reviewers{Reviewers: reviewers}, err
}

// GetAssignments returns a list of all assignments for the given course.
// Access policy: Any User.
func (s *AutograderService) GetAssignments(ctx context.Context, in *pb.CourseRequest) (*pb.Assignments, error) {
	courseID := in.GetCourseID()
	assignments, err := s.getAssignments(courseID)
	if err != nil {
		s.logger.Errorf("GetAssignments failed: %w", err)
		return nil, status.Errorf(codes.NotFound, "no assignments found for course")
	}
	return assignments, nil
}

// UpdateAssignments updates the assignments record in the database
// by fetching assignment information from the course's test repository.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateAssignments(ctx context.Context, in *pb.CourseRequest) (*pb.Void, error) {
	courseID := in.GetCourseID()
	usr, scm, err := s.getUserAndSCMForCourse(ctx, courseID)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: scm authentication error: %w", err)
		return nil, err
	}
	if !s.isTeacher(usr.ID, courseID) {
		s.logger.Error("UpdateAssignments failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update course assignments")
	}
	err = s.updateAssignments(ctx, scm, courseID)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Errorf(codes.InvalidArgument, "failed to update course assignments")
	}
	return &pb.Void{}, nil
}

// GetProviders returns a list of SCM providers supported by the backend.
// Access policy: Any User.
func (s *AutograderService) GetProviders(ctx context.Context, in *pb.Void) (*pb.Providers, error) {
	providers := auth.GetProviders()
	if len(providers.GetProviders()) < 1 {
		s.logger.Error("GetProviders failed: found no enabled SCM providers")
		return nil, status.Errorf(codes.NotFound, "found no enabled SCM providers")
	}
	return providers, nil
}

// GetOrganization fetches a github organization by name.
// Access policy: Admin
func (s *AutograderService) GetOrganization(ctx context.Context, in *pb.OrgRequest) (*pb.Organization, error) {
	usr, scm, err := s.getUserAndSCM(ctx, "github")
	if err != nil {
		s.logger.Errorf("GetOrganization failed: scm authentication error: %w", err)
		return nil, err
	}
	if !usr.IsAdmin {
		s.logger.Error("GetOrganization failed: user is not admin")
		return nil, status.Errorf(codes.PermissionDenied, "only admin can access organizations")
	}
	org, err := s.getOrganization(ctx, scm, in.GetOrgName(), usr.GetLogin())
	if err != nil {
		s.logger.Errorf("GetOrganization failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == scms.ErrNotMember {
			return nil, status.Errorf(codes.NotFound, "organization membership not confirmed, please enable third-party access")
		}
		if err == ErrFreePlan || err == ErrAlreadyExists || err == scms.ErrNotOwner {
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Errorf(codes.NotFound, "organization not found. Please make sure that 3rd-party access is enabled for your organization")
	}
	return org, nil
}

// GetRepositories returns URL strings for repositories of given type for the given course
// Access policy: Any User.
func (s *AutograderService) GetRepositories(ctx context.Context, in *pb.URLRequest) (*pb.Repositories, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetRepositories failed: authentication error: %w", err)
		return nil, ErrInvalidUserInfo
	}
	var urls = make(map[string]string)
	for _, repoType := range in.GetRepoTypes() {
		repo, _ := s.getRepositoryURL(usr, in.GetCourseID(), repoType)
		// we do not care if some repo was not found, this will append an empty url string in that case
		// frontend will take care of the rest
		urls[repoType.String()] = repo
	}
	return &pb.Repositories{URLs: urls}, nil
}

// IsEmptyRepo ensures that group repository is empty and can be deleted
// Access policy: Teacher of Course ID
func (s *AutograderService) IsEmptyRepo(ctx context.Context, in *pb.RepositoryRequest) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: scm authentication error: %w", err)
		return nil, err
	}

	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("IsEmptyRepo failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can access repository info")
	}

	if err := s.isEmptyRepo(ctx, scm, in); err != nil {
		s.logger.Errorf("IsEmptyRepo failed: %w", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Errorf(codes.FailedPrecondition, "group repository does not exist or not empty")
	}
	return &pb.Void{}, nil
}
