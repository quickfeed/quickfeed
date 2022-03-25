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
	db     database.Database
	scms   *auth.Scms
	bh     BaseHookOptions
	runner ci.Runner
	pb.UnimplementedAutograderServiceServer
}

// NewAutograderService returns an AutograderService object.
func NewAutograderService(logger *zap.Logger, db database.Database, scms *auth.Scms, bh BaseHookOptions, runner ci.Runner) *AutograderService {
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
func (s *AutograderService) GetUser(ctx context.Context, _ *pb.Void) (*pb.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUser failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	userInfo, err := s.db.GetUserWithEnrollments(usr.GetID())
	if err != nil {
		s.logger.Errorf("GetUser failed to get user with enrollments: %v ", err)
	}
	return userInfo, nil
}

// GetUsers returns a list of all users.
// Access policy: Admin.
// Frontend note: This method is called from AdminPage.
func (s *AutograderService) GetUsers(ctx context.Context, _ *pb.Void) (*pb.Users, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUsers failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsAdmin {
		s.logger.Error("GetUsers failed: user is not admin")
		return nil, status.Error(codes.PermissionDenied, "only admin can access other users")
	}
	users, err := s.getUsers()
	if err != nil {
		s.logger.Errorf("GetUsers failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get users")
	}
	return users, nil
}

// GetUserByCourse returns the user matching the given course name and GitHub login
// specified in CourseUserRequest.
// Access policy: Admins or course teachers
func (s *AutograderService) GetUserByCourse(ctx context.Context, in *pb.CourseUserRequest) (*pb.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUserByCourse failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	userInfo, err := s.getUserByCourse(in, usr)
	if err != nil {
		s.logger.Errorf("GetUserByCourse failed: %+v", err)
		return nil, status.Error(codes.FailedPrecondition, "failed to get student information")
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
		s.logger.Errorf("UpdateUser failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !(usr.IsAdmin || usr.IsOwner(in.GetID())) {
		s.logger.Errorf("UpdateUser failed to update user %d: user is not admin or course creator", in.GetID())
		return nil, status.Error(codes.PermissionDenied, "only admin can update another user's information")
	}
	if _, err = s.updateUser(usr, in); err != nil {
		s.logger.Errorf("UpdateUser failed to update user %d: %v", in.GetID(), err)
		err = status.Error(codes.InvalidArgument, "failed to update user")
	}
	return &pb.Void{}, err
}

// IsAuthorizedTeacher checks whether current user has teacher scopes.
// Access policy: Any User.
func (s *AutograderService) IsAuthorizedTeacher(ctx context.Context, _ *pb.Void) (*pb.AuthorizationResponse, error) {
	// Currently hardcoded for github only
	_, scm, err := s.getUserAndSCM(ctx, "github")
	if err != nil {
		s.logger.Errorf("IsAuthorizedTeacher failed: scm authentication error: %v", err)
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
		s.logger.Errorf("CreateCourse failed: scm authentication error: %v", err)
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
		s.logger.Errorf("CreateCourse failed: %v", err)
		// errors informing about requested organization state will have code 9: FailedPrecondition
		// error message will be displayed to the user
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == ErrAlreadyExists || err == ErrFreePlan {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to create course")
	}
	return course, nil
}

// UpdateCourse changes the course information details.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateCourse(ctx context.Context, in *pb.Course) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCM(ctx, in.Provider)
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("UpdateCourse failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can update course")
	}

	if err = s.updateCourse(ctx, scm, in); err != nil {
		s.logger.Errorf("UpdateCourse failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to update course")
	}
	return &pb.Void{}, nil
}

// GetCourse returns course information for the given course.
// Access policy: Any User.
func (s *AutograderService) GetCourse(ctx context.Context, in *pb.CourseRequest) (*pb.Course, error) {
	courseID := in.GetCourseID()
	course, err := s.getCourse(courseID)
	if err != nil {
		s.logger.Errorf("GetCourse failed: %v", err)
		return nil, status.Error(codes.NotFound, "course not found")
	}
	return course, nil
}

// GetCourses returns a list of all courses.
// Access policy: Any User.
func (s *AutograderService) GetCourses(_ context.Context, _ *pb.Void) (*pb.Courses, error) {
	courses, err := s.getCourses()
	if err != nil {
		s.logger.Errorf("GetCourses failed: %v", err)
		return nil, status.Error(codes.NotFound, "no courses found")
	}
	return courses, nil
}

// UpdateCourseVisibility allows to edit what courses are visible in the sidebar.
// Access policy: Any User.
func (s *AutograderService) UpdateCourseVisibility(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("ChangeCourseVisibility failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsOwner(in.GetUserID()) {
		s.logger.Errorf("ChangeCourseVisibility failed: user %d attempts to update enrollment for user %d", usr.GetID(), in.GetUserID())
		return nil, status.Error(codes.PermissionDenied, "users cannot set course visibility for another users")
	}
	err = s.changeCourseVisibility(in)
	if err != nil {
		s.logger.Errorf("ChangeCourseVisibility failed: %v", err)
		err = status.Error(codes.InvalidArgument, "failed to update course visibility")
	}
	return &pb.Void{}, err
}

// CreateEnrollment enrolls a new student for the course specified in the request.
// Access policy: Any User.
func (s *AutograderService) CreateEnrollment(_ context.Context, in *pb.Enrollment) (*pb.Void, error) {
	err := s.createEnrollment(in)
	if err != nil {
		s.logger.Errorf("CreateEnrollment failed: %v", err)
		err = status.Error(codes.InvalidArgument, "failed to create enrollment")
	}
	return &pb.Void{}, err
}

// UpdateEnrollments changes status of all pending enrollments for the specified course to approved.
// If the request contains a single enrollment, it will be updated to the specified status.
// Access policy: Teacher of CourseID
func (s *AutograderService) UpdateEnrollments(ctx context.Context, in *pb.Enrollments) (*pb.Void, error) {
	user, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollments failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(user.GetID(), in.GetCourseID()) {
		s.logger.Errorf("UpdateEnrollments failed: user %d is not teacher of course %d", user.GetID(), in.GetCourseID())
		return nil, status.Error(codes.PermissionDenied, "only teachers can update enrollments")
	}

	for _, enrollment := range in.GetEnrollments() {
		if s.isCourseCreator(enrollment.CourseID, enrollment.UserID) {
			s.logger.Errorf("UpdateEnrollments failed: user %s attempted to demote course creator", user.GetName())
			return nil, status.Error(codes.PermissionDenied, "course creator cannot be demoted")
		}

		if err = s.updateEnrollment(ctx, scm, user.GetLogin(), enrollment); err != nil {
			s.logger.Errorf("UpdateEnrollments failed: %v", err)
			if contextCanceled(ctx) {
				return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
			}
			if ok, parsedErr := parseSCMError(err); ok {
				return nil, parsedErr
			}
			return nil, status.Error(codes.InvalidArgument, "failed to update enrollment")
		}
	}
	return &pb.Void{}, err
}

// GetCoursesByUser returns all courses the given user is enrolled into with the given status.
// Access policy: Any User.
func (s *AutograderService) GetCoursesByUser(_ context.Context, in *pb.EnrollmentStatusRequest) (*pb.Courses, error) {
	courses, err := s.getCoursesByUser(in)
	if err != nil {
		s.logger.Errorf("GetCoursesWithEnrollment failed: %v", err)
		return nil, status.Error(codes.NotFound, "no courses with enrollment found")
	}
	return courses, nil
}

// GetEnrollmentsByUser returns all enrollments for the given user and enrollment status with preloaded courses and groups.
// Access policy: user with userID or admin
func (s *AutograderService) GetEnrollmentsByUser(ctx context.Context, in *pb.EnrollmentStatusRequest) (*pb.Enrollments, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByUser failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if usr.GetID() != in.GetUserID() && !usr.IsAdmin {
		s.logger.Errorf("GetEnrollmentsByUser failed: current user ID: %d, but requested user ID is %d", usr.ID, in.UserID)
		return nil, status.Error(codes.PermissionDenied, "only admins can request enrollments for other users")
	}

	// get all enrollments from the db (no scm)
	enrols, err := s.getEnrollmentsByUser(in)
	if err != nil {
		s.logger.Errorf("Get enrollments for user %d failed: %v", in.GetUserID(), err)
	}
	return enrols, nil
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
// Access policy: Teacher or student of CourseID.
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isEnrolled(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("GetEnrollmentsByCourse failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can get course enrollments")
	}

	enrolls, err := s.getEnrollmentsByCourse(in)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: %v", err)
		return nil, status.Error(codes.InvalidArgument, "failed to get enrollments for given course")
	}
	return enrolls, nil
}

