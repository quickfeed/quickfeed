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

// GetRepositoryURL returns a repository URL for the requested repository type.
// Access policy: Any User can request these URLs. However, only those with access
// to the different repositories can actually use the URL.
func (s *AutograderService) GetRepositoryURL(ctx context.Context, in *pb.RepositoryRequest) (*pb.URLResponse, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	repoURL, err := s.getRepositoryURL(usr, in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get repository URL")
	}
	return repoURL, nil
}

// GetUsers returns a list of all users.
// Access policy: Admin.
// Frontend note: This method is called from AdminPage.
func (s *AutograderService) GetUsers(ctx context.Context, in *pb.Void) (*pb.Users, error) {
	if !s.isAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "only admin can access other users")
	}
	usrs, err := s.getUsers()
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get users")
	}
	return usrs, nil
}

// UpdateUser updates the current users's information and returns the updated user.
// This function can also promote a user to admin or demote a user.
// Access policy: Admin can update other users's information and promote to Admin;
// Current User can update its own information.
func (s *AutograderService) UpdateUser(ctx context.Context, in *pb.User) (*pb.User, error) {
	if !s.hasAccess(ctx, in.ID) {
		return nil, status.Errorf(codes.PermissionDenied, "only admin can update another user")
	}
	usr, err := s.updateUser(s.isAdmin(ctx), in)
	if err != nil {
		s.logger.Error(err)
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
		return nil, err
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
		return nil, err
	}
	if !usr.IsAdmin {
		s.logger.Error("user must be admin to create course")
		return nil, status.Error(codes.PermissionDenied, "user must be admin to create course")
	}

	// make sure that the current user is set as course creator
	in.CourseCreatorID = usr.GetID()
	course, err := s.createCourse(ctx, scm, in)
	if err != nil {
		s.logger.Error(err)
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
		return nil, err
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update course")
	}

	if err = s.updateCourse(ctx, scm, in); err != nil {
		s.logger.Error(err)
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
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "course not found")
	}
	return course, nil
}

// GetCourses returns a list of all courses.
// Access policy: Any User.
func (s *AutograderService) GetCourses(ctx context.Context, in *pb.Void) (*pb.Courses, error) {
	courses, err := s.getCourses()
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "no courses found")
	}
	return courses, nil
}

// CreateEnrollment enrolls a new student for the course specified in the request.
// Access policy: Any User.
func (s *AutograderService) CreateEnrollment(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	err := s.createEnrollment(in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.InvalidArgument, "failed to create enrollment")
	}
	return &pb.Void{}, nil
}

// UpdateEnrollment updates the enrollment status of a student as specified in the request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) UpdateEnrollment(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	usr, scm, err := s.getUserAndSCM2(ctx, in.GetCourseID())
	if err != nil {
		return nil, err
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update enrollment status")
	}

	err = s.updateEnrollment(ctx, scm, in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.InvalidArgument, "failed to update enrollment")
	}
	return &pb.Void{}, nil
}

// GetCoursesWithEnrollment returns all courses with enrollments of the type specified in the request.
// Access policy: Any User.
func (s *AutograderService) GetCoursesWithEnrollment(ctx context.Context, in *pb.RecordRequest) (*pb.Courses, error) {
	courses, err := s.getCoursesWithEnrollment(in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "no courses with enrollment found")
	}
	return courses, nil
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !s.isTeacher(usr.GetID(), in.GetCourseID()) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can get enrollments")
	}

	enrolls, err := s.getEnrollmentsByCourse(in)
	if err != nil {
		return nil, err
	}
	return enrolls, nil
}

// GetGroup returns information about a group.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroup(ctx context.Context, in *pb.RecordRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	group, err := s.getGroup(in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get group")
	}
	if !(s.isTeacher(usr.GetID(), group.GetCourseID()) || isInGroup(usr, group)) {
		return nil, status.Errorf(codes.PermissionDenied, "only group members and teachers can access a group")
	}
	return group, nil
}

// GetGroups returns a list of groups created for the course id in the record request.
// Access policy: Teacher of CourseID.
func (s *AutograderService) GetGroups(ctx context.Context, in *pb.RecordRequest) (*pb.Groups, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can access other groups")
	}
	groups, err := s.getGroups(in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get groups")
	}
	return groups, nil
}

