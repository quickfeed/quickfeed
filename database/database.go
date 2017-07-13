package database

import "github.com/autograde/aguis/models"

// Database contains methods for manipulating the database.
type Database interface {
	NewUserFromRemoteIdentity(provider string, remoteID uint64, accessToken string) (*models.User, error)
	AssociateUserWithRemoteIdentity(userID uint64, provider string, remoteID uint64, accessToken string) error

	GetUser(uint64) (*models.User, error)
	// GetUserByRemoteIdentity gets an user by a remote identity and updates the access token.
	// TODO: The update access token functionality should be split into its own method.
	GetUserByRemoteIdentity(provider string, id uint64, accessToken string) (*models.User, error)
	GetUsers() (*[]models.User, error)

	CreateCourse(*models.Course) error
	GetCourses() (*[]models.Course, error)
	GetCoursesForUser(id uint64) (*[]models.Course, error)
	EnrollUserInCourse(userID, courseID uint64) error

	GetAssignments(id uint64) (*[]models.Assignment, error)
	CreateAssignment(assignment *models.Assignment) error
}
