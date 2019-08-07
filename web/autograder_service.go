package web

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/web/auth"
)

// AutograderService holds references to the database and
// other shared data structures.
type AutograderService struct {
	logger *zap.SugaredLogger
	db     *database.GormDB
	scms   *auth.Scms
	bh     BaseHookOptions
}

// NewAutograderService returns an AutograderService object.
func NewAutograderService(logger *zap.Logger, db *database.GormDB, scms *auth.Scms, bh BaseHookOptions) *AutograderService {
	return &AutograderService{
		logger: logger.Sugar(),
		db:     db,
		scms:   scms,
		bh:     bh,
	}
}

// GetUsers returns a list of all users.
// Access policy: Admin.
// Frontend note: This method is called from AdminPage.
func (s *AutograderService) GetUsers(ctx context.Context, in *pb.Void) (*pb.Users, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUsers failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	if !usr.IsAdmin {
		s.logger.Error("GetUsers failed: user is not admin")
		return nil, status.Errorf(codes.PermissionDenied, "only admin can access other users")
	}
	usrs, err := s.getUsers()
	if err != nil {
		s.logger.Errorf("GetUsers failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "failed to get users")
	}
	return usrs, nil
}

// UpdateUser updates the current users's information and returns the updated user.
// This function can also promote a user to admin or demote a user.
// Access policy: Admin can update other users's information and promote to Admin;
// Current User if Owner can update its own information.
func (s *AutograderService) UpdateUser(ctx context.Context, in *pb.User) (*pb.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateUser failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	// TODO(vera): this check feels a bit excessive: course creator user is always admin
	if !(usr.IsAdmin || usr.IsOwner(in.GetID())) {
		s.logger.Errorf("UpdateUser failed to update user %d: user is not admin or course creator", in.GetID())
		return nil, status.Errorf(codes.PermissionDenied, "only admin can update another user")
	}
	usr, err = s.updateUser(usr.IsAdmin, in)
	if err != nil {
		s.logger.Errorf("UpdateUser failed to update user %d: %s", in.GetID(), err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to update current user")
	}
	return usr, nil
}

// IsAuthorizedTeacher checks whether current user has teacher scopes.
// Access policy: Any User.
func (s *AutograderService) IsAuthorizedTeacher(ctx context.Context, in *pb.Void) (*pb.AuthorizationResponse, error) {
	// TODO(vera): upgrade to send provider from client. Currently not supported for other providers anyway
	// Hein @Vera: it may be easier to pass along the courseID from the client as is done for UpdateEnrollment (see below)
	// any user can ask if they have teacher scopes.
	_, scm, err := s.getUserAndSCM(ctx, "github")
	if err != nil {
		s.logger.Errorf("IsAuthorizedTeacher failed: scm authentication error (%s)", err)
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
		s.logger.Errorf("CreateCourse failed: scm authentication error (%s)", err)
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
		s.logger.Error("CreateCourse failed: ", err)
		if err == ErrAlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
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
		s.logger.Errorf("UpdateCourse failed: scm authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("UpdateCourse failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update course")
	}

	if err = s.updateCourse(ctx, scm, in); err != nil {
		s.logger.Errorf("UpdateCourse failed: %s", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to update course")
	}
	return &pb.Void{}, nil
}

// GetCourse returns course information for the given course.
// Access policy: Any User.
func (s *AutograderService) GetCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Course, error) {
	courseID := in.GetID()
	course, err := s.getCourse(courseID)
	if err != nil {
		s.logger.Errorf("GetCourse failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "course not found")
	}
	return course, nil
}

// GetCourses returns a list of all courses.
// Access policy: Any User.
func (s *AutograderService) GetCourses(ctx context.Context, in *pb.Void) (*pb.Courses, error) {
	courses, err := s.getCourses()
	if err != nil {
		s.logger.Errorf("GetCourses failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "no courses found")
	}
	return courses, nil
}

// CreateEnrollment enrolls a new student for the course specified in the request.
// Access policy: Any User.
func (s *AutograderService) CreateEnrollment(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	err := s.createEnrollment(in)
	if err != nil {
		s.logger.Errorf("CreateEnrollment failed: %s", err)
		return nil, status.Error(codes.InvalidArgument, "failed to create enrollment")
	}
	return &pb.Void{}, nil
}

// UpdateEnrollment updates the enrollment status of a student as specified in the request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateEnrollment(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCM2(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollment failed: scm authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("UpdateEnrollment failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update enrollment status")
	}

	err = s.updateEnrollment(ctx, scm, in)
	if err != nil {
		s.logger.Errorf("UpdateEnrollment failed: %s", err)
		return nil, status.Error(codes.InvalidArgument, "failed to update enrollment")
	}
	return &pb.Void{}, nil
}

// GetCoursesWithEnrollment returns all courses with enrollments of the type specified in the request.
// Access policy: Any User.
func (s *AutograderService) GetCoursesWithEnrollment(ctx context.Context, in *pb.RecordRequest) (*pb.Courses, error) {
	courses, err := s.getCoursesWithEnrollment(in)
	if err != nil {
		s.logger.Errorf("GetCoursesWithEnrollment failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "no courses with enrollment found")
	}
	return courses, nil
}

// GetEnrollment returns a single enrollment in given course for user or group.
// Access policy: Admin, Current User if Owner, Current User if Teacher of CourseID.
// Frontend note: This method is not really used, but is accessible from
// ServerProvider.getEnrollment(), and used by ServerProvider.isTeacher() = UserManager.isTeacher().
// TODO(meling) it is not clear if this will be used and by who.
func (s *AutograderService) GetEnrollment(ctx context.Context, in *pb.EnrollmentRequest) (*pb.Enrollment, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetEnrollment failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	if !(usr.IsAdmin || usr.IsOwner(in.GetUserID()) || s.isTeacher(usr.GetID(), in.GetCourseID())) {
		s.logger.Error("GetEnrollment failed: user is not teacher, admin or enrollment owner")
		return nil, status.Errorf(codes.PermissionDenied, "cannot get enrollment for given course and user")
	}
	return s.getEnrollment(in)
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.EnrollmentsRequest) (*pb.Enrollments, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("GetEnrollmentsByCourse failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can get course enrollments")
	}

	enrolls, err := s.getEnrollmentsByCourse(in)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: %s", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to get enrollments for given course")
	}
	return enrolls, nil
}

// GetGroup returns information about a group.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroup(ctx context.Context, in *pb.RecordRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroup failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	group, err := s.getGroup(in)
	if err != nil {
		s.logger.Errorf("GetGroup failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "failed to get group")
	}
	if !(group.Contains(usr) || s.isTeacher(usr.GetID(), group.GetCourseID())) {
		s.logger.Error("GetGroup failed: user is not group member or teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only group members and teachers can access a group")
	}
	return group, nil
}

// GetGroups returns a list of groups created for the course id in the record request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) GetGroups(ctx context.Context, in *pb.RecordRequest) (*pb.Groups, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroups failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("GetGroups failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can access other groups")
	}
	groups, err := s.getGroups(in)
	if err != nil {
		s.logger.Errorf("GetGroups failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "failed to get groups")
	}
	return groups, nil
}