// GetGroup returns information about a group.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroup(ctx context.Context, in *pb.GetGroupRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroup failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	group, err := s.getGroup(in)
	if err != nil {
		s.logger.Errorf("GetGroup failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get group")
	}
	if !(group.Contains(usr) || s.isTeacher(usr.GetID(), group.GetCourseID())) {
		s.logger.Error("GetGroup failed: user is not group member or teacher")
		return nil, status.Error(codes.PermissionDenied, "only group members and teachers can access a group")
	}
	return group, nil
}

// GetGroupsByCourse returns a list of groups created for the course id in the record request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) GetGroupsByCourse(ctx context.Context, in *pb.CourseRequest) (*pb.Groups, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroups failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	courseID := in.GetCourseID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("GetGroups failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can access other groups")
	}
	groups, err := s.getGroups(in)
	if err != nil {
		s.logger.Errorf("GetGroups failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get groups")
	}
	return groups, nil
}

// GetGroupByUserAndCourse returns the group of the given student for a given course.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroupByUserAndCourse(ctx context.Context, in *pb.GroupRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroupByUserAndCourse failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	group, err := s.getGroupByUserAndCourse(in)
	if err != nil {
		if err != errUserNotInGroup {
			s.logger.Errorf("GetGroupByUserAndCourse failed: %v", err)
		}
		return nil, status.Error(codes.NotFound, "failed to get group for given user and course")
	}
	if !(group.Contains(usr) || s.isTeacher(usr.GetID(), group.GetCourseID())) {
		s.logger.Error("GetGroupByUserAndCourse failed: user is not group member or teacher")
		return nil, status.Error(codes.PermissionDenied, "only group members and teachers can access another group")
	}
	return group, nil
}

