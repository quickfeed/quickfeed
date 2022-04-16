package auth_test

import (
	"testing"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/web/auth"
	"github.com/google/go-cmp/cmp"
)

func TestTokenManager(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	course := &pb.Course{}
	qtest.CreateCourse(t, db, admin, course)
	user := qtest.CreateFakeUser(t, db, 2)

	manager, err := auth.NewTokenManager(db, time.Minute*30, "notasecret", "localhost")
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Add(2); err != nil {
		t.Error(err)
	}

	claimsNoUpdate := auth.Claims{
		UserID:  10,
		Admin:   false,
		Courses: make(map[uint64]pb.Enrollment_UserStatus, 0),
	}
	if manager.UpdateRequired(&claimsNoUpdate) {
		t.Error("JWT update required is true, expected false")
	}

	claimsToUpdate := auth.Claims{
		UserID:  2,
		Admin:   false,
		Courses: map[uint64]pb.Enrollment_UserStatus{1: pb.Enrollment_STUDENT},
	}
	if !manager.UpdateRequired(&claimsToUpdate) {
		t.Error("JWT update required is false, expected true")
	}

	// User is not enrolled in any course, this must be reflected in the updated claims
	updatedClaims, err := manager.NewClaims(user.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if len(updatedClaims.Courses) > 0 {
		t.Errorf("got %d enrollments, expected 0", len(updatedClaims.Courses))
	}

	if err := manager.Add(admin.ID); err != nil {
		t.Fatal(err)
	}
	wantTokenList := []uint64{2, 1}
	haveTokenList := manager.GetTokens()
	if !cmp.Equal(wantTokenList, haveTokenList) {
		t.Errorf("token list is %v, expected %v", haveTokenList, wantTokenList)
	}
	if err := manager.Remove(user.GetID()); err != nil {
		t.Error(err)
	}
	if err := manager.Update(); err != nil {
		t.Fatal(err)
	}
	// Only the admin (user with ID = 1) must be in the refreshed list
	wantTokenList = []uint64{1}
	haveTokenList = manager.GetTokens()
	if !cmp.Equal(wantTokenList, haveTokenList) {
		t.Errorf("token list is %v, expected %v", haveTokenList, wantTokenList)
	}
	if err := manager.Remove(admin.ID); err != nil {
		t.Error(err)
	}
	haveTokenList = manager.GetTokens()
	if len(haveTokenList) > 0 {
		t.Errorf("%d tokens in the list, expected 0", len(haveTokenList))
	}
}
