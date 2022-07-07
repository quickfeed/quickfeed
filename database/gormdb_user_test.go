package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/types"
)

func TestGetUserByCourse(t *testing.T) {
	const username = "meling"
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateUser(t, db, 1, &types.User{Login: "admin"})
	course := &types.Course{
		ID:              1,
		CourseCreatorID: admin.ID,
		Code:            "DAT320",
		Name:            "Operating Systems and Systems Programming",
		Year:            2021,
	}
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateUser(t, db, 2, &types.User{Login: username})
	qtest.EnrollStudent(t, db, user, course)

	u, c, err := db.GetUserByCourse(course, username)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != user.ID {
		t.Errorf("expected user %d, got %d", user.ID, u.ID)
	}
	if c.ID != course.ID {
		t.Errorf("expected course %d, got %d", course.ID, c.ID)
	}
	if u.Login != username {
		t.Errorf("expected user %s, got %s", username, u.Login)
	}
}
