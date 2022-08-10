package web

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// QuickFeedService holds references to the database and
// other shared data structures.
type QuickFeedService struct {
	logger *zap.SugaredLogger
	db     database.Database
	scmMgr *scm.Manager
	bh     BaseHookOptions
	runner ci.Runner
	qf.UnimplementedQuickFeedServiceServer
}

// NewQuickFeedService returns a QuickFeedService object.
func NewQuickFeedService(logger *zap.Logger, db database.Database, mgr *scm.Manager, bh BaseHookOptions, runner ci.Runner) *QuickFeedService {
	return &QuickFeedService{
		logger: logger.Sugar(),
		db:     db,
		scmMgr: mgr,
		bh:     bh,
		runner: runner,
	}
}

// GetUser will return current user with active course enrollments
// to use in separating teacher and admin roles
// Access policy: everyone
func (s *QuickFeedService) GetUser(ctx context.Context, _ *qf.Void) (*qf.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUser failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	userInfo, err := s.db.GetUserWithEnrollments(usr.GetID())
	if err != nil {
		s.logger.Errorf("GetUser failed to get user with enrollments: %v ", err)
		return nil, ErrInvalidUserInfo
	}
	return userInfo, nil
}

// GetUsers returns a list of all users.
// Access policy: Admin.
// Frontend note: This method is called from AdminPage.
func (s *QuickFeedService) GetUsers(ctx context.Context, _ *qf.Void) (*qf.Users, error) {
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
func (s *QuickFeedService) GetUserByCourse(ctx context.Context, in *qf.CourseUserRequest) (*qf.User, error) {
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
func (s *QuickFeedService) UpdateUser(ctx context.Context, in *qf.User) (*qf.Void, error) {
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
	return &qf.Void{}, err
}

// CreateCourse creates a new course.
// Access policy: Admin.
func (s *QuickFeedService) CreateCourse(ctx context.Context, in *qf.Course) (*qf.Course, error) {
	// TODO(vera): Getting a current user will be unnecessary with the new access control, this
	// is why I leave it as a separate, repeating method.
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: user authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsAdmin {
		s.logger.Error("CreateCourse failed: user is not admin")
		return nil, status.Error(codes.PermissionDenied, "user must be admin to create course")
	}
	scmClient, err := s.getSCM(ctx, in.OrganizationPath)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: could not create scm client for the course %s: %v", in.Name, err)
		return nil, ErrMissingInstallation
	}
	// make sure that the current user is set as course creator
	in.CourseCreatorID = usr.GetID()
	course, err := s.createCourse(ctx, scmClient, in)
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
func (s *QuickFeedService) UpdateCourse(ctx context.Context, in *qf.Course) (*qf.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	scmClient, err := s.getSCM(ctx, in.OrganizationPath)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: could not create scm client for the course %s: %v", in.Name, err)
		return nil, ErrMissingInstallation
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("UpdateCourse failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can update course")
	}

	if err = s.updateCourse(ctx, scmClient, in); err != nil {
		s.logger.Errorf("UpdateCourse failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to update course")
	}
	return &qf.Void{}, nil
}

// GetCourse returns course information for the given course.
// Access policy: Any User.
func (s *QuickFeedService) GetCourse(_ context.Context, in *qf.CourseRequest) (*qf.Course, error) {
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
func (s *QuickFeedService) GetCourses(_ context.Context, _ *qf.Void) (*qf.Courses, error) {
	courses, err := s.getCourses()
	if err != nil {
		s.logger.Errorf("GetCourses failed: %v", err)
		return nil, status.Error(codes.NotFound, "no courses found")
	}
	return courses, nil
}

// UpdateCourseVisibility allows to edit what courses are visible in the sidebar.
// Access policy: Any User.
func (s *QuickFeedService) UpdateCourseVisibility(ctx context.Context, in *qf.Enrollment) (*qf.Void, error) {
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
	return &qf.Void{}, err
}

// CreateEnrollment enrolls a new student for the course specified in the request.
// Access policy: Any User.
func (s *QuickFeedService) CreateEnrollment(_ context.Context, in *qf.Enrollment) (*qf.Void, error) {
	err := s.createEnrollment(in)
	if err != nil {
		s.logger.Errorf("CreateEnrollment failed: %v", err)
		err = status.Error(codes.InvalidArgument, "failed to create enrollment")
	}
	return &qf.Void{}, err
}

// UpdateEnrollments changes status of all pending enrollments for the specified course to approved.
// If the request contains a single enrollment, it will be updated to the specified status.
// Access policy: Teacher of CourseID
func (s *QuickFeedService) UpdateEnrollments(ctx context.Context, in *qf.Enrollments) (*qf.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateEnrollments failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	scmClient, err := s.getSCMForCourse(ctx, in.Enrollments[0].GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollments failed: could not create scm client: %v", err)
		return nil, ErrMissingInstallation
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Errorf("UpdateEnrollments failed: user %d is not teacher of course %d", usr.GetID(), in.GetCourseID())
		return nil, status.Error(codes.PermissionDenied, "only teachers can update enrollments")
	}

	for _, enrollment := range in.GetEnrollments() {
		if s.isCourseCreator(enrollment.CourseID, enrollment.UserID) {
			s.logger.Errorf("UpdateEnrollments failed: user %s attempted to demote course creator", usr.GetName())
			return nil, status.Error(codes.PermissionDenied, "course creator cannot be demoted")
		}

		if err = s.updateEnrollment(ctx, scmClient, usr.GetLogin(), enrollment); err != nil {
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
	return &qf.Void{}, err
}

// GetCoursesByUser returns all courses the given user is enrolled into with the given status.
// Access policy: Any User.
func (s *QuickFeedService) GetCoursesByUser(_ context.Context, in *qf.EnrollmentStatusRequest) (*qf.Courses, error) {
	courses, err := s.getCoursesByUser(in)
	if err != nil {
		s.logger.Errorf("GetCoursesWithEnrollment failed: %v", err)
		return nil, status.Error(codes.NotFound, "no courses with enrollment found")
	}
	return courses, nil
}

// GetEnrollmentsByUser returns all enrollments for the given user and enrollment status with preloaded courses and groups.
// Access policy: user with userID or admin
func (s *QuickFeedService) GetEnrollmentsByUser(ctx context.Context, in *qf.EnrollmentStatusRequest) (*qf.Enrollments, error) {
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
func (s *QuickFeedService) GetEnrollmentsByCourse(ctx context.Context, in *qf.EnrollmentRequest) (*qf.Enrollments, error) {
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
func (s *QuickFeedService) GetGroup(ctx context.Context, in *qf.GetGroupRequest) (*qf.Group, error) {
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
func (s *QuickFeedService) GetGroupsByCourse(ctx context.Context, in *qf.CourseRequest) (*qf.Groups, error) {
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
func (s *QuickFeedService) GetGroupByUserAndCourse(ctx context.Context, in *qf.GroupRequest) (*qf.Group, error) {
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
func (s *QuickFeedService) CreateGroup(ctx context.Context, in *qf.Group) (*qf.Group, error) {
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

// UpdateGroup updates group information, and returns the updated group.
// Access policy: Teacher of CourseID.
func (s *QuickFeedService) UpdateGroup(ctx context.Context, in *qf.Group) (*qf.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	scmClient, err := s.getSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: could not create scm client for group %s and course %d: %v", in.GetName(), in.GetCourseID(), err)
		return nil, ErrMissingInstallation
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("UpdateGroup failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can update groups")
	}
	err = s.updateGroup(ctx, scmClient, in)
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
	group, err := s.db.GetGroup(in.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get group")
	}
	return group, nil
}

// DeleteGroup removes group record from the database.
// Access policy: Teacher of CourseID.
func (s *QuickFeedService) DeleteGroup(ctx context.Context, in *qf.GroupRequest) (*qf.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: scm authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	scmClient, err := s.getSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: could not create scm client for group %d and course %d: %v", in.GetGroupID(), in.GetCourseID(), err)
		return nil, ErrMissingInstallation
	}
	grp, err := s.getGroup(&qf.GetGroupRequest{GroupID: in.GetGroupID()})
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get group")
	}
	if !s.isTeacher(usr.GetID(), grp.GetCourseID()) {
		s.logger.Error("DeleteGroup failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can delete groups")
	}
	if err = s.deleteGroup(ctx, scmClient, in); err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(errors.Unwrap(err)); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to delete group")
	}
	return &qf.Void{}, nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
// Access policy:
// Admin enrolled in CourseID,
// Current User if Owner of submission,
// Current User if member of group for group submission,
// Teacher of CourseID.
func (s *QuickFeedService) GetSubmissions(ctx context.Context, in *qf.SubmissionRequest) (*qf.Submissions, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}

	// grp may be nil if there is no group ID in request; this is fine, since the grp.Contains() returns false in this case.
	grp, _ := s.getGroup(&qf.GetGroupRequest{GroupID: in.GetGroupID()})

	// ensure that current user is teacher, enrolled admin, or the current user is owner of the submission request
	if !s.hasCourseAccess(usr.GetID(), in.GetCourseID(), func(e *qf.Enrollment) bool {
		switch e.Status {
		case qf.Enrollment_TEACHER:
			return true
		case qf.Enrollment_STUDENT:
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
	// If the user is not a teacher, remove score and reviews from submissions that are not released.
	if !s.isTeacher(usr.ID, in.CourseID) {
		submissions.Clean()
	}
	return submissions, nil
}

// GetSubmissionsByCourse returns all the latest submissions
// for every individual or group course assignment for all course students/groups.
// Access policy: Admin enrolled in CourseID, Teacher of CourseID.
func (s *QuickFeedService) GetSubmissionsByCourse(ctx context.Context, in *qf.SubmissionsForCourseRequest) (*qf.CourseSubmissions, error) {
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
	if !s.hasCourseAccess(usr.GetID(), in.GetCourseID(), func(e *qf.Enrollment) bool {
		switch e.Status {
		case qf.Enrollment_TEACHER:
			return true
		case qf.Enrollment_STUDENT:
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
func (s *QuickFeedService) UpdateSubmission(ctx context.Context, in *qf.UpdateSubmissionRequest) (*qf.Void, error) {
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
	return &qf.Void{}, err
}

// RebuildSubmissions re-runs the tests for the given assignment.
// A single submission is executed again if the request specifies a submission ID
// or all submissions if the request specifies a course ID.
// Access policy: Teacher of CourseID.
func (s *QuickFeedService) RebuildSubmissions(ctx context.Context, in *qf.RebuildRequest) (*qf.Void, error) {
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
	case *qf.RebuildRequest_SubmissionID:
		if !s.isValidSubmission(in.GetSubmissionID()) {
			s.logger.Errorf("RebuildSubmission failed: submitter has no access to the course")
			return nil, status.Error(codes.PermissionDenied, "submitter has no course access")
		}
		if _, err := s.rebuildSubmission(in); err != nil {
			s.logger.Errorf("RebuildSubmission failed: %v", err)
			return nil, status.Error(codes.InvalidArgument, "failed to rebuild submission "+err.Error())
		}
	case *qf.RebuildRequest_CourseID:
		if err := s.rebuildSubmissions(in); err != nil {
			s.logger.Errorf("RebuildSubmissions failed: %v", err)
			return nil, status.Error(codes.InvalidArgument, "failed to rebuild submissions "+err.Error())
		}
	}
	return &qf.Void{}, nil
}

// CreateBenchmark adds a new grading benchmark for an assignment
// Access policy: Teacher of CourseID
func (s *QuickFeedService) CreateBenchmark(_ context.Context, in *qf.GradingBenchmark) (*qf.GradingBenchmark, error) {
	bm, err := s.createBenchmark(in)
	if err != nil {
		s.logger.Errorf("CreateBenchmark failed for %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to add benchmark")
	}
	return bm, nil
}

// UpdateBenchmark edits a grading benchmark for an assignment
// Access policy: Teacher of CourseID
func (s *QuickFeedService) UpdateBenchmark(_ context.Context, in *qf.GradingBenchmark) (*qf.Void, error) {
	err := s.updateBenchmark(in)
	if err != nil {
		s.logger.Errorf("UpdateBenchmark failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update benchmark")
	}
	return &qf.Void{}, err
}

// DeleteBenchmark removes a grading benchmark
// Access policy: Teacher of CourseID
func (s *QuickFeedService) DeleteBenchmark(_ context.Context, in *qf.GradingBenchmark) (*qf.Void, error) {
	err := s.deleteBenchmark(in)
	if err != nil {
		s.logger.Errorf("DeleteBenchmark failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to delete benchmark")
	}
	return &qf.Void{}, err
}

// CreateCriterion adds a new grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *QuickFeedService) CreateCriterion(_ context.Context, in *qf.GradingCriterion) (*qf.GradingCriterion, error) {
	c, err := s.createCriterion(in)
	if err != nil {
		s.logger.Errorf("CreateCriterion failed for %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to add criterion")
	}
	return c, nil
}

// UpdateCriterion edits a grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *QuickFeedService) UpdateCriterion(_ context.Context, in *qf.GradingCriterion) (*qf.Void, error) {
	err := s.updateCriterion(in)
	if err != nil {
		s.logger.Errorf("UpdateCriterion failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update criterion")
	}
	return &qf.Void{}, err
}

// DeleteCriterion removes a grading criterion for an assignment
// Access policy: Teacher of CourseID
func (s *QuickFeedService) DeleteCriterion(_ context.Context, in *qf.GradingCriterion) (*qf.Void, error) {
	err := s.deleteCriterion(in)
	if err != nil {
		s.logger.Errorf("DeleteCriterion failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to delete criterion")
	}
	return &qf.Void{}, err
}

// CreateReview adds a new submission review
// Access policy: Teacher of CourseID
func (s *QuickFeedService) CreateReview(ctx context.Context, in *qf.ReviewRequest) (*qf.Review, error) {
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
func (s *QuickFeedService) UpdateReview(ctx context.Context, in *qf.ReviewRequest) (*qf.Review, error) {
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
func (s *QuickFeedService) UpdateSubmissions(ctx context.Context, in *qf.UpdateSubmissionsRequest) (*qf.Void, error) {
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
	return &qf.Void{}, err
}

// GetReviewers returns names of all active reviewers for a student submission
// Access policy: Teacher of CourseID
func (s *QuickFeedService) GetReviewers(ctx context.Context, in *qf.SubmissionReviewersRequest) (*qf.Reviewers, error) {
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
	return &qf.Reviewers{Reviewers: reviewers}, err
}

// GetAssignments returns a list of all assignments for the given course.
// Access policy: Any User.
func (s *QuickFeedService) GetAssignments(_ context.Context, in *qf.CourseRequest) (*qf.Assignments, error) {
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
func (s *QuickFeedService) UpdateAssignments(ctx context.Context, in *qf.CourseRequest) (*qf.Void, error) {
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
	return &qf.Void{}, nil
}

// GetOrganization fetches a github organization by name.
// Access policy: Admin
func (s *QuickFeedService) GetOrganization(ctx context.Context, in *qf.OrgRequest) (*qf.Organization, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetOrganization failed: scm authentication error: %v", err)
		return nil, err
	}
	scmClient, err := s.getSCM(ctx, in.GetOrgName())
	if err != nil {
		s.logger.Errorf("GetOrganization failed: could not create scm client for organization %s: %v", in.GetOrgName(), err)
		return nil, ErrMissingInstallation
	}
	if !usr.IsAdmin {
		s.logger.Error("GetOrganization failed: user is not admin")
		return nil, status.Error(codes.PermissionDenied, "only admin can access organizations")
	}
	org, err := s.getOrganization(ctx, scmClient, in.GetOrgName(), usr.GetLogin())
	if err != nil {
		s.logger.Errorf("GetOrganization failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == scm.ErrNotMember {
			return nil, status.Error(codes.NotFound, "organization membership not confirmed, please enable third-party access")
		}
		if err == ErrFreePlan || err == ErrAlreadyExists || err == scm.ErrNotOwner {
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
func (s *QuickFeedService) GetRepositories(ctx context.Context, in *qf.URLRequest) (*qf.Repositories, error) {
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
		case qf.Repository_USER:
			id = usr.GetID()
		case qf.Repository_GROUP:
			id = enrol.GetGroupID() // will be 0 if not enrolled in a group
		}
		repo, _ := s.getRepo(course, id, repoType)
		// for repo == nil: will result in an empty URL string, which will be ignored by the frontend
		urls[repoType.String()] = repo.GetHTMLURL()
	}
	return &qf.Repositories{URLs: urls}, nil
}

// IsEmptyRepo ensures that group repository is empty and can be deleted
// Access policy: Teacher of Course ID
func (s *QuickFeedService) IsEmptyRepo(ctx context.Context, in *qf.RepositoryRequest) (*qf.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: scm authentication error: %v", err)
		return nil, err
	}
	scmClient, err := s.getSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: could not create scm client for course %d: %v", in.GetCourseID(), err)
		return nil, ErrMissingInstallation
	}

	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("IsEmptyRepo failed: user is not teacher")
		return nil, status.Error(codes.PermissionDenied, "only teachers can access repository info")
	}

	if err := s.isEmptyRepo(ctx, scmClient, in); err != nil {
		s.logger.Errorf("IsEmptyRepo failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.FailedPrecondition, "group repository does not exist or not empty")
	}
	return &qf.Void{}, nil
}