// CreateGroup creates a new group in the database.
// Access policy: Any User enrolled in course and specified as member of the group or a course teacher.
func (s *AutograderService) CreateGroup(ctx context.Context, in *pb.Group) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("CreateGroup failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isEnrolled(usr.GetID(), in.GetCourseID()) {
		s.logger.Errorf("CreateGroup failed: user %s not enrolled in course %d", usr.GetLogin(), in.GetCourseID())
		return nil, status.Error(codes.PermissionDenied, "user not enrolled in given course")
	}
	if !(in.Contains(usr) || s.isTeacher(usr.GetID(), in.GetCourseID())) {
		s.logger.Error("CreateGroup failed: user is not group member or teacher")
		return nil, status.Error(codes.PermissionDenied, "only group member or teacher can create group")
	}
	group, err := s.createGroup(in)
	if err != nil {
		s.logger.Errorf("CreateGroup failed: %v", err)
		if _, ok := status.FromError(err); ok {
			// err was already a status error; return it to client.
			return nil, err
		}
		// err was not a status error; return a generic error to client.
		return nil, status.Error(codes.InvalidArgument, "failed to create group")
	}
	return group, nil
}

// UpdateGroup updates group information.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("UpdateGroup failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can update groups")
	}
	err = s.updateGroup(ctx, scm, in)
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		if _, ok := status.FromError(err); ok {
			// err was already a status error; return it to client.
			return nil, err
		}
		// err was not a status error; return a generic error to client.
		return nil, status.Error(codes.InvalidArgument, "failed to update group")
	}
	return &pb.Void{}, nil
}

// DeleteGroup removes group record from the database.
// Access policy: Teacher of CourseID.
func (s *AutograderService) DeleteGroup(ctx context.Context, in *pb.GroupRequest) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	grp, err := s.getGroup(&pb.GetGroupRequest{GroupID: in.GetGroupID()})
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get group")
	}
	if !s.isTeacher(usr.GetID(), grp.GetCourseID()) {
		s.logger.Error("DeleteGroup failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can delete groups")
	}
	if err = s.deleteGroup(ctx, scm, in); err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(errors.Unwrap(err)); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to delete group")
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
		s.logger.Errorf("GetSubmissions failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}

	// grp may be nil if there is no group ID in request; this is fine, since the grp.Contains() returns false in this case.
	grp, _ := s.getGroup(&pb.GetGroupRequest{GroupID: in.GetGroupID()})

	// ensure that current user is teacher, enrolled admin, or the current user is owner of the submission request
	if !s.hasCourseAccess(usr.GetID(), in.GetCourseID(), func(e *pb.Enrollment) bool {
		switch e.Status {
		case pb.Enrollment_TEACHER:
			return true
		case pb.Enrollment_STUDENT:
			return usr.IsAdmin || usr.IsOwner(in.GetUserID()) || grp.Contains(usr)
		}
		return false
	}) {
		s.logger.Errorf("GetSubmissions failed: user %s is not teacher or submission author", usr.GetLogin())
		return nil, status.Error(codes.PermissionDenied, "only owner and teachers can get submissions")
	}
	s.logger.Debugf("GetSubmissions: %v", in)

	submissions, err := s.getSubmissions(in)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: %v", err)
		return nil, status.Error(codes.NotFound, "no submissions found")
	}
	return submissions, nil
}

