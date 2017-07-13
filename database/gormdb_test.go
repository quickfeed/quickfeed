package database_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func setup(t *testing.T) (*database.GormDB, func()) {
	const (
		driver = "sqlite3"
		prefix = "testdb"
	)

	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	db, err := database.NewGormDB(driver, f.Name(), envSet("LOGDB"))
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

func TestGormDBGetUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetUser(10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetUsers(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetUsers(); err != nil {
		t.Errorf("have error '%v' wanted '%v'", err, nil)
	}
}

func TestGormDBGetCourses(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := models.Course{
		Name:        "Test",
		Code:        "T100",
		Year:        2017,
		Tag:         "something",
		Provider:    "github",
		DirectoryID: 1,
	}

	if err := db.CreateCourse(&course); err != nil {
		t.Fatal(err)
	}

	courses, err := db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}

	if len(courses) != 1 {
		t.Errorf("have size %v wanted %v", len(courses), 1)
	}

	if !reflect.DeepEqual(courses[0], &course) {
		t.Fatalf("want %v have %v", courses[0], &course)
	}
}

func TestGormDBGetAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetAssignmentsByCourse(10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBCreateAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if err := db.CreateAssignment(&models.Assignment{CourseID: 1, Name: "Lab 1"}); err != nil {
		t.Errorf("have error '%v' wanted '%v'", err, nil)
	}
}

func TestGormDBGetAssignmentExists(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()
	errCreate := db.CreateCourse(&models.Course{Code: "", DirectoryID: 1, Name: "Test", Provider: "Test", Tag: "", Year: 2017})
	if errCreate != nil {
		t.Fatal(errCreate)
	}
	errCreate = db.CreateAssignment(&models.Assignment{CourseID: 1, Name: "Lab 1"})
	if errCreate != nil {
		t.Fatal(errCreate)
	}

	if _, err := db.GetAssignmentsByCourse(1); err != nil {
		t.Errorf("have error '%v' wanted '%v'", err, nil)
	}
}

func TestGormDBCreateEnrollmentNoRecord(t *testing.T) {
	const (
		userID   = 1
		courseID = 1
	)

	db, cleanup := setup(t)
	defer cleanup()

	if err := db.CreateEnrollment(&models.Enrollment{
		UserID:   userID,
		CourseID: courseID,
	}); err != gorm.ErrRecordNotFound {
		t.Errorf("expected error '%v' have '%v'", gorm.ErrRecordNotFound, err)
	}
}

func TestGormDBCreateEnrollment(t *testing.T) {
	const (
		secret   = "123"
		provider = "github"
		remoteID = 10
	)

	db, cleanup := setup(t)
	defer cleanup()

	var course models.Course
	if err := db.CreateCourse(&course); err != nil {
		t.Fatal(err)
	}

	user, err := db.CreateUserFromRemoteIdentity(provider, remoteID, secret)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.CreateEnrollment(&models.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Error(err)
	}
}

func TestGormDBAcceptRejectEnrollment(t *testing.T) {
	const (
		secret   = "123"
		provider = "github"
		remoteID = 10
	)

	db, cleanup := setup(t)
	defer cleanup()

	var course models.Course
	if err := db.CreateCourse(&course); err != nil {
		t.Fatal(err)
	}

	user, err := db.CreateUserFromRemoteIdentity(provider, remoteID, secret)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.CreateEnrollment(&models.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// Get user's pending enrollments.
	userEnrollments, err := db.GetEnrollmentsByUser(user.ID, models.Pending)
	if err != nil {
		t.Fatal(err)
	}

	if len(userEnrollments) != 1 {
		t.Fatal("there should be 1 pending enrollment")
	}

	// Get course's pending enrollments.
	courseEnrollments, err := db.GetEnrollmentsByCourse(course.ID, models.Pending)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that GetEnrollmentsForCourse returns the same enrollments.
	if !reflect.DeepEqual(userEnrollments, courseEnrollments) {
		t.Fatalf("want %v have %v", userEnrollments, courseEnrollments)
	}

	enrollmentID := userEnrollments[0].ID
	// Accept enrollment.
	if err := db.AcceptEnrollment(enrollmentID); err != nil {
		t.Fatal(err)
	}

	// Get user's accepted enrollments.
	userEnrollments, err = db.GetEnrollmentsByUser(user.ID, models.Accepted)
	if err != nil {
		t.Fatal(err)
	}

	if len(userEnrollments) != 1 {
		t.Fatal("there should be 1 accepted enrollment")
	}

	// Get course's accepted enrollments.
	courseEnrollments, err = db.GetEnrollmentsByCourse(course.ID, models.Accepted)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that GetEnrollmentsForCourse returns the same enrollments.
	if !reflect.DeepEqual(userEnrollments, courseEnrollments) {
		t.Fatalf("want %v have %v", userEnrollments, courseEnrollments)
	}

	// Reject enrollment.
	if err := db.RejectEnrollment(enrollmentID); err != nil {
		t.Fatal(err)
	}

	// Get user's rejected enrollments.
	userEnrollments, err = db.GetEnrollmentsByUser(user.ID, models.Rejected)
	if err != nil {
		t.Fatal(err)
	}

	if len(userEnrollments) != 1 {
		t.Fatal("there should be 1 rejected enrollment")
	}

	// Get course's rejected enrollments.
	courseEnrollments, err = db.GetEnrollmentsByCourse(course.ID, models.Rejected)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that GetEnrollmentsForCourse returns the same enrollments.
	if !reflect.DeepEqual(userEnrollments, courseEnrollments) {
		t.Fatalf("want %v have %v", userEnrollments, courseEnrollments)
	}
}

func TestGormDBDuplicateIdentity(t *testing.T) {
	const (
		uID  = 1
		rID1 = 1

		secret1   = "123"
		provider1 = "github"
		remoteID1 = 10
	)

	var (
		wantUser1 = &models.User{
			ID: uID,
			RemoteIdentities: []*models.RemoteIdentity{{
				ID:          rID1,
				Provider:    provider1,
				RemoteID:    remoteID1,
				AccessToken: secret1,
				UserID:      uID,
			}},
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	user1, err := db.CreateUserFromRemoteIdentity(provider1, remoteID1, secret1)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(user1, wantUser1) {
		t.Errorf("have user %+v want %+v", user1, wantUser1)
	}

	if _, err := db.CreateUserFromRemoteIdentity(provider1, remoteID1, secret1); err == nil {
		t.Errorf("expected error '%v'", database.ErrDuplicateIdentity)
	}
}

func TestGormDBAssociateUserWithRemoteIdentity(t *testing.T) {
	const (
		uID  = 1
		rID1 = 1
		rID2 = 2

		secret1   = "123"
		provider1 = "github"
		remoteID1 = 10

		secret2   = "ABC"
		provider2 = "gitlab"
		remoteID2 = 20

		secret3 = "DEF"
	)

	var (
		wantUser1 = &models.User{
			ID: uID,
			RemoteIdentities: []*models.RemoteIdentity{{
				ID:          rID1,
				Provider:    provider1,
				RemoteID:    remoteID1,
				AccessToken: secret1,
				UserID:      uID,
			}},
		}

		wantUser2 = &models.User{
			ID: uID,
			RemoteIdentities: []*models.RemoteIdentity{
				{
					ID:          rID1,
					Provider:    provider1,
					RemoteID:    remoteID1,
					AccessToken: secret1,
					UserID:      uID,
				},
				{
					ID:          rID2,
					Provider:    provider2,
					RemoteID:    remoteID2,
					AccessToken: secret2,
					UserID:      uID,
				},
			},
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	user1, err := db.CreateUserFromRemoteIdentity(provider1, remoteID1, secret1)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(user1, wantUser1) {
		t.Errorf("have user %+v want %+v", user1, wantUser1)
	}

	if err := db.AssociateUserWithRemoteIdentity(user1.ID, provider2, remoteID2, secret2); err != nil {
		t.Fatal(err)
	}

	user2, err := db.GetUser(uID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(user2, wantUser2) {
		t.Errorf("have user %+v want %+v", user2, wantUser2)
	}

	if err := db.AssociateUserWithRemoteIdentity(user1.ID, provider2, remoteID2, secret3); err != nil {
		t.Fatal(err)
	}

	user3, err := db.GetUser(uID)
	if err != nil {
		t.Fatal(err)
	}

	wantUser2.RemoteIdentities[1].AccessToken = secret3
	if !reflect.DeepEqual(user3, wantUser2) {
		t.Errorf("have user %+v want %+v", user3, wantUser2)
	}
}

func TestGormDBSetAdminNoRecord(t *testing.T) {
	const id = 1

	db, cleanup := setup(t)
	defer cleanup()

	if err := db.SetAdmin(id); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBSetAdmin(t *testing.T) {
	const (
		uID = 1
		rID = 1

		secret   = "123"
		provider = "github"
		remoteID = 10
	)

	var (
		wantUser = &models.User{
			ID: uID,
			RemoteIdentities: []*models.RemoteIdentity{{
				ID:          rID,
				Provider:    provider,
				RemoteID:    remoteID,
				AccessToken: secret,
				UserID:      uID,
			}},
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	user, err := db.CreateUserFromRemoteIdentity(provider, remoteID, secret)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(user, wantUser) {
		t.Errorf("have user %+v want %+v", user, wantUser)
	}

	if err := db.SetAdmin(user.ID); err != nil {
		t.Error(err)
	}

	admin, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantUser.IsAdmin = true
	if !reflect.DeepEqual(admin, wantUser) {
		t.Errorf("have user %+v want %+v", admin, wantUser)
	}
}

func TestGormDBCreateCourse(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := models.Course{
		Name: "name",
		Code: "code",
		Year: 2017,
		Tag:  "tag",

		Provider:    "github",
		DirectoryID: 1,
	}

	if err := db.CreateCourse(&course); err != nil {
		t.Fatal(err)
	}

	if course.ID == 0 {
		t.Error("expected id to be set")
	}
}

func TestGormDBGetCourse(t *testing.T) {
	course := &models.Course{
		Name:        "Test Course",
		Code:        "DAT100",
		Year:        2017,
		Tag:         "Spring",
		Provider:    "github",
		DirectoryID: 1234,
	}

	db, cleanup := setup(t)
	defer cleanup()

	err := db.CreateCourse(course)
	if err != nil {
		t.Fatal(err)
	}

	// Get the created course.
	createdCourse, err := db.GetCourse(course.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(createdCourse, course) {
		t.Errorf("have course %+v want %+v", createdCourse, course)
	}

}

func TestGormDBGetCourseNonExist(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetCourse(20); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}

}
func TestGormDBUpdateCourse(t *testing.T) {
	var (
		course = &models.Course{
			Name:        "Test Course",
			Code:        "DAT100",
			Year:        2017,
			Tag:         "Spring",
			Provider:    "github",
			DirectoryID: 1234,
		}
		updates = &models.Course{
			Name:        "Test Course Edit",
			Code:        "DAT100-1",
			Year:        2018,
			Tag:         "Autumn",
			Provider:    "gitlab",
			DirectoryID: 12345,
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	err := db.CreateCourse(course)
	if err != nil {
		t.Fatal(err)
	}

	updates.ID = course.ID
	err = db.UpdateCourse(updates)
	if err != nil {
		t.Fatal(err)
	}

	// Get the updated course.
	updatedCourse, err := db.GetCourse(course.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(updatedCourse, updates) {
		t.Errorf("have course %+v want %+v", updatedCourse, course)
	}
}

func envSet(env string) database.GormLogger {
	if os.Getenv(env) != "" {
		return database.Logger{Logger: logrus.New()}
	}
	return nil
}
