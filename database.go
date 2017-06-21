package aguis

import (
	"encoding/gob"
	"errors"
	"os"
	"sync"
)

// User represents a user account.
type User struct {
	ID       int
	GithubID int
}

// UserDatabase contains methods for manipulating a database user.
type UserDatabase interface {
	GetUsers() (map[int]*User, error)
	GetUserWithGithubID(int) (*User, error)
}

// NewStructOnFileDB creates a new database which saves the whole database to a
// file on every change.
func NewStructOnFileDB(path string, truncate bool) (*StructOnFileDB, error) {
	newDB := truncate || !fileExists(path)

	if !newDB {
		f, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		var db StructOnFileDB
		dec := gob.NewDecoder(f)
		if err := dec.Decode(&db); err != nil {
			return nil, err
		}
		db.path = path

		return &db, nil
	}

	db := &StructOnFileDB{
		path:  path,
		Users: make(map[int]*User),
	}

	return db, db.save()
}

// StructOnFileDB implements UserDatabase.
type StructOnFileDB struct {
	mu    sync.Mutex
	path  string
	Users map[int]*User
}

// ErrUserNotExist indicates that the user does not exist.
var ErrUserNotExist = errors.New("user does not exist")

// GetUsers returns all the user accounts in the database.
func (db *StructOnFileDB) GetUsers() (map[int]*User, error) {
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
func (db *StructOnFileDB) GetUserWithGithubID(githubID int) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, user := range db.Users {
		if user.GithubID == githubID {
			return user, nil
		}
	}

	user := &User{
		ID:       0,
		GithubID: githubID,
	}

	db.Users[user.ID] = user
	if err := db.save(); err != nil {
		delete(db.Users, user.ID)
		return nil, err
	}

	return user, nil
}

// Caller must hold lock on db.
func (db *StructOnFileDB) save() error {
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
