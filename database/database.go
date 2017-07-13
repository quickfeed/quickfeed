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
	GetUsers() (*[]models.User, error)

	// SetAdmin makes an existing user an administrator.
	SetAdmin(uint64) error

	CreateCourse(*models.Course) error
	GetCourses() (*[]models.Course, error)
	GetCourse(uint64) (*models.Course, error)
	UpdateCourse(*models.Course) error

	CreateEnrollment(*models.Enrollment) error
	AcceptEnrollment(uint64) error
	RejectEnrollment(uint64) error
	GetEnrollmentsByUser(id uint64, statuses ...uint) ([]*models.Enrollment, error)
	GetEnrollmentsByCourse(id uint64, statuses ...uint) ([]*models.Enrollment, error)

	GetAssignments(uint64) (*[]models.Assignment, error)
	CreateAssignment(*models.Assignment) error
}
