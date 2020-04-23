package database

import (
	pb "github.com/autograde/aguis/ag"
)

// Database contains methods for manipulating the database.
type Database interface {
	// CreateUserFromRemoteIdentity creates new user record from remote identity, sets user with ID 1 as admin.
	CreateUserFromRemoteIdentity(*pb.User, *pb.RemoteIdentity) error
	// AssociateUserWithRemoteIdentity associates user with the given remote identity.
	AssociateUserWithRemoteIdentity(userID uint64, provider string, remoteID uint64, accessToken string) error
	// UpdateAccessToken updates the access token for the given remote identity.
	// The supplied remote identity must contain Provider, RemoteID and AccessToken.
	UpdateAccessToken(*pb.RemoteIdentity) error
	// GetUserByRemoteIdentity returns the user for the given remote identity.
	// The supplied remote identity must contain Provider and RemoteID.
	GetUserByRemoteIdentity(*pb.RemoteIdentity) (*pb.User, error)

	// GetUser returns the given user, including remote identities.
	GetUser(uint64) (*pb.User, error)
	// GetUserWithEnrollments returns the user by ID with preloaded user enrollments.
	GetUserWithEnrollments(uint64) (*pb.User, error)
	// GetUsers returns the users for the given set of user IDs.
	GetUsers(...uint64) ([]*pb.User, error)
	// UpdateUser updates the user's details, excluding remote identities.
	UpdateUser(*pb.User) error

	// CreateCourse creates a new course if user with given ID is admin, enrolls user as course teacher.
	CreateCourse(uint64, *pb.Course) error
	// GetCourse fetches course by ID. If withInfo is true, preloads course
	// assignments, active enrollments and groups.
	GetCourse(uint64, bool) (*pb.Course, error)
	// GetCourseByOrganizationID fetches course by organization ID.
	GetCourseByOrganizationID(organizationID uint64) (*pb.Course, error)
	// GetCourses returns a list of courses. If one or more course IDs are provided,
	// the corresponding courses are returned. Otherwise, all courses are returned.
	GetCourses(...uint64) ([]*pb.Course, error)
	// GetCoursesByUser returns all courses (with enrollment status)
	// for the given user id.
	// If enrollment statuses is provided, the set of courses returned
	// is filtered according to these enrollment statuses.
	GetCoursesByUser(userID uint64, states []pb.Enrollment_DisplayState, statuses ...pb.Enrollment_UserStatus) ([]*pb.Course, error)
	// UpdateCourse updates course information.
	UpdateCourse(*pb.Course) error

	// CreateEnrollment creates a new pending enrollment.
	CreateEnrollment(*pb.Enrollment) error
	// RejectEnrollment removes the user enrollment from the database
	RejectEnrollment(userID, courseID uint64) error
	// UpdateEnrollmentStatus changes status of the course enrollment for the given user and course.
	UpdateEnrollment(*pb.Enrollment) error
	// GetEnrollmentByCourseAndUser returns a user enrollment for the given course ID.
	GetEnrollmentByCourseAndUser(courseID uint64, userID uint64) (*pb.Enrollment, error)
	// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
	GetEnrollmentsByCourse(courseID uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error)
	// GetEnrollmentsByUser fetches all enrollments for the given user
	GetEnrollmentsByUser(userID uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error)

	// CreateGroup creates a new group and assign users to newly created group.
	CreateGroup(*pb.Group) error
	// UpdateGroup updates a group with the specified users and enrollments.
	UpdateGroup(group *pb.Group) error
	// UpdateGroupStatus updates status field of a group.
	UpdateGroupStatus(*pb.Group) error
	// DeleteGroup deletes a group and its corresponding enrollments.
	DeleteGroup(uint64) error
	// GetGroup returns the group with the specified group ID.
	GetGroup(uint64) (*pb.Group, error)
	// GetGroupsByCourse returns the groups for the given course.
	GetGroupsByCourse(courseID uint64, statuses ...pb.Group_GroupStatus) ([]*pb.Group, error)

	// CreateAssignment creates a new or updates an existing assignment.
	CreateAssignment(*pb.Assignment) error
	// GetAssignment returns assignment mathing the given query.
	GetAssignment(query *pb.Assignment) (*pb.Assignment, error)
	// GetAssignmentsByCourse returns a list of all assignments for the given course ID.
	GetAssignmentsByCourse(uint64) ([]*pb.Assignment, error)
	// UpdateAssignments updates the specified list of assignments.
	UpdateAssignments([]*pb.Assignment) error

	// CreateSubmission creates a new submission record or updates the most
	// recent submission, as defined by the provided submissionQuery.
	// The submissionQuery must always specify the assignment, and may specify the ID of
	// either an individual student or a group, but not both.
	CreateSubmission(*pb.Submission) error
	// GetSubmission returns a single submission matching the given query.
	GetSubmission(query *pb.Submission) (*pb.Submission, error)
	// GetSubmissions returns a list of submission entries for the given course, matching the given query.
	GetSubmissions(courseID uint64, query *pb.Submission) ([]*pb.Submission, error)
	// GetCourseSubmissions returns a list of all the latest submissions
	// for every active course assignment for the given course ID
	GetCourseSubmissions(uint64, bool) ([]pb.Submission, error)
	// UpdateSubmission updates the specified submission with approved or not approved.
	UpdateSubmission(submissionID uint64, approved bool) error

	// CreateRepository creates a new repository.
	CreateRepository(repo *pb.Repository) error
	// GetRepository returns the repository for the SCM provider's repository ID.
	GetRepositoryByRemoteID(uint64) (*pb.Repository, error)
	// GetRepositories returns repositories that match the given query.
	GetRepositories(query *pb.Repository) ([]*pb.Repository, error)
	// DeleteRepository deletes repository by the given provider's ID
	DeleteRepositoryByRemoteID(uint64) error
}