// GetSubmissionsByCourse returns all the latest submissions
// for every individual or group course assignment for all course students/groups.
// Access policy: Admin enrolled in CourseID, Teacher of CourseID.
func (s *AutograderService) GetSubmissionsByCourse(ctx context.Context, in *pb.SubmissionsForCourseRequest) (*pb.CourseSubmissions, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetSubmissionsByCourse failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	// TODO(meling) This is a hack to give access to cmd/approvelist via the root usr admin
	// Normally, the root admin should not have access to submissions for all courses.
	//	if usr.IsAdmin {
	//		goto BYPASS
	//	}

	// ensure that current user is teacher or enrolled admin to process the submission request
	if !s.hasCourseAccess(usr.GetID(), in.GetCourseID(), func(e *pb.Enrollment) bool {
		switch e.Status {
		case pb.Enrollment_TEACHER:
			return true
		case pb.Enrollment_STUDENT:
			return usr.IsAdmin
		}
		return false
	}) {
		s.logger.Errorf("GetSubmissionsByCourse failed: user %s is not teacher or submission author", usr.GetLogin())
		return nil, status.Error(codes.PermissionDenied, "only teachers can get all lab submissions")
	}
	// BYPASS:
	s.logger.Debugf("GetSubmissionsByCourse: %v", in)

	courseLinks, err := s.getAllCourseSubmissions(in)
	if err != nil {
		s.logger.Errorf("GetSubmissionsByCourse failed: %v", err)
		return nil, status.Error(codes.NotFound, "no submissions found")
	}
	return courseLinks, nil
}

// UpdateSubmission is called to approve the given submission or to undo approval.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateSubmission(ctx context.Context, in *pb.UpdateSubmissionRequest) (*pb.Void, error) {
	if !s.isValidSubmission(in.SubmissionID) {
		s.logger.Errorf("UpdateSubmission failed: submission author has no access to the course")
		return nil, status.Error(codes.PermissionDenied, "submission author has no course access")
	}
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateSubmission failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("UpdateSubmission failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can approve submissions")
	}
	err = s.updateSubmission(in.GetCourseID(), in.GetSubmissionID(), in.GetStatus(), in.GetReleased(), in.GetScore())
	if err != nil {
		s.logger.Errorf("UpdateSubmission failed: %v", err)
		err = status.Error(codes.InvalidArgument, "failed to approve submission")
	}
	return &pb.Void{}, err
}

// RebuildSubmissions re-runs the tests for the given assignment.
// A single submission is executed again if the request specifies a submission ID
// or all submissions if the request specifies a course ID.
// Access policy: Teacher of CourseID.
func (s *AutograderService) RebuildSubmissions(ctx context.Context, in *pb.RebuildRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("RebuildSubmissions failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("RebuildSubmissions failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can rebuild all submissions")
	}
	// RebuildType can be either SubmissionID or CourseID, but not both.
	switch in.GetRebuildType().(type) {
	case *pb.RebuildRequest_SubmissionID:
		if !s.isValidSubmission(in.GetSubmissionID()) {
			s.logger.Errorf("RebuildSubmission failed: submitter has no access to the course")
			return nil, status.Error(codes.PermissionDenied, "submitter has no course access")
		}
		if _, err := s.rebuildSubmission(in); err != nil {
			s.logger.Errorf("RebuildSubmission failed: %v", err)
			return nil, status.Error(codes.InvalidArgument, "failed to rebuild submission")
		}
	case *pb.RebuildRequest_CourseID:
		if err := s.rebuildSubmissions(in); err != nil {
			s.logger.Errorf("RebuildSubmissions failed: %v", err)
			return nil, status.Error(codes.InvalidArgument, "failed to rebuild submissions")
		}
	}
	return &pb.Void{}, nil
}

// CreateBenchmark adds a new grading benchmark for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) CreateBenchmark(_ context.Context, in *pb.GradingBenchmark) (*pb.GradingBenchmark, error) {
	bm, err := s.createBenchmark(in)
	if err != nil {
		s.logger.Errorf("CreateBenchmark failed for %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to add benchmark")
	}
	return bm, nil
}

