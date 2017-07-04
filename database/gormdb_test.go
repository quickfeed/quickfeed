package database_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func setup(t *testing.T) (*database.GormDB, func()) {
	const (
		driver = "sqlite3"
		prefix = "testdb"
	)

	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	db, err := database.NewGormDB(driver, f.Name(), true)
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

func TestGormDBGetUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetUser(10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetUsers(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetUsers(); err != nil {
		t.Errorf("have error '%v' wanted '%v'", err, nil)
	}
}

func TestGormDBGetUserByRemoteIdentity(t *testing.T) {
	const (
		initialToken = "123"
		newToken     = "ABC"
		provider     = "github"
		remoteID     = 10
	)

	wantUser1 := &models.User{
		ID: 1,
		RemoteIdentities: []models.RemoteIdentity{{
			ID:          1,
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: initialToken,
			UserID:      1,
		}},
	}

	wantUser2 := &models.User{
		ID: 1,
		RemoteIdentities: []models.RemoteIdentity{{
			ID:          1,
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: newToken,
			UserID:      1,
		}},
	}

	db, cleanup := setup(t)
	defer cleanup()

	user1, err := db.GetUserByRemoteIdentity(provider, remoteID, initialToken)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(user1, wantUser1) {
		t.Errorf("have user %+v want %+v", user1, wantUser1)
	}

	user2, err := db.GetUserByRemoteIdentity(provider, remoteID, newToken)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(user2, wantUser2) {
		t.Errorf("have user %+v want %+v", user2, wantUser2)
	}
}
