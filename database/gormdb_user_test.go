package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func TestGetUserByCourse(t *testing.T) {
	const username = "meling"
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{
		ID:              1,
		CourseCreatorID: admin.ID,
		Code:            "DAT320",
		Name:            "Operating Systems and Systems Programming",
		Year:            2021,
	}
	qtest.CreateCourse(t, db, admin, course)

	user := &qf.User{Login: username}
	if err := db.CreateUser(user); err != nil {
		t.Error(err)
	}
	qtest.EnrollStudent(t, db, user, course)

	u, err := db.GetUserByCourse(course, username)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != user.ID {
		t.Errorf("expected user %d, got %d", user.ID, u.ID)
	}
	if u.Login != username {
		t.Errorf("expected user %s, got %s", username, u.Login)
	}
}