// UpdateBenchmark edits a grading benchmark for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) UpdateBenchmark(_ context.Context, in *pb.GradingBenchmark) (*pb.Void, error) {
	err := s.updateBenchmark(in)
	if err != nil {
		s.logger.Errorf("UpdateBenchmark failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update benchmark")
	}
	return &pb.Void{}, err
}

// DeleteBenchmark removes a grading benchmark
// Access policy: Teacher of CourseID
func (s *AutograderService) DeleteBenchmark(_ context.Context, in *pb.GradingBenchmark) (*pb.Void, error) {
	err := s.deleteBenchmark(in)
	if err != nil {
		s.logger.Errorf("DeleteBenchmark failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to delete benchmark")
	}
	return &pb.Void{}, err
}

// CreateCriterion adds a new grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) CreateCriterion(_ context.Context, in *pb.GradingCriterion) (*pb.GradingCriterion, error) {
	c, err := s.createCriterion(in)
	if err != nil {
		s.logger.Errorf("CreateCriterion failed for %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to add criterion")
	}
	return c, nil
}

// UpdateCriterion edits a grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) UpdateCriterion(_ context.Context, in *pb.GradingCriterion) (*pb.Void, error) {
	err := s.updateCriterion(in)
	if err != nil {
		s.logger.Errorf("UpdateCriterion failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update criterion")
	}
	return &pb.Void{}, err
}

// DeleteCriterion removes a grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *AutograderService) DeleteCriterion(_ context.Context, in *pb.GradingCriterion) (*pb.Void, error) {
	err := s.deleteCriterion(in)
	if err != nil {
		s.logger.Errorf("DeleteCriterion failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to delete criterion")
	}
	return &pb.Void{}, err
}

// CreateReview adds a new submission review
// Access policy: Teacher of CourseID
func (s *AutograderService) CreateReview(ctx context.Context, in *pb.ReviewRequest) (*pb.Review, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("CreateReview failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("CreateReview failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can add reviews")
	}
	if !usr.IsOwner(in.Review.GetReviewerID()) {
		s.logger.Errorf("CreateReview failed: current user's ID: %d, when the reviewer's ID is %d ", usr.ID, in.Review.ReviewerID)
		return nil, status.Error(codes.PermissionDenied, "failed to create review: reviewers' IDs don't match")
	}
	review, err := s.createReview(in.Review)
	if err != nil {
		s.logger.Errorf("CreateReview failed for review %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to create review")
	}
	return review, nil
}

// UpdateReview updates a submission review
// Access policy: Teacher of CourseID, Author of the given Review
func (s *AutograderService) UpdateReview(ctx context.Context, in *pb.ReviewRequest) (*pb.Review, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateReview failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("UpdateReview failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can update reviews")
	}
	if !(usr.IsOwner(in.Review.GetReviewerID()) || s.isCourseCreator(in.CourseID, usr.ID)) {
		s.logger.Errorf("UpdateReview failed: current user's ID: %d, when the original reviewer's ID is %d ", usr.ID, in.Review.ReviewerID)
		return nil, status.Error(codes.PermissionDenied, "reviews can only be updated by original authors or course creator")
	}
	review, err := s.updateReview(in.Review)
	if err != nil {
		s.logger.Errorf("UpdateReview failed for review %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update review")
	}
	return review, err
}

// UpdateSubmissions approves and/or releases all manual reviews for student submission for the given assignment
// with the given score.
// Access policy: Creator of CourseID
func (s *AutograderService) UpdateSubmissions(ctx context.Context, in *pb.UpdateSubmissionsRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateSubmissions failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isCourseCreator(in.CourseID, usr.ID) {
		s.logger.Error("UpdateSubmissions failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can update reviews")
	}

	if err = s.updateSubmissions(in); err != nil {
		s.logger.Errorf("UpdateSubmissions failed for request %+v", in)
		err = status.Error(codes.InvalidArgument, "failed to update submissions")
	}
	return &pb.Void{}, err
}

// GetReviewers returns names of all active reviewers for a student submission
// Access policy: Teacher of CourseID
func (s *AutograderService) GetReviewers(ctx context.Context, in *pb.SubmissionReviewersRequest) (*pb.Reviewers, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetReviewers failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("GetReviewers failed: user is not course creator")
		return nil, status.Error(codes.PermissionDenied, "only course creator teacher can request information about reviewers")
	}
	reviewers, err := s.getReviewers(in.SubmissionID)
	if err != nil {
		s.logger.Errorf("GetReviewers failed: error fetching from database: %v", err)
		return nil, status.Error(codes.InvalidArgument, "failed to get reviewers")
	}
	return &pb.Reviewers{Reviewers: reviewers}, err
}

