package database_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/autograde/aguis/database"
	"github.com/go-kit/kit/log"
)

func tempFile(name string) string {
	return os.TempDir() + string(filepath.Separator) + name
}

func TestNewStructOnFileDB(t *testing.T) {
	const (
		dbpath      = "agdb_test.db"
		oldSuffix   = "_old"
		userID      = 0
		accessToken = "secret"
	)

	dbfile := tempFile(dbpath)
	// Remove existing database and continue on error if file did not exist.
	if err := os.Remove(dbfile); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}

	// Create new database.
	db, err := database.NewStructDB(dbfile, false, log.NewNopLogger())
	if err != nil {
		t.Error(err)
	}
	// Verify that the database was created.
	if !fileExists(dbfile) {
		t.Error("database not created")
	}

	user, err := db.GetUserWithGithubID(userID, accessToken)
	if err != nil {
		t.Error(err)
	}

	// Load previously created database.
	db, err = database.NewStructDB(dbfile, false, log.NewNopLogger())
	if err != nil {
		t.Error(err)
	}

	sameUser, err := db.GetUserWithGithubID(userID, accessToken)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(user, sameUser) {
		t.Errorf("have %+v want %+v", sameUser, user)
	}

	// Create new database truncating any existing database.
	db, err = database.NewStructDB(dbfile, true, log.NewNopLogger())
	if err != nil {
		t.Error(err)
	}

	users, err := db.GetUsers()
	if err != nil {
		t.Error(err)
	}
	if len(users) != 0 {
		t.Errorf("have %d users want 0 users", len(users))
	}

	// Remove current database and continue on error if file did not exist.
	if err := os.Remove(dbfile); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
	// Remove old database and continue on error if file did not exist.
	if err := os.Remove(dbfile + oldSuffix); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
