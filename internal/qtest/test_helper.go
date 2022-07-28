package qtest

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"google.golang.org/grpc/metadata"
)

// TestDB returns a test database and close function.
// This function should only be used as a test helper.
func TestDB(t *testing.T) (database.Database, func()) {
	t.Helper()

	f, err := ioutil.TempFile(t.TempDir(), "test.db")
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
	user := &qf.User{Name: name}
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
	user := &qf.User{}
	err := db.CreateUserFromRemoteIdentity(user,
		&qf.RemoteIdentity{
			Provider:    provider,
			RemoteID:    1,
			AccessToken: scm.GetAccessToken(t),
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

// FakeProviderMap is a test helper function to create an SCM map.
func FakeProviderMap(t *testing.T) (scm.SCM, *scm.Scms) {
	t.Helper()
	scms := scm.NewScms()
	env.SetFakeProvider(t)
	scm, err := scms.GetOrCreateSCMEntry(Logger(t).Desugar(), "token")
	if err != nil {
		t.Fatal(err)
	}
	return scm, scms
}

func RandomString(t *testing.T) string {
	t.Helper()
	randomness := make([]byte, 10)
	if _, err := rand.Read(randomness); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(randomness))[:6]
}

// WithUserContext is a test helper function to create metadata for the
// given user mimicking the context coming from the browser.
func WithUserContext(ctx context.Context, user *qf.User) context.Context {
	userID := strconv.Itoa(int(user.GetID()))
	meta := metadata.New(map[string]string{"user": userID})
	return metadata.NewIncomingContext(ctx, meta)
}

// AssignmentsWithTasks returns a list of test assignments with tasks for the given course.
func AssignmentsWithTasks(courseID uint64) []*qf.Assignment {
	return []*qf.Assignment{
		{
			CourseID:         courseID,
			Name:             "lab1",
			RunScriptContent: "Script for lab1",
			Deadline:         "12.01.2022",
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
			Deadline:         "12.12.2021",
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

// PopulateDatabaseWithInitialData creates initial data-records based on organization
// This function was created with the intent of being used for testing task and pull request related functionality.
func PopulateDatabaseWithInitialData(t *testing.T, db database.Database, sc scm.SCM, course *qf.Course) error {
	t.Helper()

	ctx := context.Background()
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.OrganizationPath})
	if err != nil {
		return err
	}
	course.OrganizationID = org.GetID()
	admin := CreateAdminUser(t, db, course.GetProvider())
	db.UpdateUser(admin)
	CreateCourse(t, db, admin, course)

	repos, err := sc.GetRepositories(ctx, org)
	if err != nil {
		return err
	}

	// Create repositories
	nxtRemoteID := uint64(2)
	for _, repo := range repos {
		dbRepo := &qf.Repository{
			RepositoryID:   repo.ID,
			OrganizationID: org.GetID(),
			HTMLURL:        repo.HTMLURL,
			RepoType:       qf.RepoType(repo.Path),
		}
		if dbRepo.IsUserRepo() {
			user := &qf.User{}
			CreateUser(t, db, nxtRemoteID, user)
			nxtRemoteID++
			EnrollStudent(t, db, user, course)
			group := &qf.Group{
				Name:     dbRepo.UserName(),
				CourseID: course.GetID(),
				Users:    []*qf.User{user},
			}
			if err := db.CreateGroup(group); err != nil {
				return err
			}
			// For testing purposes, assume all student repositories are group repositories
			// since tasks and pull requests are only supported for groups anyway.
			dbRepo.RepoType = qf.Repository_GROUP
			dbRepo.GroupID = group.GetID()
		}

		t.Logf("create repo: %v", dbRepo)
		if err = db.CreateRepository(dbRepo); err != nil {
			return err
		}
	}
	return nil
}
