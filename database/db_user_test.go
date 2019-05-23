package database_test

import (
	"reflect"
	"testing"

	"github.com/autograde/aguis/models"
)

func TestGormDBUpdateAccessToken(t *testing.T) {
	const (
		uID = 1
		rID = 1

		accessToken    = "123"
		newAccessToken = "4567890"
		provider       = "github"
		remoteID       = 10
	)
	admin := true
	var (
		wantUser = &models.User{
			ID:      uID,
			IsAdmin: &admin, // first user is always admin
			RemoteIdentities: []*models.RemoteIdentity{{
				ID:          rID,
				Provider:    provider,
				RemoteID:    remoteID,
				AccessToken: accessToken,
				UserID:      uID,
			}},
		}
		updateAccessToken = &models.RemoteIdentity{
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: accessToken,
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&models.RemoteIdentity{
			Provider: provider,
			RemoteID: remoteID,
		},
	); err != nil {
		t.Fatal(err)
	}

	if err := db.UpdateAccessToken(updateAccessToken); err != nil {
		t.Error(err)
	}
	updatedUser, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(updatedUser, wantUser) {
		t.Errorf("have user %+v want %+v", updatedUser, wantUser)
	}

	// do another update
	updateAccessToken.AccessToken = newAccessToken
	wantUser.RemoteIdentities[0].AccessToken = newAccessToken
	if err := db.UpdateAccessToken(updateAccessToken); err != nil {
		t.Error(err)
	}
	updatedUser, err = db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(updatedUser, wantUser) {
		t.Errorf("have user %+v want %+v", updatedUser, wantUser)
	}
}
