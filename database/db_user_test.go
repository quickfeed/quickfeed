package database_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
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
			ID:      uID,
			IsAdmin: admin, // first user is always admin
			RemoteIdentities: []*pb.RemoteIdentity{{
				ID:          rID,
				Provider:    provider,
				RemoteID:    remoteID,
				AccessToken: accessToken,
				UserID:      uID,
			}},
		}
		updateAccessToken = &pb.RemoteIdentity{
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: accessToken,
		}
	)

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	var user pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&pb.RemoteIdentity{
			Provider: provider,
			RemoteID: remoteID,
		},
	); err != nil {
		t.Fatal(err)
	}

	if err := db.UpdateAccessToken(updateAccessToken); err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	gotUser.Enrollments = nil
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):n%s", diff)
	}

	// do another update
	updateAccessToken.AccessToken = newAccessToken
	wantUser.RemoteIdentities[0].AccessToken = newAccessToken
	if err := db.UpdateAccessToken(updateAccessToken); err != nil {
		t.Error(err)
	}
	gotUser, err = db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	gotUser.Enrollments = nil
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):n%s", diff)
	}
}
