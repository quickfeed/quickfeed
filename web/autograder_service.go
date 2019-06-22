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
func (s *AutograderService) GetRepositoryURL(ctx context.Context, in *pb.RepositoryRequest) (*pb.URLResponse, error) {
	if !in.IsValidRepoRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	currentUser, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	repoURL, err := s.getRepositoryURL(currentUser, in)
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
	usr, err := GetUser(in, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get user")
	}
	usr.RemoveRemoteID()
	return usr, nil
}

// GetUsers returns a list of all users.
func (s *AutograderService) GetUsers(ctx context.Context, in *pb.Void) (*pb.Users, error) {
	usrs, err := GetUsers(s.db)
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
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	currentUser, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	usr, err := PatchUser(currentUser, in, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to update current user")
	}
	usr.RemoveRemoteID()
	return usr, nil
}

// IsAuthorizedTeacher checks whether current user has teacher scopes
func (s *AutograderService) IsAuthorizedTeacher(ctx context.Context, in *pb.Void) (*pb.AuthorizationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	// TODO(vera): upgrade to send provider from client. Currently not supported for other clients anyway
	// Hein @Vera: it may be easier to pass along the courseID from the client as is done for UpdateEnrollment (see below)
	// Hein @Vera: should the current user be admin; if so, the bool param should be set to true
	_, scm, err := s.getUserAndSCM(ctx, "github", false)
	if err != nil {
		return nil, err
	}

	isAuthorized := HasTeacherScopes(ctx, scm)
	return &pb.AuthorizationResponse{IsAuthorized: isAuthorized}, nil
}

// CreateCourse creates a new course.
// Only users with teacher role (admin) can create new courses.
func (s *AutograderService) CreateCourse(ctx context.Context, in *pb.Course) (*pb.Course, error) {
	if !in.IsValidCourse() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	usr, scm, err := s.getUserAndSCM(ctx, in.Provider, true)
	if err != nil {
		return nil, err
	}

	// make sure that the current user is set as course creator
	in.CourseCreatorID = usr.GetID()
	course, err := NewCourse(ctx, in, s.db, scm, s.bh)
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
// Only users with teacher role (admin) can update the course details.
func (s *AutograderService) UpdateCourse(ctx context.Context, in *pb.Course) (*pb.Void, error) {
	if !in.IsValidCourse() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	_, scm, err := s.getUserAndSCM(ctx, in.Provider, true)
	if err != nil {
		return nil, err
	}

	if err = UpdateCourse(ctx, in, s.db, scm); err != nil {
		s.logger.Error(err)
		err = status.Errorf(codes.InvalidArgument, "failed to update course")
	}
	return &pb.Void{}, err
}

// GetCourse returns course information for the given course.
func (s *AutograderService) GetCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Course, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return GetCourse(in, s.db)
}

// GetCourses returns a list of all courses.
func (s *AutograderService) GetCourses(ctx context.Context, in *pb.Void) (*pb.Courses, error) {
	return ListCourses(s.db)
}

// CreateEnrollment enrolls a new student for the course specified in the request.
func (s *AutograderService) CreateEnrollment(ctx context.Context, in *pb.ActionRequest) (*pb.Void, error) {
	if !in.IsValidEnrollment() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return &pb.Void{}, CreateEnrollment(in, s.db)
}

// UpdateEnrollment updates the enrollment status of a student as specified in the request.
func (s *AutograderService) UpdateEnrollment(ctx context.Context, in *pb.ActionRequest) (*pb.Void, error) {
	if !in.IsValidEnrollment() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	crs, err := GetCourse(&pb.RecordRequest{ID: in.CourseID}, s.db)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get course")
	}
	_, scm, err := s.getUserAndSCM(ctx, crs.Provider, true)
	if err != nil {
		return nil, err
	}

	return &pb.Void{}, UpdateEnrollment(ctx, in, s.db, scm)
}

