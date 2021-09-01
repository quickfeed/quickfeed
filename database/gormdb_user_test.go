package database_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
)

func TestGetUserByCourse(t *testing.T) {
	const username = "meling"
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateUser(t, db, 1, &pb.User{Login: "admin"})
	course := &pb.Course{
		ID:              1,
		CourseCreatorID: admin.ID,
		Code:            "DAT320",
		Name:            "Operating Systems and Systems Programming",
		Year:            2021,
	}
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}

	user := qtest.CreateUser(t, db, 2, &pb.User{Login: username})
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
