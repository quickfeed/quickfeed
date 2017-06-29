package database

import (
	"errors"

	"github.com/labstack/gommon/log"
)

// User represents a user account.
type User struct {
	ID          int
	GithubID    int
	AccessToken string
}

// ErrUserNotExist indicates that the user does not exist.
var ErrUserNotExist = errors.New("user does not exist")

// GetUser returns the user with the given id.
func (db *StructDB) GetUser(id int) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if user, ok := db.Users[id]; ok {
		return user, nil
	}

	return nil, ErrUserNotExist
}

// GetUsers returns all the user accounts in the database.
func (db *StructDB) GetUsers() (map[int]*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	users := make(map[int]*User, len(db.Users))
	for id, user := range db.Users {
		users[id] = user
	}

	return users, nil
}

// GetUserWithGithubID tries to get the user associated with the given GitHub
// account. If there is no such user, a new user account is created.
func (db *StructDB) GetUserWithGithubID(githubID int, accessToken string) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, user := range db.Users {
		if user.GithubID == githubID {
			user.AccessToken = accessToken
			if err := db.save(); err != nil {
				db.logger.Infoj(log.JSON{
					"userid":   user.ID,
					"githubid": user.GithubID,
					"message":  "could not update access token",
					"err":      err.Error(),
				})
				return nil, err
			}
			db.logger.Infoj(log.JSON{
				"userid":   user.ID,
				"githubid": user.GithubID,
				"message":  "user found",
				"new":      false,
			})
			return user, nil
		}
	}

	user := &User{
		ID:          0,
		GithubID:    githubID,
		AccessToken: accessToken,
	}

	db.Users[user.ID] = user
	if err := db.save(); err != nil {
		delete(db.Users, user.ID)
		db.logger.Infoj(log.JSON{
			"userid":   user.ID,
			"githubid": user.GithubID,
			"message":  "could not persist user to database",
			"err":      err.Error(),
		})
		return nil, err
	}

	db.logger.Infoj(log.JSON{
		"userid":   user.ID,
		"githubid": user.GithubID,
		"message":  "user found",
		"new":      true,
	})

	return user, nil
}
