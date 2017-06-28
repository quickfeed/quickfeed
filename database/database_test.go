package database_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/autograde/aguis/database"
	"github.com/labstack/gommon/log"
)

const (
	userID         = 0
	accessToken    = "secret"
	newAccessToken = "SECRET"
)

func TestNewStructOnFileDB(t *testing.T) {
	const (
		dbpath    = "agdb_test.db"
		oldSuffix = "_old"
	)

	logger := log.New("")
	logger.SetOutput(ioutil.Discard)

	dbfile := tempFile(dbpath)
	// Remove existing database and continue on error if file did not exist.
	if err := os.Remove(dbfile); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}

	// Create new database.
	db, err := database.NewStructDB(dbfile, false, logger)
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
	db, err = database.NewStructDB(dbfile, false, logger)
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
	db, err = database.NewStructDB(dbfile, true, logger)
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

func TestUpdateAccessToken(t *testing.T) {
	logger := log.New("")
	logger.SetOutput(ioutil.Discard)

	// Create new database.
	db, err := database.NewStructDB("", false, logger)
	if err != nil {
		t.Error(err)
	}

	// Add new user.
	if _, err := db.GetUserWithGithubID(userID, accessToken); err != nil {
		t.Error(err)
	}

	// Try to get new user with new access token.
	user, err := db.GetUserWithGithubID(userID, newAccessToken)
	if err != nil {
		t.Error(err)
	}

	if user.AccessToken != newAccessToken {
		t.Errorf("have '%s' access token want '%s'", user.AccessToken, newAccessToken)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func tempFile(name string) string {
	return os.TempDir() + string(filepath.Separator) + name
}
