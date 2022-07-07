package database

import "github.com/quickfeed/quickfeed/qf/types"

// Database contains methods for manipulating the database.
type Database interface {
	// CreateUserFromRemoteIdentity creates new user record from remote identity, sets user with ID 1 as admin.
	CreateUserFromRemoteIdentity(*types.User, *types.RemoteIdentity) error
	// AssociateUserWithRemoteIdentity associates user with the given remote identity.
	AssociateUserWithRemoteIdentity(userID uint64, provider string, remoteID uint64, accessToken string) error
	// UpdateAccessToken updates the access token for the given remote identity.
	// The supplied remote identity must contain Provider, RemoteID and AccessToken.
	UpdateAccessToken(*types.RemoteIdentity) error
	// GetUserByRemoteIdentity returns the user for the given remote identity.
	// The supplied remote identity must contain Provider and RemoteID.
	GetUserByRemoteIdentity(*types.RemoteIdentity) (*types.User, error)

	// GetUser returns the given user, including remote identities.
	GetUser(uint64) (*types.User, error)
	// GetUserByCourse returns the owner of the given login
	// with preloaded course matching the given query.
	GetUserByCourse(*types.Course, string) (*types.User, *types.Course, error)
	// GetUserWithEnrollments returns the user by ID with preloaded user enrollments.
	GetUserWithEnrollments(uint64) (*types.User, error)
	// GetUsers returns the users for the given set of user IDs.
	GetUsers(...uint64) ([]*types.User, error)
	// UpdateUser updates the user's details, excluding remote identities.
	UpdateUser(*types.User) error

	// CreateCourse creates a new course if user with given ID is admin, enrolls user as course teacher.
	CreateCourse(uint64, *types.Course) error
	// GetCourse fetches course by ID. If withInfo is true, preloads course
	// assignments, active enrollments and groups.
	GetCourse(uint64, bool) (*types.Course, error)
	// GetCourseByOrganizationID fetches course by organization ID.
	GetCourseByOrganizationID(organizationID uint64) (*types.Course, error)
	// GetCourses returns a list of courses. If one or more course IDs are provided,
	// the corresponding courses are returned. Otherwise, all courses are returned.
	GetCourses(...uint64) ([]*types.Course, error)
	// GetCoursesByUser returns all courses (with enrollment status)
	// for the given user id.
	// If enrollment statuses is provided, the set of courses returned
	// is filtered according to these enrollment statuses.
	GetCoursesByUser(userID uint64, statuses ...types.Enrollment_UserStatus) ([]*types.Course, error)
	// GetCourseTeachers returns a list of all teachers in a course.
	GetCourseTeachers(query *types.Course) ([]*types.User, error)
	// UpdateCourse updates course information.
	UpdateCourse(*types.Course) error

	// CreateEnrollment creates a new pending enrollment.
	CreateEnrollment(*types.Enrollment) error
	// RejectEnrollment removes the user enrollment from the database
	RejectEnrollment(userID, courseID uint64) error
	// UpdateEnrollmentStatus changes status of the course enrollment for the given user and course.
	UpdateEnrollment(*types.Enrollment) error
	// GetEnrollmentByCourseAndUser returns a user enrollment for the given course ID.
	GetEnrollmentByCourseAndUser(courseID uint64, userID uint64) (*types.Enrollment, error)
	// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
	GetEnrollmentsByCourse(courseID uint64, statuses ...types.Enrollment_UserStatus) ([]*types.Enrollment, error)
	// GetEnrollmentsByUser fetches all enrollments for the given user
	GetEnrollmentsByUser(userID uint64, statuses ...types.Enrollment_UserStatus) ([]*types.Enrollment, error)

	// CreateGroup creates a new group and assign users to newly created group.
	CreateGroup(*types.Group) error
	// UpdateGroup updates a group with the specified users and enrollments.
	UpdateGroup(group *types.Group) error
	// UpdateGroupStatus updates status field of a group.
	UpdateGroupStatus(*types.Group) error
	// DeleteGroup deletes a group and its corresponding enrollments.
	DeleteGroup(uint64) error
	// GetGroup returns the group with the specified group ID.
	GetGroup(uint64) (*types.Group, error)
	// GetGroupsByCourse returns the groups for the given course.
	GetGroupsByCourse(courseID uint64, statuses ...types.Group_GroupStatus) ([]*types.Group, error)

	// CreateAssignment creates a new or updates an existing assignment.
	CreateAssignment(*types.Assignment) error
	// GetAssignment returns assignment mathing the given query.
	GetAssignment(query *types.Assignment) (*types.Assignment, error)
	// GetAssignmentsByCourse returns a list of all assignments for the given course ID.
	GetAssignmentsByCourse(uint64, bool) ([]*types.Assignment, error)
	// UpdateAssignments updates the specified list of assignments.
	UpdateAssignments([]*types.Assignment) error
	// CreateBenchmark creates a new grading benchmark.
	CreateBenchmark(*types.GradingBenchmark) error
	// UpdateBenchmark updates the given benchmark.
	UpdateBenchmark(*types.GradingBenchmark) error
	// DeleteBenchmark deletes the given benchmark.
	DeleteBenchmark(*types.GradingBenchmark) error
	// CreateCriterion creates a new grading criterion.
	CreateCriterion(*types.GradingCriterion) error
	// UpdateCriterion updates the given criterion.
	UpdateCriterion(*types.GradingCriterion) error
	// DeleteCriterion deletes the given criterion.
	DeleteCriterion(*types.GradingCriterion) error

	// CreateSubmission creates a new submission record or updates the most
	// recent submission, as defined by the provided submissionQuery.
	// The submissionQuery must always specify the assignment, and may specify the ID of
	// either an individual student or a group, but not both.
	CreateSubmission(*types.Submission) error
	// GetSubmission returns a single submission matching the given query.
	GetSubmission(query *types.Submission) (*types.Submission, error)
	// GetLastSubmissions returns a list of submission entries for the given course, matching the given query.
	GetLastSubmissions(courseID uint64, query *types.Submission) ([]*types.Submission, error)
	// GetSubmissions returns all submissions matching the query.
	GetSubmissions(*types.Submission) ([]*types.Submission, error)
	// GetAssignmentsWithSubmissions returns a list of assignments with the latest submissions for the given course.
	GetAssignmentsWithSubmissions(courseID uint64, requestType types.SubmissionsForCourseRequest_Type, withBuildInfo bool) ([]*types.Assignment, error)
	// UpdateSubmission updates the specified submission with approved or not approved.
	UpdateSubmission(*types.Submission) error
	// UpdateSubmissions releases and/or approves all submissions with a certain score
	UpdateSubmissions(uint64, *types.Submission) error
	// GetReview returns a single review matching the given query.
	GetReview(query *types.Review) (*types.Review, error)
	// CreateReview adds a new submission review.
	CreateReview(*types.Review) error
	// UpdateReview updates the given review.
	UpdateReview(*types.Review) error
	// DeleteReview removes all review records matching the query.
	DeleteReview(*types.Review) error
	// GetBenchmarks return all benchmarks and criteria for an assignmend
	GetBenchmarks(*types.Assignment) ([]*types.GradingBenchmark, error)
	// CreateRepository creates a new repository.
	CreateRepository(repo *types.Repository) error
	// GetRepositories returns repositories that match the given query.
	GetRepositories(query *types.Repository) ([]*types.Repository, error)
	// DeleteRepository deletes repository for the given remote provider's ID.
	DeleteRepository(remoteID uint64) error
	// GetRepositoriesWithIssues gets repositories with issues
	GetRepositoriesWithIssues(query *types.Repository) ([]*types.Repository, error)

	// GetTasks returns tasks that match the given query.
	GetTasks(query *types.Task) ([]*types.Task, error)
	// CreateIssues creates a batch of issues
	CreateIssues(issues []*types.Issue) error
	// SynchronizeAssignmentTasks synchronizes all tasks of each assignment in a given course. Returns created, updated and deleted tasks
	SynchronizeAssignmentTasks(course *types.Course, taskMap map[uint32]map[string]*types.Task) ([]*types.Task, []*types.Task, error)
	// CreatePullRequest creates a pull request
	CreatePullRequest(pullRequest *types.PullRequest) error
	// GetPullRequest returns the pull request matching the given query
	GetPullRequest(query *types.PullRequest) (*types.PullRequest, error)
	// HandleMergingPR handles merging a pull request
	HandleMergingPR(query *types.PullRequest) error
	// DeletePullRequest updates the pull request matching the given query
	UpdatePullRequest(pullRequest *types.PullRequest) error

	// UpdateSlipDays updates used slipdays for the given course enrollment
	UpdateSlipDays([]*types.UsedSlipDays) error
}