// GetCoursesWithEnrollment returns all courses with enrollments of the type specified in the request.
func (s *AutograderService) GetCoursesWithEnrollment(ctx context.Context, in *pb.RecordRequest) (*pb.Courses, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	//TODO(meling) these direct calls and returns needs to be logged here and return status.Error instead
	return ListCoursesWithEnrollment(in, s.db)
}

// GetAssignments returns a list of all assignments.
func (s *AutograderService) GetAssignments(ctx context.Context, in *pb.RecordRequest) (*pb.Assignments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return ListAssignments(in, s.db)
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	enrolls, err := GetEnrollmentsByCourse(in, s.db)
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
	group, err := GetGroup(in, s.db)
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
	groups, err := GetGroups(in, s.db)
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
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	group, err := NewGroup(in, s.db, usr)
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
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	crs, err := GetCourse(&pb.RecordRequest{ID: in.CourseID}, s.db)
	if err != nil {
		return nil, err
	}
	usr, scm, err := s.getUserAndSCM(ctx, crs.Provider, false)
	if err != nil {
		return nil, err
	}

	return &pb.Void{}, UpdateGroup(ctx, in, s.db, scm, usr)
}

// DeleteGroup removes group record from the database
func (s *AutograderService) DeleteGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	if in.GetID() < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return &pb.Void{}, DeleteGroup(in, s.db)
}

// GetSubmission returns a student submission
func (s *AutograderService) GetSubmission(ctx context.Context, in *pb.RecordRequest) (*pb.Submission, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	return GetSubmission(in, s.db, usr)
}

// GetSubmissions returns a list of submissions
func (s *AutograderService) GetSubmissions(ctx context.Context, in *pb.ActionRequest) (*pb.Submissions, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return ListSubmissions(in, s.db)
}

// GetGroupSubmissions returns all submissions of a student group
func (s *AutograderService) GetGroupSubmissions(ctx context.Context, in *pb.ActionRequest) (*pb.Submissions, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return ListGroupSubmissions(in, s.db)
}

// UpdateSubmission changes submission information
func (s *AutograderService) UpdateSubmission(ctx context.Context, in *pb.RecordRequest) (*pb.Void, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	//TODO(meling) UpdateSubmission requires administrator/teacher access
	return &pb.Void{}, UpdateSubmission(in, s.db)
}

// RefreshCourse returns latest information about the course
func (s *AutograderService) RefreshCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Assignments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	crs, err := GetCourse(in, s.db)
	if err != nil {
		return nil, err
	}
	usr, scm, err := s.getUserAndSCM(ctx, crs.Provider, false)
	if err != nil {
		return nil, err
	}
	return RefreshCourse(ctx, in, scm, s.db, usr)
}

// GetGroupByUserAndCourse returns a student group
func (s *AutograderService) GetGroupByUserAndCourse(ctx context.Context, in *pb.ActionRequest) (*pb.Group, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	group, err := GetGroupByUserAndCourse(in, s.db)
	if err != nil {
		return nil, err
	}
	group.RemoveRemoteIDs()
	return group, nil
}

// GetProviders returns a list of providers
func (s *AutograderService) GetProviders(ctx context.Context, in *pb.Void) (*pb.Providers, error) {
	providers := auth.GetProviders()
	if len(providers.GetProviders()) < 1 {
		s.logger.Error("found no enabled SCM providers")
		return nil, status.Errorf(codes.NotFound, "found no enabled SCM providers")
	}
	return providers, nil
}

// GetOrganizations returns a list of directories for a course
func (s *AutograderService) GetOrganizations(ctx context.Context, in *pb.ActionRequest) (*pb.Organizations, error) {
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	_, scm, err := s.getUserAndSCM(ctx, in.Provider, false)
	if err != nil {
		return nil, err
	}
	return ListOrganizations(ctx, s.db, scm)
}

// GetRepository is not yet implemented
func (s *AutograderService) GetRepository(ctx context.Context, in *pb.RepositoryRequest) (*pb.Repository, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
