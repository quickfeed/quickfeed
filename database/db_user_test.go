package database_test

import (
	"reflect"
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
	updatedUser, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	updatedUser.Enrollments = nil
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
	updatedUser.Enrollments = nil
	if !reflect.DeepEqual(updatedUser, wantUser) {
		t.Errorf("have user %+v want %+v", updatedUser, wantUser)
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
	if err := db.UpdateAccessToken(&pb.RemoteIdentity{
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
	if err := db.UpdateAccessToken(&pb.RemoteIdentity{
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

func TestGormDBUpdateAccessTokenCourseTokenCache(t *testing.T) {
	const (
		stud           = "student1"
		newAccessToken = "123"
		anotherToken   = "456"
		provider       = "fake"
		remoteID       = 10
		remoteID2      = 11
	)
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateNamedUser(t, db, remoteID, "admin")
	initialAdminToken, err := admin.GetAccessToken(provider)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		ID:              1,
		CourseCreatorID: admin.ID,
		Code:            "DAT320",
		Name:            "Operating Systems and Systems Programming",
		Provider:        provider,
		Year:            2021,
	}
	qtest.CreateCourse(t, db, admin, course)

	cr, err := db.GetCourse(1, false)
	if err != nil {
		t.Fatal(err)
	}
	cachedToken := cr.GetAccessToken()
	if cachedToken != initialAdminToken {
		t.Errorf("cached token different from expected initial token: %s != %s", cachedToken, initialAdminToken)
	}

	// Update the access token for the user.
	if err := db.UpdateAccessToken(&pb.RemoteIdentity{
		Provider:    provider,
		RemoteID:    remoteID,
		AccessToken: newAccessToken,
	}); err != nil {
		t.Error(err)
	}

	cr, err = db.GetCourse(1, false)
	if err != nil {
		t.Fatal(err)
	}
	cachedToken = cr.GetAccessToken()
	if cachedToken != newAccessToken {
		t.Errorf("cached token different from expected updated token: %s != %s", cachedToken, newAccessToken)
	}

	// Update the access token for the user again.
	if err := db.UpdateAccessToken(&pb.RemoteIdentity{
		Provider:    provider,
		RemoteID:    remoteID,
		AccessToken: anotherToken,
	}); err != nil {
		t.Error(err)
	}

	cr, err = db.GetCourse(1, false)
	if err != nil {
		t.Fatal(err)
	}
	cachedToken = cr.GetAccessToken()
	if cachedToken != anotherToken {
		t.Errorf("cached token different from expected updated token: %s != %s", cachedToken, anotherToken)
	}
}
