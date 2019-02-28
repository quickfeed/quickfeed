package database

import "github.com/autograde/aguis/models"

// Database contains methods for manipulating the database.
type Database interface {
	GetRemoteIdentity(provider string, rid uint64) (*models.RemoteIdentity, error)

	CreateUserFromRemoteIdentity(*models.User, *models.RemoteIdentity) error
	AssociateUserWithRemoteIdentity(uid uint64, provider string, rid uint64, accessToken string) error
	// GetUserByRemoteIdentity returns the user for the given remote identity.
	// The supplied remote identity must contain Provider and RemoteID.
	GetUserByRemoteIdentity(*models.RemoteIdentity) (*models.User, error)
	// UpdateAccessToken updates the access token for the given remote identity.
	// The supplied remote identity must contain Provider, RemoteID and AccessToken.
	UpdateAccessToken(*models.RemoteIdentity) error

	// GetUser returns the user for the given user ID,
	// including the user's remote identities.
	GetUser(uint64) (*models.User, error)
	// GetUsers returns the users for the given set of user IDs.
	// The returned users's remote identities are included if withRemoteIDs
	// is true, otherwise remote identities won't be include.
	// Note: Remote identities holds the user's access token and should not
	// be returned to the frontend.
	GetUsers(withRemoteIDs bool, userIDs ...uint64) ([]*models.User, error)
	// UpdateUser updates the user's details, excluding remote identities.
	UpdateUser(*models.User) error

	// SetAdmin makes an existing user an administrator. The admin role is allowed to
	// create courses, so it makes sense that teachers are made admins.
	SetAdmin(uint64) error

	CreateCourse(uint64, *models.Course) error
	GetCourse(uint64) (*models.Course, error)
	GetCourseByDirectoryID(did uint64) (*models.Course, error)
	GetCourses(...uint64) ([]*models.Course, error)
	GetCoursesByUser(uid uint64, statuses ...uint) ([]*models.Course, error)
	UpdateCourse(*models.Course) error

	CreateEnrollment(*models.Enrollment) error
	RejectEnrollment(uid uint64, cid uint64) error
	EnrollStudent(uid uint64, cid uint64) error
	EnrollTeacher(uid uint64, cid uint64) error
	SetPendingEnrollment(uid, cid uint64) error
	GetEnrollmentsByCourse(cid uint64, statuses ...uint) ([]*models.Enrollment, error)
	GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*models.Enrollment, error)

	// CreateAssignment creates a new or updates an existing assignment.
	CreateAssignment(*models.Assignment) error
	// UpdateAssignments updates the specified list of assignments.
	UpdateAssignments([]*models.Assignment) error
	GetAssignmentsByCourse(uint64) ([]*models.Assignment, error)
	GetNextAssignment(cid, uid, gid uint64) (*models.Assignment, error)

	CreateSubmission(*models.Submission) error
	GetSubmissionForUser(aid uint64, uid uint64) (*models.Submission, error)
	GetSubmissionForGroup(aid uint64, gid uint64) (*models.Submission, error)
	GetSubmissions(cid uint64, uid uint64) ([]*models.Submission, error)
	GetGroupSubmissions(cid uint64, gid uint64) ([]*models.Submission, error)
	GetSubmissionsByID(sid uint64) (*models.Submission, error)
	UpdateSubmissionByID(sid uint64, approved bool) error

	CreateGroup(*models.Group) error
	// GetGroup returns the group with the specified group id.
	// The returned users's remote identities are included if withRemoteIDs
	// is true, otherwise remote identities won't be include.
	// Note: Remote identities holds the user's access token and should not
	// be returned to the frontend.
	GetGroup(withRemoteIDs bool, groupID uint64) (*models.Group, error)
	GetGroupsByCourse(cid uint64) ([]*models.Group, error)
	UpdateGroupStatus(*models.Group) error
	UpdateGroup(group *models.Group) error
	DeleteGroup(uint64) error

	CreateRepository(repo *models.Repository) error
	GetRepository(uint64) (*models.Repository, error)
	GetRepositoriesByDirectory(uint64) ([]*models.Repository, error)
	GetRepositoriesByCourseIDAndType(uint64, models.RepoType) ([]*models.Repository, error)
	GetRepositoriesByCourseIDandUserID(uint64, uint64) (*models.Repository, error)
	GetRepoByCourseIDUserIDandType(uint64, uint64, models.RepoType) (*models.Repository, error)
}
