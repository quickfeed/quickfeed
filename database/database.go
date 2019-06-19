package database

import (
	pb "github.com/autograde/aguis/ag"
)

// Database contains methods for manipulating the database.
type Database interface {
	GetRemoteIdentity(provider string, rid uint64) (*pb.RemoteIdentity, error)

	CreateUserFromRemoteIdentity(*pb.User, *pb.RemoteIdentity) error
	AssociateUserWithRemoteIdentity(uid uint64, provider string, rid uint64, accessToken string) error
	// GetUserByRemoteIdentity returns the user for the given remote identity.
	// The supplied remote identity must contain Provider and RemoteID.
	GetUserByRemoteIdentity(*pb.RemoteIdentity) (*pb.User, error)
	// UpdateAccessToken updates the access token for the given remote identity.
	// The supplied remote identity must contain Provider, RemoteID and AccessToken.
	UpdateAccessToken(*pb.RemoteIdentity) error
	// GetUser returns the user for the given user ID,
	// including the user's remote identities.
	GetUser(uint64) (*pb.User, error)
	// GetUsers returns the users for the given set of user IDs.
	GetUsers(...uint64) ([]*pb.User, error)
	// UpdateUser updates the user's details, excluding remote identities.
	UpdateUser(*pb.User) error

	// SetAdmin makes an existing user an administrator. The admin role is allowed to
	// create courses, so it makes sense that teachers are made admins.
	SetAdmin(uint64) error

	CreateCourse(uint64, *pb.Course) error
	GetCourse(uint64) (*pb.Course, error)
	GetCourseByOrganizationID(did uint64) (*pb.Course, error)
	GetCourses(...uint64) ([]*pb.Course, error)
	GetCoursesByUser(uid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Course, error)
	UpdateCourse(*pb.Course) error

	CreateEnrollment(*pb.Enrollment) error
	RejectEnrollment(uid uint64, cid uint64) error
	EnrollStudent(uid uint64, cid uint64) error
	EnrollTeacher(uid uint64, cid uint64) error
	SetPendingEnrollment(uid, cid uint64) error
	// UpdateGroupEnrollment is used to reset group ID when previously aproved group is
	// being removed or a user is removed from the group
	UpdateGroupEnrollment(uid, cid uint64) error

	GetEnrollmentsByCourse(cid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error)
	GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*pb.Enrollment, error)
	// CreateAssignment creates a new or updates an existing assignment.
	CreateAssignment(*pb.Assignment) error
	// UpdateAssignments updates the specified list of assignments.
	UpdateAssignments([]*pb.Assignment) error
	GetAssignmentsByCourse(uint64) ([]*pb.Assignment, error)
	GetNextAssignment(cid, uid, gid uint64) (*pb.Assignment, error)

	CreateSubmission(*pb.Submission) error
	GetSubmission(query *pb.Submission) (*pb.Submission, error)

	GetSubmissions(cid uint64, uid uint64) ([]*pb.Submission, error)
	GetGroupSubmissions(cid uint64, gid uint64) ([]*pb.Submission, error)
	// UpdateSubmission updates the specified submission with approved or not approved.
	UpdateSubmission(submissionID uint64, approved bool) error

	CreateGroup(*pb.Group) error
	// GetGroup returns the group with the specified group id.
	GetGroup(uint64) (*pb.Group, error)
	GetGroupsByCourse(cid uint64) ([]*pb.Group, error)
	UpdateGroupStatus(*pb.Group) error
	UpdateGroup(group *pb.Group) error
	DeleteGroup(uint64) error

	// CreateRepository creates a new repository.
	CreateRepository(repo *pb.Repository) error
	// GetRepository returns the repository for the SCM provider's repository ID.
	GetRepository(uint64) (*pb.Repository, error)
	// GetRepositories returns repositories that match the given query.
	GetRepositories(query *pb.Repository) ([]*pb.Repository, error)
}
