package auth_test

import (
	"testing"

	"github.com/autograde/quickfeed/ag"
	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/web/auth"
	"github.com/google/go-cmp/cmp"
)

func TestTokenManager(t *testing.T) {
	manager := &auth.TokenManager{
		TokensToUpdate: []uint64{2, 3, 4},
	}

	claimsToUpdate := auth.Claims{
		UserID:  2,
		Admin:   false,
		Courses: map[uint64]ag.Enrollment_UserStatus{1: pb.Enrollment_STUDENT},
	}

	claimsNoUpdate := auth.Claims{
		UserID:  10,
		Admin:   false,
		Courses: make(map[uint64]pb.Enrollment_UserStatus, 0),
	}

	if manager.UpdateRequired(&claimsNoUpdate) {
		t.Error("expected false (update not required), got true")
	}

	if !manager.UpdateRequired(&claimsToUpdate) {
		t.Error("expected true (update required), got false")
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	manager.DB = db

	admin := qtest.CreateFakeUser(t, db, 10)
	course := &pb.Course{}
	qtest.CreateCourse(t, db, admin, course)
	user := qtest.CreateFakeUser(t, db, 2)

	// User is not enrolled in any course, this must be reflected in the updated claims
	updatedClaims, err := manager.UpdateClaims(user.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if len(updatedClaims.Courses) > 0 {
		t.Errorf("expected 0 enrollments, got %d", len(updatedClaims.Courses))
	}

	if err := manager.Add(admin.ID); err != nil {
		t.Fatal(err)
	}
	wantTokenList := []uint64{2, 3, 4, 1}
	if !cmp.Equal(wantTokenList, manager.TokensToUpdate) {
		t.Errorf("mismatch: expected %v got %v", wantTokenList, manager.TokensToUpdate)
	}
	if err := manager.Update(); err != nil {
		t.Fatal(err)
	}
	// Only the admin (user with ID = 1) must be in the refreshed list
	wantTokenList = []uint64{1}
	if !cmp.Equal(wantTokenList, manager.TokensToUpdate) {
		t.Errorf("mismatch: expected %v got %v", wantTokenList, manager.TokensToUpdate)
	}
	if err := manager.Remove(admin.ID); err != nil {
		t.Fatal(err)
	}
	if len(manager.TokensToUpdate) > 0 {
		t.Errorf("expected 0 elements in the list, got %d", len(manager.TokensToUpdate))
	}
}