// GetGroupByUserAndCourse returns the group of the given student for a given course.
// Access policy: Group members, Teacher of CourseID.
func (s *AutograderService) GetGroupByUserAndCourse(ctx context.Context, in *pb.GroupRequest) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	group, err := s.getGroupByUserAndCourse(in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get group for given user and course")
	}
	if !(s.isTeacher(usr.GetID(), group.GetCourseID()) || isInGroup(usr, group)) {
		return nil, status.Errorf(codes.PermissionDenied, "only group members and teachers can access another group")
	}
	return group, nil
}

// CreateGroup creates a new group.
// Access policy: Any User.
func (s *AutograderService) CreateGroup(ctx context.Context, in *pb.Group) (*pb.Group, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	group, err := s.createGroup(usr, in)
	if err != nil {
		s.logger.Error(err)
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
		return nil, err
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update groups")
	}

	err = s.updateGroup(ctx, scm, usr, in)
	if err != nil {
		s.logger.Error(err)
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
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	courseID := in.GetID()
	if !s.isTeacher(usr.GetID(), courseID) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can delete groups")
	}
	if err = s.deleteGroup(in); err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to delete group")
	}
	return &pb.Void{}, nil
}

// GetSubmission returns a student submission.
// Access policy: Current User.
func (s *AutograderService) GetSubmission(ctx context.Context, in *pb.RecordRequest) (*pb.Submission, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	submission, err := s.getSubmission(usr, in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "no submission found")
	}
	return submission, nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
// Access policy: Current User if owner of submission, Teacher of CourseID.
func (s *AutograderService) GetSubmissions(ctx context.Context, in *pb.SubmissionRequest) (*pb.Submissions, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	// ensure that current user is teacher or that the current user is owner of the submission request
	if !(s.isTeacher(usr.GetID(), in.GetCourseID()) || s.isOwner(usr.GetID(), in.GetUserID(), in.GetCourseID())) {
		return nil, status.Errorf(codes.PermissionDenied, "only owner and teachers can get submissions")
	}
	submissions, err := s.getSubmissions(in)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "no submissions found")
	}
	return submissions, nil
}

// ApproveSubmission approves the given submission.
// Access policy: Teacher of CourseID.
func (s *AutograderService) ApproveSubmission(ctx context.Context, in *pb.ApproveSubmissionRequest) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	if !s.isTeacher(usr.ID, in.GetCourseID()) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can approve submissions")
	}
	err = s.approveSubmission(in.GetSubmissionID())
	if err != nil {
		s.logger.Error(err)
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
		s.logger.Error(err)
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
		return nil, err
	}
	if !s.isTeacher(usr.ID, courseID) {
		return nil, status.Errorf(codes.PermissionDenied, "only teachers can update course assignments")
	}
	err = s.updateAssignments(ctx, scm, courseID)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to update assignments for course")
	}
	return &pb.Void{}, nil
}

// GetProviders returns a list of SCM providers supported by the backend.
// Access policy: Any User.
func (s *AutograderService) GetProviders(ctx context.Context, in *pb.Void) (*pb.Providers, error) {
	providers := auth.GetProviders()
	if len(providers.GetProviders()) < 1 {
		s.logger.Error("found no enabled SCM providers")
		return nil, status.Errorf(codes.NotFound, "found no enabled SCM providers")
	}
	return providers, nil
}

// GetOrganizations returns list of organizations available for course creation.
// Access policy: Any User.
func (s *AutograderService) GetOrganizations(ctx context.Context, in *pb.Provider) (*pb.Organizations, error) {
	_, scm, err := s.getUserAndSCM(ctx, in.Provider)
	if err != nil {
		return nil, err
	}
	orgs, err := s.getAvailableOrganizations(ctx, scm)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "found no organizations to host course")
	}
	return orgs, nil
}

// GetRepository is not yet implemented
// Access policy: Any User.
func (s *AutograderService) GetRepository(ctx context.Context, in *pb.RepositoryRequest) (*pb.Repository, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