// GetGroupByUserAndCourse returns the group of the given student for a given course.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroupByUserAndCourse(ctx context.Context, in *pb.GroupRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetGroupByUserAndCourse failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	group, err := s.getGroupByUserAndCourse(in)
	if err != nil {
		s.logger.Errorf("GetGroupByUserAndCourse failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "failed to get group for given user and course")
	}
	if !(group.Contains(usr) || s.isTeacher(usr.GetID(), group.GetCourseID())) {
		s.logger.Error("GetGroupByUserAndCourse failed: user is not group member or teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only group members and teachers can access another group")
	}
	return group, nil
}

// CreateGroup creates a new group.
// Access policy: Any User enrolled in course and specified as member of the group or a course teacher.
func (s *AutograderService) CreateGroup(ctx context.Context, in *pb.Group) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("CreateGroup failed: authentication error (%s)", err)
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
		s.logger.Errorf("CreateGroup failed: %s", err)
		if _, ok := status.FromError(err); !ok {
			// set err to generic error for the frontend
			err = status.Error(codes.InvalidArgument, "failed to create group")
		}
		// this err may be a grpc status type to be returned to the client
		return nil, err
	}
	return group, nil
}

// UpdateGroup updates group information.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCM2(ctx, in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: scm authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		s.logger.Error("UpdateGroup failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update groups")
	}
	err = s.updateGroup(ctx, scm, in)
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: %s", err)
		if _, ok := status.FromError(err); !ok {
			// set err to generic error for the frontend
			err = status.Error(codes.InvalidArgument, "failed to update group")
		}
		// this err may be a grpc status type to be returned to the client
		return nil, err
	}
	return &pb.Void{}, nil
}

