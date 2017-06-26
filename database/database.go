package database

import (
	"encoding/gob"
	"errors"
	"os"
	"sync"

	"github.com/go-kit/kit/log"
)

// User represents a user account.
type User struct {
	ID          int
	GithubID    int
	AccessToken string
}

// UserDatabase contains methods for manipulating a database user.
type UserDatabase interface {
	GetUser(int) (*User, error)
	GetUsers() (map[int]*User, error)
	GetUserWithGithubID(int, string) (*User, error)
}

// NewStructDB creates a new database which saves the whole database to a file
// on every change. If no path is set, the database will operate in memory only.
func NewStructDB(path string, truncate bool, logger log.Logger) (*StructDB, error) {
	if path == "" {
		return &StructDB{
			// Leave path unset to indicate in memory DB.
			Users:  make(map[int]*User),
			logger: logger,
		}, nil
	}

	newDB := truncate || !fileExists(path)

	if !newDB {
		f, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		var db StructDB
		dec := gob.NewDecoder(f)
		if err := dec.Decode(&db); err != nil {
			return nil, err
		}
		db.path = path
		db.logger = logger

		return &db, nil
	}

	db := &StructDB{
		path:   path,
		Users:  make(map[int]*User),
		logger: logger,
	}

	return db, db.save()
}

// StructDB implements UserDatabase.
type StructDB struct {
	mu    sync.Mutex
	path  string
	Users map[int]*User

	logger log.Logger
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
			db.logger.Log(
				"userid", user.ID,
				"githubid", user.GithubID,
				"msg", "user found",
				"new", false,
			)
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
		db.logger.Log(
			"userid", user.ID,
			"githubid", user.GithubID,
			"msg", "could not persist user to database",
			"err", err.Error(),
		)
		return nil, err
	}

	db.logger.Log(
		"userid", user.ID,
		"githubid", user.GithubID,
		"msg", "user found",
		"new", true,
	)

	return user, nil
}

// Caller must hold lock on db.
func (db *StructDB) save() error {
	// Don't write to disk if in memory DB.
	if db.path == "" {
		return nil
	}

	oldPath := db.path + "_old"

	// Move existing database and continue on error if file did not exist.
	if err := os.Rename(db.path, oldPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	f, err := os.Create(db.path)
	defer f.Close()
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(f)
	if err := enc.Encode(db); err != nil {
		return err
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
