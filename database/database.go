package database

import (
	pb "github.com/autograde/quickfeed/ag"
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
	// GetUserByCourse returns the owner of the given login
	// with preloaded course matching the given query.
	GetUserByCourse(*pb.Course, string) (*pb.User, *pb.Course, error)
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
	GetCoursesByUser(userID uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Course, error)
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
	GetAssignmentsByCourse(uint64, bool) ([]*pb.Assignment, error)
	// UpdateAssignments updates the specified list of assignments.
	UpdateAssignments([]*pb.Assignment) error
	// CreateBenchmark creates a new grading benchmark.
	CreateBenchmark(*pb.GradingBenchmark) error
	// UpdateBenchmark updates the given benchmark.
	UpdateBenchmark(*pb.GradingBenchmark) error
	// DeleteBenchmark deletes the given benchmark.
	DeleteBenchmark(*pb.GradingBenchmark) error
	// CreateCriterion creates a new grading criterion.
	CreateCriterion(*pb.GradingCriterion) error
	// UpdateCriterion updates the given criterion.
	UpdateCriterion(*pb.GradingCriterion) error
	// DeleteCriterion deletes the given criterion.
	DeleteCriterion(*pb.GradingCriterion) error

	// CreateSubmission creates a new submission record or updates the most
	// recent submission, as defined by the provided submissionQuery.
	// The submissionQuery must always specify the assignment, and may specify the ID of
	// either an individual student or a group, but not both.
	CreateSubmission(*pb.Submission) error
	// GetSubmission returns a single submission matching the given query.
	GetSubmission(query *pb.Submission) (*pb.Submission, error)
	// GetLastSubmissions returns a list of submission entries for the given course, matching the given query.
	GetLastSubmissions(courseID uint64, query *pb.Submission) ([]*pb.Submission, error)
	// GetSubmissions returns all submissions matching the query.
	GetSubmissions(*pb.Submission) ([]*pb.Submission, error)
	// GetAssignmentsWithSubmissions returns a list of assignments with the latest submissions for the given course.
	GetAssignmentsWithSubmissions(courseID uint64, requestType pb.SubmissionsForCourseRequest_Type, withBuildInfo bool) ([]*pb.Assignment, error)
	// UpdateSubmission updates the specified submission with approved or not approved.
	UpdateSubmission(*pb.Submission) error
	// UpdateSubmissions releases and/or approves all submissions with a certain score
	UpdateSubmissions(uint64, *pb.Submission) error
	// GetReview returns a single review matching the given query.
	GetReview(query *pb.Review) (*pb.Review, error)
	// CreateReview adds a new submission review.
	CreateReview(*pb.Review) error
	// UpdateReview updates the given review.
	UpdateReview(*pb.Review) error
	// DeleteReview removes all review records matching the query.
	DeleteReview(*pb.Review) error
	// GetBenchmarks return all benchmarks and criteria for an assignmend
	GetBenchmarks(*pb.Assignment) ([]*pb.GradingBenchmark, error)
	// CreateRepository creates a new repository.
	CreateRepository(repo *pb.Repository) error
	// GetRepositories returns repositories that match the given query.
	GetRepositories(query *pb.Repository) ([]*pb.Repository, error)
	// DeleteRepository deletes repository for the given remote provider's ID.
	DeleteRepository(remoteID uint64) error

	// UpdateSlipDays updates used slipdays for the given course enrollment
	UpdateSlipDays([]*pb.UsedSlipDays) error
}
