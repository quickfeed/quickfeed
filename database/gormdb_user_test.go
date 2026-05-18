package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func TestDatabaseGetUserByCourse(t *testing.T) {
	for _, tc := range dbImplementations {
		t.Run(tc.name, func(t *testing.T) {
			databaseGetUserByCourse(t, tc.dbFunc)
		})
	}
}

func databaseGetUserByCourse(t *testing.T, dbFunc func(*testing.T) (database.Database, func())) {
	const username = "meling"
	db, cleanup := dbFunc(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{
		ID:              1,
		CourseCreatorID: admin.GetID(),
		Code:            "DAT320",
		Name:            "Operating Systems and Systems Programming",
		Year:            2021,
	}
	qtest.CreateCourse(t, db, admin, course)

	user := &qf.User{
		Login:     username,
		Name:      "Test User",
		Email:     "test@example.com",
		StudentID: "12345",
	}
	if err := db.CreateUser(user); err != nil {
		t.Error(err)
	}
	qtest.EnrollStudent(t, db, user, course)

	u, err := db.GetUserByCourse(course, username)
	if err != nil {
		t.Fatal(err)
	}
	if u.GetID() != user.GetID() {
		t.Errorf("expected user %d, got %d", user.GetID(), u.GetID())
	}
	if u.GetLogin() != username {
		t.Errorf("expected user %s, got %s", username, u.GetLogin())
	}
}
