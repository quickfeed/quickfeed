package qtest

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/qlog"
)

// TestDB returns a test database and close function.
// This function should only be used as a test helper.
func TestDB(t *testing.T) (database.Database, func()) {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "test.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	logger, err := qlog.Zap()
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.NewGormDB(f.Name(), logger)
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

// CreateFakeUser is a test helper to create a user in the database
// with the given remote id and the fake scm provider.
func CreateFakeUser(t *testing.T, db database.Database, remoteID uint64) *qf.User {
	t.Helper()
	var user qf.User
	err := db.CreateUserFromRemoteIdentity(&user,
		&qf.RemoteIdentity{
			Provider:    "fake",
			RemoteID:    remoteID,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return &user
}

func CreateUserFromRemoteIdentity(t *testing.T, db database.Database, remoteID *qf.RemoteIdentity) *qf.User {
	t.Helper()
	var user qf.User
	if err := db.CreateUserFromRemoteIdentity(&user, remoteID); err != nil {
		t.Fatal(err)
	}
	return &user
}

func CreateNamedUser(t *testing.T, db database.Database, remoteID uint64, name string) *qf.User {
	t.Helper()
	user := &qf.User{Name: name, Login: name}
	err := db.CreateUserFromRemoteIdentity(user,
		&qf.RemoteIdentity{
			Provider:    "fake",
			RemoteID:    remoteID,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func CreateUser(t *testing.T, db database.Database, remoteID uint64, user *qf.User) *qf.User {
	t.Helper()
	err := db.CreateUserFromRemoteIdentity(user,
		&qf.RemoteIdentity{
			Provider:    "fake",
			RemoteID:    remoteID,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func CreateAdminUser(t *testing.T, db database.Database, provider string) *qf.User {
	t.Helper()
	user := &qf.User{Name: "admin", Login: "admin"}
	err := db.CreateUserFromRemoteIdentity(user,
		&qf.RemoteIdentity{
			Provider:    provider,
			RemoteID:    1,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func CreateCourse(t *testing.T, db database.Database, user *qf.User, course *qf.Course) {
	t.Helper()
	if course.Provider == "" {
		for _, rid := range user.RemoteIdentities {
			if rid.Provider != "" {
				course.Provider = rid.Provider
			}
		}
	}
	if err := db.CreateCourse(user.ID, course); err != nil {
		t.Fatal(err)
	}
}

func EnrollStudent(t *testing.T, db database.Database, student *qf.User, course *qf.Course) {
	t.Helper()
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: student.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   student.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
}

func EnrollTeacher(t *testing.T, db database.Database, student *qf.User, course *qf.Course) {
	t.Helper()
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: student.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   student.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}
}

func RandomString(t *testing.T) string {
	t.Helper()
	randomness := make([]byte, 10)
	if _, err := rand.Read(randomness); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(randomness))[:6]
}

// AssignmentsWithTasks returns a list of test assignments with tasks for the given course.
func AssignmentsWithTasks(t *testing.T, courseID uint64) []*qf.Assignment {
	t.Helper()
	return []*qf.Assignment{
		{
			CourseID:         courseID,
			Name:             "lab1",
			RunScriptContent: "Script for lab1",
			Deadline:         Timestamp(t, "12.01.2022"),
			AutoApprove:      false,
			Order:            1,
			IsGroupLab:       false,
			Tasks: []*qf.Task{
				{Title: "Fibonacci", Name: "fib", AssignmentOrder: 1, Body: "Implement fibonacci"},
				{Title: "Lucas Numbers", Name: "luc", AssignmentOrder: 1, Body: "Implement lucas numbers"},
			},
		},
		{
			CourseID:         courseID,
			Name:             "lab2",
			RunScriptContent: "Script for lab2",
			Deadline:         Timestamp(t, "12.12.2021"),
			AutoApprove:      false,
			Order:            2,
			IsGroupLab:       false,
			Tasks: []*qf.Task{
				{Title: "Addition", Name: "add", AssignmentOrder: 2, Body: "Implement addition"},
				{Title: "Subtraction", Name: "sub", AssignmentOrder: 2, Body: "Implement subtraction"},
				{Title: "Multiplication", Name: "mul", AssignmentOrder: 2, Body: "Implement multiplication"},
			},
		},
	}
}

func QuickFeedClient(url string) qfconnect.QuickFeedServiceClient {
	serverUrl := url
	if serverUrl == "" {
		serverUrl = "http://127.0.0.1:8081"
	}
	return qfconnect.NewQuickFeedServiceClient(http.DefaultClient, serverUrl)
}
