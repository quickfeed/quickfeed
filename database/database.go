package database

import "github.com/autograde/aguis/models"

// Database contains methods for manipulating the database.
type Database interface {
	GetUser(uint64) (*models.User, error)
	GetUsers() (*[]models.User, error)
	GetUserByRemoteIdentity(string, uint64, string) (*models.User, error)

	CreateCourse(string, string) error
	GetCourses() (*[]models.Course, error)
}
