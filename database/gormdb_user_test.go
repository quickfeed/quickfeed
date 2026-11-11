package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func TestGetUsersByStudentIDOrEmail(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	users := []*qf.User{
		{Login: "alice", Name: "Alice", StudentID: "1001", Email: "Alice@Example.com"},
		{Login: "bob", Name: "Bob", StudentID: "1002", Email: "bob@example.com"},
		{Login: "carol", Name: "Carol"}, // incomplete profile; no student ID or email
		{Login: "dave", Name: "Dave"},   // incomplete profile; no student ID or email
	}
	for _, user := range users {
		if err := db.CreateUser(user); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		studentID string
		email     string
		want      []string // logins of the expected users
	}{
		{name: "student ID only", studentID: "1001", want: []string{"alice"}},
		{name: "email only", email: "bob@example.com", want: []string{"bob"}},
		{name: "email ignores case", email: "ALICE@EXAMPLE.COM", want: []string{"alice"}},
		{name: "student ID ignores whitespace", studentID: "  1002  ", want: []string{"bob"}},
		{name: "email ignores whitespace", email: " bob@example.com ", want: []string{"bob"}},
		{name: "student ID and email of different users", studentID: "1001", email: "bob@example.com", want: []string{"alice", "bob"}},
		{name: "student ID and email of same user", studentID: "1002", email: "Bob@Example.com", want: []string{"bob"}},
		{name: "no match", studentID: "9999", email: "nobody@example.com", want: nil},
		{name: "empty student ID and email match no users", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsers, err := db.GetUsersByStudentIDOrEmail(tt.studentID, tt.email)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, user := range gotUsers {
				got = append(got, user.GetLogin())
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetUsersByStudentIDOrEmail(%q, %q) mismatch (-want +got):\n%s", tt.studentID, tt.email, diff)
			}
		})
	}
}

func TestGetUserByCourse(t *testing.T) {
	const username = "meling"
	db, cleanup := qtest.TestDB(t)
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
