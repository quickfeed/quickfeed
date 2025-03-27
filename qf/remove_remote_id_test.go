package qf_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

func TestUserRemoveRemoteID(t *testing.T) {
	user := &qf.User{
		ID:           1,
		ScmRemoteID:  123,
		RefreshToken: "abc",
	}
	user.RemoveRemoteID()
	checkUser(t, user)
}

func TestGroupRemoveRemoteID(t *testing.T) {
	user1 := &qf.User{
		ID:           1,
		ScmRemoteID:  123,
		RefreshToken: "abc",
	}
	user2 := &qf.User{
		ID:           2,
		ScmRemoteID:  456,
		RefreshToken: "def",
	}
	group := &qf.Group{
		ID: 1,
		Users: []*qf.User{
			user1,
			user2,
		},
	}
	group.RemoveRemoteID()
	checkUser(t, user1)
	checkUser(t, user2)
}

func TestEnrollmentRemoveRemoteID(t *testing.T) {
	user := &qf.User{
		ID:           1,
		ScmRemoteID:  123,
		RefreshToken: "abc",
	}
	course := &qf.Course{
		ID: 1,
	}
	enrollment := &qf.Enrollment{
		ID:       1,
		CourseID: 1,
		UserID:   1,
		User:     user,
		Course:   course,
	}
	user.Enrollments = []*qf.Enrollment{enrollment}
	user.RemoveRemoteID()
	checkUser(t, user)
	enrollment.RemoveRemoteID()
	checkUser(t, enrollment.GetUser())
}

func checkUser(t *testing.T, user *qf.User) {
	t.Helper()
	if user.GetScmRemoteID() != 0 {
		t.Errorf("user.GetScmRemoteID() = %d, want 0", user.GetScmRemoteID())
	}
	if user.GetRefreshToken() != "" {
		t.Errorf(`user.GetRefreshToken() = %s, want ""`, user.GetRefreshToken())
	}
}
