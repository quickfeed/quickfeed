package auth_test

import (
	"testing"

	"github.com/autograde/quickfeed/ag"
	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/web/auth"
)

func TestUpdateJWTClaims(t *testing.T) {
	tokens := []*pb.UpdateTokenRecord{
		{
			ID:     1,
			UserID: 1,
		},
		{
			ID:     2,
			UserID: 2,
		},
		{
			ID:     3,
			UserID: 3,
		},
	}
	manager := &auth.TokenManager{TokensToUpdate: tokens}

	claimsToUpdate := auth.Claims{
		UserID:  2,
		Admin:   false,
		Courses: map[uint64]ag.Enrollment_UserStatus{1: pb.Enrollment_STUDENT},
	}

	claimsNoUpdate := auth.Claims{
		UserID:  10,
		Admin:   true,
		Courses: make(map[uint64]pb.Enrollment_UserStatus, 0),
	}

	idNoUpdate := manager.JWTUpdateRequired(&claimsNoUpdate)
	if idNoUpdate != -1 {
		t.Errorf("expected index -1 (update not required), got %d", idNoUpdate)
	}
	idToUpdate := manager.JWTUpdateRequired(&claimsToUpdate)
	if idToUpdate != int(claimsToUpdate.UserID) {
		t.Errorf("expected user ID %d, got %d", claimsToUpdate.UserID, idToUpdate)
	}
}

func TestJWTDatabaseUpdates(t *testing.T) {
	// db, cleanup := qtest.TestDB(t)
	// defer cleanup()

	// TODO(vera): create user, create user claims with different info, update,
	// make sure new claims are correct
}

func TestJWTmanagerUpdates(t *testing.T) {
	// db, cleanup := qtest.TestDB(t)
	// defer cleanup()

	// TODO(vera): make manager, add records to db

	// add new record, check that db and manager have same records

	// remove a record, check manager and db
}
