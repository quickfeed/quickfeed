package database

// Database contains methods for manipulating the database.
type Database interface {
	GetUser(int) (*User, error)
	GetUsers() (map[int]*User, error)
	GetUserWithGithubID(int, string) (*User, error)

}