// GetAssignments returns a list of all assignments for the given course.
// Access policy: Any User.
func (s *AutograderService) GetAssignments(_ context.Context, in *pb.CourseRequest) (*pb.Assignments, error) {
	courseID := in.GetCourseID()
	assignments, err := s.getAssignments(courseID)
	if err != nil {
		s.logger.Errorf("GetAssignments failed: %v", err)
		return nil, status.Error(codes.NotFound, "no assignments found for course")
	}
	return assignments, nil
}

// UpdateAssignments updates the assignments record in the database
// by fetching assignment information from the course's test repository.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateAssignments(ctx context.Context, in *pb.CourseRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: scm authentication error: %v", err)
		return nil, err
	}
	courseID := in.GetCourseID()
	if !s.isTeacher(usr.ID, courseID) {
		s.logger.Error("UpdateAssignments failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can update course assignments")
	}
	err = s.updateAssignments(courseID)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: %v", err)
		return nil, status.Error(codes.NotFound, "course not found")
	}
	return &pb.Void{}, nil
}

// GetProviders returns a list of SCM providers supported by the backend.
// Access policy: Any User.
func (s *AutograderService) GetProviders(_ context.Context, _ *pb.Void) (*pb.Providers, error) {
	providers := auth.GetProviders()
	if len(providers.GetProviders()) < 1 {
		s.logger.Error("GetProviders failed: found no enabled SCM providers")
		return nil, status.Error(codes.NotFound, "found no enabled SCM providers")
	}
	return providers, nil
}

// GetOrganization fetches a github organization by name.
// Access policy: Admin
func (s *AutograderService) GetOrganization(ctx context.Context, in *pb.OrgRequest) (*pb.Organization, error) {
	usr, scm, err := s.getUserAndSCM(ctx, "github")
	if err != nil {
		s.logger.Errorf("GetOrganization failed: scm authentication error: %v", err)
		return nil, err
	}
	if !usr.IsAdmin {
		s.logger.Error("GetOrganization failed: user is not admin")
		return nil, status.Error(codes.PermissionDenied, "only admin can access organizations")
	}
	org, err := s.getOrganization(ctx, scm, in.GetOrgName(), usr.GetLogin())
	if err != nil {
		s.logger.Errorf("GetOrganization failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == scms.ErrNotMember {
			return nil, status.Error(codes.NotFound, "organization membership not confirmed, please enable third-party access")
		}
		if err == ErrFreePlan || err == ErrAlreadyExists || err == scms.ErrNotOwner {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.NotFound, "organization not found. Please make sure that 3rd-party access is enabled for your organization")
	}
	return org, nil
}

// GetRepositories returns URL strings for repositories of given type for the given course
// Access policy: Any User.
func (s *AutograderService) GetRepositories(ctx context.Context, in *pb.URLRequest) (*pb.Repositories, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetRepositories failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}

	course, err := s.getCourse(in.GetCourseID())
	if err != nil {
		s.logger.Errorf("GetRepositories failed: course %d not found: %v", in.GetCourseID(), err)
		return nil, status.Error(codes.NotFound, "course not found")
	}

	enrol, _ := s.db.GetEnrollmentByCourseAndUser(course.GetID(), usr.GetID())

	urls := make(map[string]string)
	for _, repoType := range in.GetRepoTypes() {
		var id uint64
		switch repoType {
		case pb.Repository_USER:
			id = usr.GetID()
		case pb.Repository_GROUP:
			id = enrol.GetGroupID() // will be 0 if not enrolled in a group
		}
		repo, _ := s.getRepo(course, id, repoType)
		// for repo == nil: will result in an empty URL string, which will be ignored by the frontend
		urls[repoType.String()] = repo.GetHTMLURL()
	}
	return &pb.Repositories{URLs: urls}, nil
}

// IsEmptyRepo ensures that group repository is empty and can be deleted
// Access policy: Teacher of Course ID
func (s *AutograderService) IsEmptyRepo(ctx context.Context, in *pb.RepositoryRequest) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: scm authentication error: %v", err)
		return nil, err
	}

	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("IsEmptyRepo failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can access repository info")
	}

	if err := s.isEmptyRepo(ctx, scm, in); err != nil {
		s.logger.Errorf("IsEmptyRepo failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.FailedPrecondition, "group repository does not exist or not empty")
	}
	return &pb.Void{}, nil
}
