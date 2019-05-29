package grpcservice

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/autograde/aguis/web"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
)

// AutograderService holds references to the database and shared structures
type AutograderService struct {
	db   *database.GormDB
	scms map[string]scm.SCM
	bh   web.BaseHookOptions
}

// NewAutograderService is an AutograderService constructor
func NewAutograderService(db *database.GormDB, scms map[string]scm.SCM, bh web.BaseHookOptions) *AutograderService {
	return &AutograderService{
		db:   db,
		scms: scms,
		bh:   bh,
	}

}

// GetRepositoryURL returns a repository of requested type
func (s *AutograderService) GetRepositoryURL(ctx context.Context, in *pb.RepositoryRequest) (*pb.URLResponse, error) {
	if !in.IsValidRepoRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	currentUser, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	return web.GetRepositoryURL(currentUser, in, s.db)
}

// GetUser returns information about the user excluding remote identity
func (s *AutograderService) GetUser(ctx context.Context, in *pb.RecordRequest) (*pb.User, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := web.GetUser(in, s.db)
	if err != nil {
		return nil, err
	}
	usr.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
	return usr, nil
}

// GetUsers returns a list of all existing users
func (s *AutograderService) GetUsers(ctx context.Context, in *pb.Void) (*pb.Users, error) {
	usrs, err := web.GetUsers(s.db)
	if err != nil {
		return nil, err
	}
	if len(usrs.Users) > 0 {
		for _, usr := range usrs.Users {
			usr.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
		}
	}
	return usrs, nil
}

// UpdateUser inserts the new user information and returns the updated user
func (s *AutograderService) UpdateUser(ctx context.Context, in *pb.User) (*pb.User, error) {
	if !in.IsValidUser() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	currentUser, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	usr, err := web.PatchUser(currentUser, in, s.db)
	if err != nil {
		return nil, err
	}
	usr.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
	return usr, nil
}

// CreateCourse used to make a new course
func (s *AutograderService) CreateCourse(ctx context.Context, in *pb.Course) (*pb.Course, error) {
	if !in.IsValidCourse() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	// only teacher with admin rights can create a new course
	if !usr.IsAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "user must be admin to create a new course")
	}
	scm, err := getSCM(ctx, s.scms, s.db, in.Provider)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()
	// make sure that the current user is set as course creator
	in.CourseCreatorID = usr.GetID()
	return web.NewCourse(ctx, in, s.db, scm, s.bh)
}

// GetCourse returns course information
func (s *AutograderService) GetCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Course, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.GetCourse(in, s.db)
}

// UpdateCourse is used to change the course information
func (s *AutograderService) UpdateCourse(ctx context.Context, in *pb.Course) (*pb.Void, error) {
	if !in.IsValidCourse() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	scm, err := getSCM(ctx, s.scms, s.db, in.Provider)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()
	return &pb.Void{}, web.UpdateCourse(ctx, in, s.db, scm)
}

// GetCourses returns a list with all courses
func (s *AutograderService) GetCourses(ctx context.Context, in *pb.Void) (*pb.Courses, error) {
	return web.ListCourses(s.db)
}

// GetCoursesWithEnrollment returns all courses for the user where user enrollment is of a certain type
func (s *AutograderService) GetCoursesWithEnrollment(ctx context.Context, in *pb.RecordRequest) (*pb.Courses, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.ListCoursesWithEnrollment(in, s.db)
}

// GetAssignments returns a list of assignments
func (s *AutograderService) GetAssignments(ctx context.Context, in *pb.RecordRequest) (*pb.Assignments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.ListAssignments(in, s.db)
}

// GetEnrollmentsByCourse returns all existing enrollments for a course
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Enrollments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.GetEnrollmentsByCourse(in, s.db)
}

// CreateEnrollment inserts a new student enrollment
func (s *AutograderService) CreateEnrollment(ctx context.Context, in *pb.ActionRequest) (*pb.Void, error) {
	if !in.IsValidEnrollment() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return &pb.Void{}, web.CreateEnrollment(in, s.db)
}

