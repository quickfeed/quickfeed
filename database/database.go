package database

import "github.com/autograde/aguis/models"

// Database contains methods for manipulating the database.
type Database interface {
	NewUserFromRemoteIdentity(provider string, remoteID uint64, accessToken string) (*models.User, error)
	AssociateUserWithRemoteIdentity(userID uint64, provider string, remoteID uint64, accessToken string) error

	GetUser(uint64) (*models.User, error)
	GetUserByRemoteIdentity(provider string, id uint64, accessToken string) (*models.User, error)
	GetUsers() (*[]models.User, error)

	CreateCourse(*models.Course) error
	GetCourses() (*[]models.Course, error)
	GetCoursesForUser(id uint64) (*[]models.Course, error)
	EnrollUserInCourse(userID, courseID uint64) error
}
