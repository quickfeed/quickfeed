package qtest

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	db, err := database.NewGormDB(f.Name(), log.Zap(true))
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

// CreateFakeUser is a test helper to create a user in the database
// with the given remote id and the fake scm provider.
func CreateFakeUser(t *testing.T, db database.Database, remoteID uint64) *pb.User {
	t.Helper()
	var user pb.User
	err := db.CreateUserFromRemoteIdentity(&user,
		&pb.RemoteIdentity{
			Provider:    "fake",
			RemoteID:    remoteID,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return &user
}

func CreateNamedUser(t *testing.T, db database.Database, remoteID uint64, name string) *pb.User {
	t.Helper()
	user := &pb.User{Name: name}
	err := db.CreateUserFromRemoteIdentity(user,
		&pb.RemoteIdentity{
			Provider:    "fake",
			RemoteID:    remoteID,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func EnrollStudent(t *testing.T, db database.Database, student *pb.User, course *pb.Course) {
	t.Helper()
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
}

// FakeProviderMap is a test helper function to create an SCM map.
func FakeProviderMap(t *testing.T) (scm.SCM, *auth.Scms) {
	t.Helper()
	scms := auth.NewScms()
	scm, err := scms.GetOrCreateSCMEntry(zap.NewNop(), "fake", "token")
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
	return fmt.Sprintf("%x", sha1.Sum(randomness))[:6]
}

// Timestamp returns a protobuf timestamp representation of the given string time.
func Timestamp(t *testing.T, tim string) *timestamppb.Timestamp {
	t.Helper()
	timeTime, err := time.Parse(pb.TimeLayout, tim)
	if err != nil {
		t.Fatal(err)
	}
	return timestamppb.New(timeTime)
}
