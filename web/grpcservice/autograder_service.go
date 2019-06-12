package grpcservice

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"

	"github.com/autograde/aguis/web"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
)

// AutograderService holds references to the database and
// other shared data structures.
type AutograderService struct {
	logger *zap.SugaredLogger
	db     *database.GormDB
	scms   *web.Scms
	bh     web.BaseHookOptions
}

// NewAutograderService returns an AutograderService object.
func NewAutograderService(logger *zap.Logger, db *database.GormDB, scms *web.Scms, bh web.BaseHookOptions) *AutograderService {
	return &AutograderService{
		logger: logger.Sugar(),
		db:     db,
		scms:   scms,
		bh:     bh,
	}
}

// GetRepositoryURL returns a repository URL for the requested repository type.
func (s *AutograderService) GetRepositoryURL(ctx context.Context, in *pb.RepositoryRequest) (*pb.URLResponse, error) {
	if !in.IsValidRepoRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	currentUser, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	repoURL, err := web.GetRepositoryURL(currentUser, in, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to fetch repository URL")
	}
	return repoURL, nil
}

// GetUser returns user information for the given user, excluding remote identities.
func (s *AutograderService) GetUser(ctx context.Context, in *pb.RecordRequest) (*pb.User, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := web.GetUser(in, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get user")
	}
	usr.RemoveRemoteID()
	return usr, nil
}

// GetUsers returns a list of all users.
func (s *AutograderService) GetUsers(ctx context.Context, in *pb.Void) (*pb.Users, error) {
	usrs, err := web.GetUsers(s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get users")
	}
	usrs.RemoveRemoteIDs()
	return usrs, nil
}

// UpdateUser updates the current users's information and returns the updated user.
//TODO(meling) should this also allow teacher/admin to update another user?
func (s *AutograderService) UpdateUser(ctx context.Context, in *pb.User) (*pb.User, error) {
	if !in.IsValidUser() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	currentUser, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	usr, err := web.PatchUser(currentUser, in, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to update current user")
	}
	usr.RemoveRemoteID()
	return usr, nil
}

// CreateCourse creates a new course.
// Only users with teacher role (admin) can create new courses.
func (s *AutograderService) CreateCourse(ctx context.Context, in *pb.Course) (*pb.Course, error) {
	if !in.IsValidCourse() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	if !usr.IsAdmin {
		s.logger.Error(err)
		return nil, status.Errorf(codes.PermissionDenied, "user must be admin to create or update")
	}
	scm, err := s.getSCM(ctx, in.Provider)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get SCM for user")
	}

	// make sure that the current user is set as course creator
	in.CourseCreatorID = usr.GetID()
	course, err := web.NewCourse(ctx, in, s.db, scm, s.bh)
	if err != nil {
		s.logger.Error(err)
		if err == web.ErrAlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.InvalidArgument, "failed to create course")
	}
	return course, nil
}

// UpdateCourse changes the course information details.
// Only users with teacher role (admin) can update the course details.
func (s *AutograderService) UpdateCourse(ctx context.Context, in *pb.Course) (*pb.Void, error) {
	if !in.IsValidCourse() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	if !usr.IsAdmin {
		s.logger.Error(err)
		return nil, status.Errorf(codes.PermissionDenied, "user must be admin to create or update")
	}
	scm, err := s.getSCM(ctx, in.Provider)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get SCM for user")
	}
	return &pb.Void{}, web.UpdateCourse(ctx, in, s.db, scm)
}

// GetCourse returns course information for the given course.
func (s *AutograderService) GetCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Course, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.GetCourse(in, s.db)
}

// GetCourses returns a list of all courses.
func (s *AutograderService) GetCourses(ctx context.Context, in *pb.Void) (*pb.Courses, error) {
	return web.ListCourses(s.db)
}

// CreateEnrollment enrolls a new student for the course specified in the request.
func (s *AutograderService) CreateEnrollment(ctx context.Context, in *pb.ActionRequest) (*pb.Void, error) {
	if !in.IsValidEnrollment() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return &pb.Void{}, web.CreateEnrollment(in, s.db)
}

// UpdateEnrollment updates the enrollment status of a student as specified in the request.
func (s *AutograderService) UpdateEnrollment(ctx context.Context, in *pb.ActionRequest) (*pb.Void, error) {
	if !in.IsValidEnrollment() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	if !usr.IsAdmin {
		s.logger.Error(err)
		return nil, status.Errorf(codes.PermissionDenied, "user must be admin to create or update")
	}
	crs, err := web.GetCourse(&pb.RecordRequest{ID: in.CourseID}, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get course")
	}
	scm, err := s.getSCM(ctx, crs.Provider)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get SCM for user")
	}

	return &pb.Void{}, web.UpdateEnrollment(ctx, in, s.db, scm)
}

