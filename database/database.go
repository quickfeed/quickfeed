package database

import "github.com/autograde/aguis/models"

// Database contains methods for manipulating the database.
type Database interface {
	GetRemoteIdentity(provider string, rid uint64) (*models.RemoteIdentity, error)

	CreateUserFromRemoteIdentity(*models.User, *models.RemoteIdentity) error
	AssociateUserWithRemoteIdentity(uid uint64, provider string, rid uint64, accessToken string) error

	GetUser(uint64) (*models.User, error)
	// GetUserByRemoteIdentity gets an user by a remote identity and updates the access token.
	// TODO: The update access token functionality should be split into its own method.
	GetUserByRemoteIdentity(provider string, rid uint64, accessToken string) (*models.User, error)
	GetUsers(...uint64) ([]*models.User, error)
	UpdateUser(*models.User) error

	// SetAdmin makes an existing user an administrator. The admin role is allowed to
	// create courses, so it makes sense that teachers are made admins.
	SetAdmin(uint64) error

	CreateCourse(*models.Course) error
	GetCourse(uint64) (*models.Course, error)
	GetCourseByDirectoryID(did uint64) (*models.Course, error)
	GetCourses(...uint64) ([]*models.Course, error)
	GetCoursesByUser(uid uint64, statuses ...uint) ([]*models.Course, error)
	UpdateCourse(*models.Course) error

	CreateEnrollment(*models.Enrollment) error
	RejectEnrollment(uid uint64, cid uint64) error
	EnrollStudent(uid uint64, cid uint64) error
	EnrollTeacher(uid uint64, cid uint64) error
	GetEnrollmentsByCourse(cid uint64, statuses ...uint) ([]*models.Enrollment, error)
	GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*models.Enrollment, error)

	CreateAssignment(*models.Assignment) error
	GetAssignmentsByCourse(uint64) ([]*models.Assignment, error)

	CreateSubmission(*models.Submission) error
	GetSubmissionForUser(aid uint64, uid uint64) (*models.Submission, error)
	GetSubmissions(cid uint64, uid uint64) ([]*models.Submission, error)
	GetSubmissionsByID(sid uint64) (*models.Submission, error)
	UpdateSubmissionByID(sid uint64, approved bool) error

	CreateGroup(*models.Group) error
	GetGroup(uint64) (*models.Group, error)
	GetGroupsByCourse(cid uint64) ([]*models.Group, error)
	UpdateGroupStatus(*models.Group) error
	UpdateGroup(group *models.Group) error
	DeleteGroup(uint64) error

	CreateRepository(repo *models.Repository) error
	GetRepository(uint64) (*models.Repository, error)
	GetRepositoriesByDirectory(uint64) ([]*models.Repository, error)
}
