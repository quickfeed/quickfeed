package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestDBUpdateUserAccessToken(t *testing.T) {
	const (
		accessToken    = "123"
		newAccessToken = "4567890"
		remoteID       = 10
	)
	wantUser := &qf.User{
		ID:      1,
		IsAdmin: true, // first user is always admin
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	var user qf.User
	if err := db.CreateUser(&user); err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(user.ID)
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}

	user.RefreshToken = accessToken
	user.ScmRemoteID = remoteID
	if err := db.UpdateUser(&user); err != nil {
		t.Error(err)
	}
	gotUser, err = db.GetUser(user.ID)
	if err != nil {
		t.Error(err)
	}
	wantUser.RefreshToken = accessToken
	wantUser.ScmRemoteID = remoteID
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}

	// do another update
	user.RefreshToken = newAccessToken
	if err := db.UpdateUser(&user); err != nil {
		t.Error(err)
	}
	gotUser, err = db.GetUser(user.ID)
	if err != nil {
		t.Error(err)
	}
	wantUser.RefreshToken = newAccessToken
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}
}

func TestDBUpdateAccessTokenUserGetAccessToken(t *testing.T) {
	const (
		newAccessToken = "123"
		anotherToken   = "456"
	)
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wantUser := qtest.CreateFakeUser(t, db)

	cachedAccessToken := wantUser.GetRefreshToken()

	// Update the access token for the user.
	wantUser.RefreshToken = newAccessToken
	if err := db.UpdateUser(wantUser); err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(wantUser.ID)
	if err != nil {
		t.Error(err)
	}

	// Assign the access token we expect to the wantUser object.
	wantUser.RefreshToken = newAccessToken
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
	cachedAccessToken2 := gotUser.GetRefreshToken()

	if cachedAccessToken == cachedAccessToken2 {
		t.Errorf("cached access token before and after are the same: %s == %s", cachedAccessToken, cachedAccessToken2)
	}

	// Update the access token again for the user.
	wantUser.RefreshToken = anotherToken
	if err := db.UpdateUser(wantUser); err != nil {
		t.Error(err)
	}
	gotUser, err = db.GetUser(wantUser.ID)
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
	cachedAccessToken3 := gotUser.GetRefreshToken()

	if cachedAccessToken2 == cachedAccessToken3 {
		t.Errorf("cached access token before and after are the same: %s == %s", cachedAccessToken2, cachedAccessToken3)
	}
}