// UpdateEnrollment is used to change an enrollment status of a student
func (s *AutograderService) UpdateEnrollment(ctx context.Context, in *pb.ActionRequest) (*pb.Void, error) {
	if !in.IsValidEnrollment() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !usr.IsAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}
	crs, err := web.GetCourse(&pb.RecordRequest{ID: in.CourseID}, s.db)
	if err != nil {
		return nil, err
	}
	scm, err := getSCM(ctx, s.scms, s.db, crs.Provider)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()
	return &pb.Void{}, web.UpdateEnrollment(ctx, in, s.db, scm)
}

// GetSelf returns information about the user with user ID sent in the context
func (s *AutograderService) GetSelf(ctx context.Context, in *pb.Void) (*pb.User, error) {
	currentUser, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	user, err := web.GetUser(&pb.RecordRequest{ID: currentUser.ID}, s.db)
	if err != nil {
		return nil, err
	}
	user.RemoteIdentities = nil
	return user, nil
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
	for _, user := range group.Users {
		user.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
	}
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
	for _, group := range groups.Groups {
		for _, user := range group.Users {
			user.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
		}
	}
	return groups, nil
}

// CreateGroup makes a new group
func (s *AutograderService) CreateGroup(ctx context.Context, in *pb.Group) (*pb.Group, error) {
	if !in.IsValidGroup() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	group, err := web.NewGroup(in, s.db, usr)
	if err != nil {
		return nil, err
	}
	for _, user := range group.Users {
		user.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
	}
	return group, nil
}

// UpdateGroup is called by UpdateGroup client method, changes group information
func (s *AutograderService) UpdateGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	if !in.IsValidGroup() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "invalid user ID")
	}
	crs, err := web.GetCourse(&pb.RecordRequest{ID: in.CourseID}, s.db)
	if err != nil {
		return nil, err
	}
	scm, err := getSCM(ctx, s.scms, s.db, crs.Provider)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()
	return &pb.Void{}, web.UpdateGroup(ctx, in, s.db, scm, usr)
}

//TODO(vera): this should not be called anymore, but must be removed from the proto first

// UpdateGroupStatus is called by UpdateGroupStatus client method, changes group enrollment status
func (s *AutograderService) UpdateGroupStatus(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	if !in.IsValidGroup() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	crs, err := web.GetCourse(&pb.RecordRequest{ID: in.CourseID}, s.db)
	if err != nil {
		return nil, err
	}
	scm, err := getSCM(ctx, s.scms, s.db, crs.Provider)
	if err != nil {
		return nil, err
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
	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
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

// GetCourseInformationURL returns URL of repository containing information about the course
func (s *AutograderService) GetCourseInformationURL(ctx context.Context, in *pb.RecordRequest) (*pb.URLResponse, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	return web.GetCourseInformationURL(in, s.db)
}

// RefreshCourse returns latest information about the course
func (s *AutograderService) RefreshCourse(ctx context.Context, in *pb.RecordRequest) (*pb.Assignments, error) {
	if !in.IsValidRequest() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
	}
	usr, err := getCurrentUser(ctx, s.db)
	if err != nil {
		return nil, err
	}
	crs, err := web.GetCourse(in, s.db)
	if err != nil {
		return nil, err
	}
	scm, err := getSCM(ctx, s.scms, s.db, crs.Provider)
	if err != nil {
		return nil, err
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
	for _, user := range group.Users {
		user.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
	}
	return group, nil
}

// GetProviders returns a list of providers
func (s *AutograderService) GetProviders(ctx context.Context, in *pb.Void) (*pb.Providers, error) {
	return web.GetProviders()
}

// GetDirectories returns a list of directories for a course
func (s *AutograderService) GetDirectories(ctx context.Context, in *pb.DirectoryRequest) (*pb.Directories, error) {
	scm, err := getSCM(ctx, s.scms, s.db, in.Provider)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, web.MaxWait)
	defer cancel()
	return web.ListDirectories(ctx, s.db, scm)
}

// GetRepository is not yet implemented
func (s *AutograderService) GetRepository(ctx context.Context, in *pb.RepositoryRequest) (*pb.Repository, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
