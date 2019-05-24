package database_test

import (
	"reflect"
	"testing"

	pb "github.com/autograde/aguis/ag"
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
		wantUser = &pb.User{
			Id:      uID,
			IsAdmin: admin, // first user is always admin
			RemoteIdentities: []*pb.RemoteIdentity{{
				Id:          rID,
				Provider:    provider,
				RemoteId:    remoteID,
				AccessToken: accessToken,
				UserId:      uID,
			}},
		}
		updateAccessToken = &pb.RemoteIdentity{
			Provider:    provider,
			RemoteId:    remoteID,
			AccessToken: accessToken,
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	var user pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&pb.RemoteIdentity{
			Provider: provider,
			RemoteId: remoteID,
		},
	); err != nil {
		t.Fatal(err)
	}

	if err := db.UpdateAccessToken(updateAccessToken); err != nil {
		t.Error(err)
	}
	updatedUser, err := db.GetUser(user.Id)
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
	updatedUser, err = db.GetUser(user.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(updatedUser, wantUser) {
		t.Errorf("have user %+v want %+v", updatedUser, wantUser)
	}
}