// DeleteGroup removes group record from the database.
// Access policy: Teacher of CourseID.
func (s *AutograderService) DeleteGroup(ctx context.Context, in *pb.RecordRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		s.logger.Error("DeleteGroup failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can delete groups")
	}
	if err = s.deleteGroup(in); err != nil {
		s.logger.Errorf("DeleteGroup failed: %s", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to delete group")
	}
	return &pb.Void{}, nil
}

// GetSubmission returns a student submission.
// Access policy: Current User.
func (s *AutograderService) GetSubmission(ctx context.Context, in *pb.RecordRequest) (*pb.Submission, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetSubmission failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	submission, err := s.getSubmission(usr, in)
	if err != nil {
		s.logger.Errorf("GetSubmission failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "no submission found")
	}
	return submission, nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
// Access policy: Current User if Owner of submission, Teacher of CourseID.
func (s *AutograderService) GetSubmissions(ctx context.Context, in *pb.SubmissionRequest) (*pb.Submissions, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	// ensure that current user is teacher or the current user is owner of the submission request
	if !s.hasCourseAccess(usr.GetID(), in.GetCourseID(), func(e *pb.Enrollment) bool {
		return e.Status == pb.Enrollment_TEACHER ||
			(e.Status == pb.Enrollment_STUDENT && usr.IsOwner(in.GetUserID()))
	}) {
		s.logger.Error("GetSubmissions failed: user is not teacher or submission author")
		return nil, status.Errorf(codes.PermissionDenied, "only owner and teachers can get submissions")
	}
	submissions, err := s.getSubmissions(in)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "no submissions found")
	}
	return submissions, nil
}

// ApproveSubmission approves the given submission.
// Access policy: Teacher of CourseID.
func (s *AutograderService) ApproveSubmission(ctx context.Context, in *pb.ApproveSubmissionRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("ApproveSubmission failed: authentication error (%s)", err)
		return nil, ErrInvalidUserInfo
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		s.logger.Error("ApproveSubmision failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can approve submissions")
	}
	err = s.approveSubmission(in.GetSubmissionID())
	if err != nil {
		s.logger.Errorf("ApproveSubmission failed: %s", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to approve submission")
	}
	return &pb.Void{}, nil
}

// GetAssignments returns a list of all assignments for the given course.
// Access policy: Any User.
func (s *AutograderService) GetAssignments(ctx context.Context, in *pb.RecordRequest) (*pb.Assignments, error) {
	courseID := in.GetID()
	assignments, err := s.getAssignments(courseID)
	if err != nil {
		s.logger.Errorf("GetAssignments failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "no assignments found for course")
	}
	return assignments, nil
}

// UpdateAssignments returns latest information about the course
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateAssignments(ctx context.Context, in *pb.RecordRequest) (*pb.Void, error) {
	courseID := in.GetID()
	usr, scm, err := s.getUserAndSCM2(ctx, courseID)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: scm authentication error (%s)", err)
		return nil, err
	}
	if !s.isTeacher(usr.ID, courseID) {
		s.logger.Error("UpdateAssignments failed: user is not teacher")
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update course assignments")
	}
	err = s.updateAssignments(ctx, scm, courseID)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: %s", err)
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

// GetOrganizations returns list of organizations available for course creation.
// Access policy: Any User.
func (s *AutograderService) GetOrganizations(ctx context.Context, in *pb.Provider) (*pb.Organizations, error) {
	_, scm, err := s.getUserAndSCM(ctx, in.Provider)
	if err != nil {
		s.logger.Errorf("GetOrganizations failed: scm authentication error (%s)", err)
		return nil, err
	}
	orgs, err := s.getAvailableOrganizations(ctx, scm)
	if err != nil {
		s.logger.Errorf("GetOrganizations failed: %s", err)
		return nil, status.Errorf(codes.NotFound, "found no organizations to host course")
	}
	return orgs, nil
}

// GetRepositories returns URL strings for repositories of given type for the given course
// Access policy: Any User.
func (s *AutograderService) GetRepositories(ctx context.Context, in *pb.URLRequest) (*pb.Repositories, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetRepositories failed: authentication error (%s)", err)
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
