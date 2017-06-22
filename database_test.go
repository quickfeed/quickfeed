package aguis_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/autograde/aguis"
	"github.com/go-kit/kit/log"
)

func tempFile(name string) string {
	return os.TempDir() + string(filepath.Separator) + name
}

func TestNewStructOnFileDB(t *testing.T) {
	dbpath := tempFile("agdb_test.db")
	// Remove existing database and continue on error if file did not exist.
	if err := os.Remove(dbpath); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}

	// Create new database.
	db, err := aguis.NewStructOnFileDB(dbpath, false, log.NewNopLogger())
	if err != nil {
		t.Error(err)
	}
	// Verify that the database was created.
	if !fileExists(dbpath) {
		t.Error("database not created")
	}

	user, err := db.GetUserWithGithubID(123)
	if err != nil {
		t.Error(err)
	}

	// Load previously created database.
	db, err = aguis.NewStructOnFileDB(dbpath, false, log.NewNopLogger())
	if err != nil {
		t.Error(err)
	}

	sameUser, err := db.GetUserWithGithubID(123)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(user, sameUser) {
		t.Errorf("have %+v want %+v", sameUser, user)
	}

	// Create new database truncating any existing database.
	db, err = aguis.NewStructOnFileDB(dbpath, true, log.NewNopLogger())
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
	if err := os.Remove(dbpath); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
	// Remove old database and continue on error if file did not exist.
	if err := os.Remove(dbpath + "_old"); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
