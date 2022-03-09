package database_test

import (
	"errors"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)
func TestGormDBCourseUniqueContraint(t *testing.T) {
	// Test that a course with the same organization ID or code and year cannot be created.
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wantCourse := &pb.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		Provider:       "github",
		OrganizationID: 1235,
	}
	course := &pb.Course{
		Name:           "Test Course 2",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		Provider:       "github",
		OrganizationID: 1234,
	}

	remoteID := &pb.RemoteIdentity{Provider: wantCourse.Provider, RemoteID: 10, AccessToken: "token"}
	admin := qtest.CreateUserFromRemoteIdentity(t, db, remoteID)

	if err := db.CreateCourse(admin.ID, wantCourse); err != nil {
		t.Fatal(err)
	}

	// CreateCourse should fail because the unique constraint (course.code, course.year) is violated
	if err := db.CreateCourse(admin.ID, course); err != nil && !errors.Is(err, database.ErrCourseExists) {
		t.Fatal(err)
	}

	// CreateCourse should fail because OrganizationID is not unique
	if err := db.CreateCourse(admin.ID, &pb.Course{OrganizationID: wantCourse.OrganizationID}); err != nil && !errors.Is(err, database.ErrCourseExists) {
		t.Fatal(err)
	}

	gotCourse, err := db.GetCourse(wantCourse.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("course mismatch (-want +got):\n%s", diff)
	}

	// Now create a course with same code but different year
	course.Year = 2018
	// CreateCourse should succeed because the unique constraint (course.code, course.year) is not violated
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}
}