// GetCoursesWithEnrollment returns all courses with enrollments of the type specified in the request.
func (s *AutograderService) GetCoursesWithEnrollment(ctx context.Context, in *pb.RecordRequest) (*pb.Courses, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	//TODO(meling) these direct calls and returns needs to be logged here and return status.Error instead
	return web.ListCoursesWithEnrollment(in, s.db)
}

// GetAssignments returns a list of all assignments.
func (s *AutograderService) GetAssignments(ctx context.Context, in *pb.RecordRequest) (*pb.Assignments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.ListAssignments(in, s.db)
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	enrolls, err := web.GetEnrollmentsByCourse(in, s.db)
	if err != nil {
		return nil, err
	}
	enrolls.RemoveRemoteIDs()
	return enrolls, nil
}

// GetGroup returns information about a group
func (s *AutograderService) GetGroup(ctx context.Context, in *pb.RecordRequest) (*pb.Group, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	group, err := web.GetGroup(in, s.db)
	if err != nil {
		return nil, err
	}
	group.RemoveRemoteIDs()
	return group, nil
}

// GetGroups returns a list of student groups created for the course
func (s *AutograderService) GetGroups(ctx context.Context, in *pb.RecordRequest) (*pb.Groups, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	groups, err := web.GetGroups(in, s.db)
	if err != nil {
		return nil, err
	}
	groups.RemoveRemoteIDs()
	return groups, nil
}

// CreateGroup makes a new group
func (s *AutograderService) CreateGroup(ctx context.Context, in *pb.Group) (*pb.Group, error) {
	if !in.IsValidGroup() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	group, err := web.NewGroup(in, s.db, usr)
	if err != nil {
		return nil, err
	}
	group.RemoveRemoteIDs()
	return group, nil
}

// UpdateGroup updates group information
func (s *AutograderService) UpdateGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	if !in.IsValidGroup() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}

	crs, err := web.GetCourse(&pb.RecordRequest{ID: in.CourseID}, s.db)
	if err != nil {
		return nil, err
	}

	scm, err := s.getSCM(ctx, crs.Provider)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get SCM for user")
	}

	return &pb.Void{}, web.UpdateGroup(ctx, in, s.db, scm, usr)
}

// DeleteGroup removes group record from the database
func (s *AutograderService) DeleteGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	if in.GetID() < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return &pb.Void{}, web.DeleteGroup(in, s.db)
}

// GetSubmission returns a student submission
func (s *AutograderService) GetSubmission(ctx context.Context, in *pb.RecordRequest) (*pb.Submission, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	return web.GetSubmission(in, s.db, usr)
}

// GetSubmissions returns a list of submissions
func (s *AutograderService) GetSubmissions(ctx context.Context, in *pb.ActionRequest) (*pb.Submissions, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.ListSubmissions(in, s.db)
}

// GetGroupSubmissions returns all submissions of a student group
func (s *AutograderService) GetGroupSubmissions(ctx context.Context, in *pb.ActionRequest) (*pb.Submissions, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.ListGroupSubmissions(in, s.db)
}

// UpdateSubmission changes submission information
func (s *AutograderService) UpdateSubmission(ctx context.Context, in *pb.RecordRequest) (*pb.Void, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return &pb.Void{}, web.UpdateSubmission(in, s.db)
}

// RefreshCourse returns latest information about the course
func (s *AutograderService) RefreshCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Assignments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	crs, err := web.GetCourse(in, s.db)
	if err != nil {
		return nil, err
	}
	scm, err := s.getSCM(ctx, crs.Provider)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get SCM for user")
	}
	return web.RefreshCourse(ctx, in, scm, s.db, usr)
}

// GetGroupByUserAndCourse returns a student group
func (s *AutograderService) GetGroupByUserAndCourse(ctx context.Context, in *pb.ActionRequest) (*pb.Group, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	group, err := web.GetGroupByUserAndCourse(in, s.db)
	if err != nil {
		return nil, err
	}
	group.RemoveRemoteIDs()
	return group, nil
}

// GetProviders returns a list of providers
func (s *AutograderService) GetProviders(ctx context.Context, in *pb.Void) (*pb.Providers, error) {
	return web.GetProviders()
}

// GetDirectories returns a list of directories for a course
func (s *AutograderService) GetDirectories(ctx context.Context, in *pb.DirectoryRequest) (*pb.Directories, error) {
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()

	scm, err := s.getSCM(ctx, in.Provider)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get SCM for user")
	}
	return web.ListDirectories(ctx, s.db, scm)
}

// GetRepository is not yet implemented
func (s *AutograderService) GetRepository(ctx context.Context, in *pb.RepositoryRequest) (*pb.Repository, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
