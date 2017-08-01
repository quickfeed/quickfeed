package database

import "github.com/autograde/aguis/models"

// Database contains methods for manipulating the database.
type Database interface {
	CreateUserFromRemoteIdentity(provider string, remoteID uint64, accessToken string) (*models.User, error)
	AssociateUserWithRemoteIdentity(userID uint64, provider string, remoteID uint64, accessToken string) error

	GetUser(uint64) (*models.User, error)
	// GetUserByRemoteIdentity gets an user by a remote identity and updates the access token.
	// TODO: The update access token functionality should be split into its own method.
	GetUserByRemoteIdentity(provider string, id uint64, accessToken string) (*models.User, error)
	GetUsers(ids ...uint64) ([]*models.User, error)

	// SetAdmin makes an existing user an administrator.
	SetAdmin(uint64) error

	CreateCourse(*models.Course) error
	GetCourse(uint64) (*models.Course, error)
	GetCourses(ids ...uint64) ([]*models.Course, error)
	GetCoursesByUser(id uint64, statuses ...uint) ([]*models.Course, error)
	UpdateCourse(*models.Course) error

	CreateEnrollment(*models.Enrollment) error
	AcceptEnrollment(uint64) error
	RejectEnrollment(uint64) error
	GetEnrollmentsByCourse(id uint64, statuses ...uint) ([]*models.Enrollment, error)
	GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*models.Enrollment, error)

	CreateAssignment(*models.Assignment) error
	GetAssignmentsByCourse(uint64) ([]*models.Assignment, error)

	CreateSubmission(*models.Submission) error
	GetSubmissionForUser(assignmentID uint64, userID uint64) (*models.Submission, error)
	GetSubmissions(courseID uint64, userID uint64) ([]*models.Submission, error)

	CreateGroup(*models.Group) error
	GetGroup(id uint64) (*models.Group, error)
	GetGroups(cid uint64) ([]*models.Group, error)
	UpdateGroupStatus(*models.Group) error
}
