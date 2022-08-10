package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
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
		wantUser = &qf.User{
			ID:      uID,
			IsAdmin: admin, // first user is always admin
			RemoteIdentities: []*qf.RemoteIdentity{{
				ID:          rID,
				Provider:    provider,
				RemoteID:    remoteID,
				AccessToken: accessToken,
				UserID:      uID,
			}},
		}
		updateAccessToken = &qf.RemoteIdentity{
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: accessToken,
		}
	)

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	var user qf.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&qf.RemoteIdentity{
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
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
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
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}
}

func TestGormDBUpdateAccessTokenUserGetAccessToken(t *testing.T) {
	const (
		newAccessToken = "123"
		anotherToken   = "456"
		provider       = "fake"
		remoteID       = 10
	)
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wantUser := qtest.CreateFakeUser(t, db, remoteID)

	cachedAccessToken, err := wantUser.GetAccessToken(provider)
	if err != nil {
		t.Fatal(err)
	}

	// Update the access token for the user.
	if err := db.UpdateAccessToken(&qf.RemoteIdentity{
		Provider:    provider,
		RemoteID:    remoteID,
		AccessToken: newAccessToken,
	}); err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(wantUser.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Assign the access token we expect to the wantUser object.
	wantUser.RemoteIdentities[0].AccessToken = newAccessToken
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
	cachedAccessToken2, err := gotUser.GetAccessToken(provider)
	if err != nil {
		t.Fatal(err)
	}

	if cachedAccessToken == cachedAccessToken2 {
		t.Errorf("cached access token before and after are the same: %s == %s", cachedAccessToken, cachedAccessToken2)
	}

	// Update the access token again for the user.
	if err := db.UpdateAccessToken(&qf.RemoteIdentity{
		Provider:    provider,
		RemoteID:    remoteID,
		AccessToken: anotherToken,
	}); err != nil {
		t.Error(err)
	}
	gotUser, err = db.GetUser(wantUser.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Assign the access token we expect to the wantUser object.
	wantUser.RemoteIdentities[0].AccessToken = anotherToken
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
	cachedAccessToken3, err := gotUser.GetAccessToken(provider)
	if err != nil {
		t.Fatal(err)
	}

	if cachedAccessToken2 == cachedAccessToken3 {
		t.Errorf("cached access token before and after are the same: %s == %s", cachedAccessToken2, cachedAccessToken3)
	}
}
