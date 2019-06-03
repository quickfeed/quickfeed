package database_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
)

func setup(t *testing.T) (database.Database, func()) {
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

func TestGormDBUpdateUser(t *testing.T) {
	const (
		uID = 1
		rID = 1

		secret   = "123"
		provider = "github"
		remoteID = 10
	)
	admin := true
	var (
		wantUser = &pb.User{
			ID:        uID,
			IsAdmin:   admin, // first user is always admin
			Name:      "Scrooge McDuck",
			StudentID: "22",
			Email:     "scrooge@mc.duck",
			AvatarURL: "https://github.com",
			RemoteIdentities: []*pb.RemoteIdentity{{
				ID:          rID,
				Provider:    provider,
				RemoteID:    remoteID,
				AccessToken: secret,
				UserID:      uID,
			}},
		}
		updates = &pb.User{
			ID:        uID,
			Name:      "Scrooge McDuck",
			StudentID: "22",
			Email:     "scrooge@mc.duck",
			AvatarURL: "https://github.com",
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	var user pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&pb.RemoteIdentity{
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: secret,
		},
	); err != nil {
		t.Fatal(err)
	}

	if err := db.UpdateUser(updates); err != nil {
		t.Error(err)
	}

	updatedUser, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(updatedUser, wantUser) {
		t.Errorf("have user %+v want %+v", updatedUser, wantUser)
	}
}

func TestGormDBGetCourses(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	user := createFakeUser(t, db, 10)
	c1 := pb.Course{DirectoryID: 1}
	if err := db.CreateCourse(user.ID, &c1); err != nil {
		t.Fatal(err)
	}

	c2 := pb.Course{DirectoryID: 2}
	if err := db.CreateCourse(user.ID, &c2); err != nil {
		t.Fatal(err)
	}

	c3 := pb.Course{DirectoryID: 3}
	if err := db.CreateCourse(user.ID, &c3); err != nil {
		t.Fatal(err)
	}

	courses, err := db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}
	wantCourses := []*pb.Course{&c1, &c2, &c3}
	if !reflect.DeepEqual(courses, wantCourses) {
		t.Errorf("have %v want %v", courses, wantCourses)
	}
	// An empty list should return the same as no argument, it makes no
	// sense to ask the database to return no courses.
	coursesNoArg, err := db.GetCourses([]uint64{}...)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(coursesNoArg, wantCourses) {
		t.Errorf("have %v want %v", coursesNoArg, wantCourses)
	}

	course1, err := db.GetCourses(c1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantCourse1 := []*pb.Course{&c1}
	if !reflect.DeepEqual(course1, wantCourse1) {
		t.Errorf("have %v want %v", course1, wantCourse1)
	}

	course1and2, err := db.GetCourses(c1.ID, c2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantCourse1and2 := []*pb.Course{&c1, &c2}
	if !reflect.DeepEqual(course1and2, wantCourse1and2) {
		t.Errorf("have %v want %v", course1and2, wantCourse1and2)
	}
}

func TestGormDBGetAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetAssignmentsByCourse(10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBCreateAssignmentNoRecord(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	assignment := pb.Assignment{
		CourseID: 1,
		Name:     "Lab 1",
	}

	// Should fail as course 1 does not exist.
	if err := db.CreateAssignment(&assignment); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBCreateAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	user := createFakeUser(t, db, 10)
	if err := db.CreateCourse(user.ID, &pb.Course{}); err != nil {
		t.Fatal(err)
	}

	assignment := pb.Assignment{
		CourseID: 1,
		Order:    1,
	}

	if err := db.CreateAssignment(&assignment); err != nil {
		t.Fatal(err)
	}

	assignments, err := db.GetAssignmentsByCourse(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(assignments) != 1 {
		t.Fatalf("have size %v wanted %v", len(assignments), 1)
	}

	if !reflect.DeepEqual(assignments[0], &assignment) {
		t.Fatalf("want %v have %v", assignments[0], &assignment)
	}
}

func TestGormDBCreateEnrollmentNoRecord(t *testing.T) {
	const (
		userId   = 1
		courseId = 1
	)

	db, cleanup := setup(t)
	defer cleanup()

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   userId,
		CourseID: courseId,
	}); err != gorm.ErrRecordNotFound {
		t.Errorf("expected error '%v' have '%v'", gorm.ErrRecordNotFound, err)
	}
}

func TestGormDBCreateEnrollment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 1)
	var course pb.Course
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 10)
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Error(err)
	}

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err == nil {
		t.Fatal("expected duplicate enrollment creation to fail")
	}
}

func TestGormDBAcceptRejectEnrollment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 1)
	var course pb.Course
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 10)
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// Get course's pending enrollments.
	pendingEnrollments, err := db.GetEnrollmentsByCourse(course.ID, pb.Enrollment_PENDING)
	if err != nil {
		t.Fatal(err)
	}

	if len(pendingEnrollments) != 1 && pendingEnrollments[0].Status == pb.Enrollment_PENDING {
		t.Fatalf("have %v want 1 pending enrollment", pendingEnrollments)
	}

	// Accept enrollment.
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// Get course's accepted enrollments.
	acceptedEnrollments, err := db.GetEnrollmentsByCourse(course.ID, pb.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}

	if len(acceptedEnrollments) != 1 && acceptedEnrollments[0].Status == pb.Enrollment_STUDENT {
		t.Fatalf("have %v want 1 accepted enrollment", acceptedEnrollments)
	}

	// Reject enrollment.
	if err := db.RejectEnrollment(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// Get course's rejected enrollments.
	rejectedEnrollments, err := db.GetEnrollmentsByCourse(course.ID, pb.Enrollment_REJECTED)
	if err != nil {
		t.Fatal(err)
	}

	if len(rejectedEnrollments) != 1 && rejectedEnrollments[0].Status == pb.Enrollment_REJECTED {
		t.Fatalf("have %v want 1 rejected enrollment", rejectedEnrollments)
	}
}

func TestGormDBGetCoursesByUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 1)
	c1 := pb.Course{DirectoryID: 1}
	if err := db.CreateCourse(teacher.ID, &c1); err != nil {
		t.Fatal(err)
	}

	c2 := pb.Course{DirectoryID: 2}
	if err := db.CreateCourse(teacher.ID, &c2); err != nil {
		t.Fatal(err)
	}

	c3 := pb.Course{DirectoryID: 3}
	if err := db.CreateCourse(teacher.ID, &c3); err != nil {
		t.Fatal(err)
	}

	c4 := pb.Course{DirectoryID: 4}
	if err := db.CreateCourse(teacher.ID, &c4); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 10)
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: c1.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: c2.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: c3.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.ID, c2.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, c3.ID); err != nil {
		t.Fatal(err)
	}

	courses, err := db.GetCoursesByUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*pb.Course{
		{ID: c1.ID, DirectoryID: 1, Enrolled: pb.Enrollment_PENDING},
		{ID: c2.ID, DirectoryID: 2, Enrolled: pb.Enrollment_REJECTED},
		{ID: c3.ID, DirectoryID: 3, Enrolled: pb.Enrollment_STUDENT},
		{ID: c4.ID, DirectoryID: 4, Enrolled: pb.Enrollment_NONE},
	}
	if !reflect.DeepEqual(courses, wantCourses) {
		t.Errorf("have course %+v want %+v", courses, wantCourses)
	}
}

func TestGetRemoteIdentity(t *testing.T) {
	const (
		provider = "github"
		remoteID = 10
	)

	db, cleanup := setup(t)
	defer cleanup()

	var user pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&pb.RemoteIdentity{
			Provider: provider,
			RemoteID: remoteID,
		},
	); err != nil {
		t.Fatal(err)
	}
	if len(user.RemoteIdentities) != 1 {
		t.Fatalf("have %d remote identites want %d", len(user.RemoteIdentities), 1)
	}

	remoteIdentity, err := db.GetRemoteIdentity(provider, remoteID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(remoteIdentity, user.RemoteIdentities[0]) {
		t.Errorf("have remote identity %+v want %+v", remoteIdentity, user.RemoteIdentities[0])
	}
}

func TestGormDBDuplicateIdentity(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if err := db.CreateUserFromRemoteIdentity(
		&pb.User{}, &pb.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	if err := db.CreateUserFromRemoteIdentity(
		&pb.User{}, &pb.RemoteIdentity{},
	); err == nil {
		t.Fatal("expected duplicate remote identity creation to fail")
	}
}

func TestGormDBAssociateUserWithRemoteIdentity(t *testing.T) {
	const (
		uID  = 2
		rID1 = 2
		rID2 = 3

		secret1   = "123"
		provider1 = "github"
		remoteID1 = 10

		secret2   = "ABC"
		provider2 = "gitlab"
		remoteID2 = 20

		secret3 = "DEF"
	)

	var (
		wantUser1 = &pb.User{
			ID: uID,
			RemoteIdentities: []*pb.RemoteIdentity{{
				ID:          rID1,
				Provider:    provider1,
				RemoteID:    remoteID1,
				AccessToken: secret1,
				UserID:      uID,
			}},
		}

		wantUser2 = &pb.User{
			ID: uID,
			RemoteIdentities: []*pb.RemoteIdentity{
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

	// Create first user (the admin).
	if err := db.CreateUserFromRemoteIdentity(
		&pb.User{},
		&pb.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	var user1 pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user1,
		&pb.RemoteIdentity{
			Provider:    provider1,
			RemoteID:    remoteID1,
			AccessToken: secret1,
		},
	); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(&user1, wantUser1) {
		t.Errorf("have user %+v want %+v", &user1, wantUser1)
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
		github = "github"
		gitlab = "gitlab"
	)

	db, cleanup := setup(t)
	defer cleanup()

	// Create first user (the admin).
	if err := db.CreateUserFromRemoteIdentity(
		&pb.User{},
		&pb.RemoteIdentity{
			Provider: github,
		},
	); err != nil {
		t.Fatal(err)
	}

	var user pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&pb.RemoteIdentity{
			Provider: gitlab,
		},
	); err != nil {
		t.Fatal(err)
	}

	if user.IsAdmin {
		t.Error("user should not yet be an administrator")
	}

	if err := db.SetAdmin(user.ID); err != nil {
		t.Error(err)
	}

	admin, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !admin.IsAdmin {
		t.Error("user should be an administrator")
	}
}

func TestGormDBCreateCourse(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := pb.Course{
		Name: "name",
		Code: "code",
		Year: 2017,
		Tag:  "tag",

		Provider:    "github",
		DirectoryID: 1,
	}

	user := createFakeUser(t, db, 10)
	if err := db.CreateCourse(user.ID, &course); err != nil {
		t.Fatal(err)
	}

	if course.ID == 0 {
		t.Error("expected id to be set")
	}
}

func TestGormDBCreateCourseNonAdmin(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 10)
	if err := db.CreateCourse(admin.ID, &pb.Course{}); err != nil {
		t.Fatal(err)
	}
	nonAdmin := createFakeUser(t, db, 11)
	// the following should fail to create a course
	if err := db.CreateCourse(nonAdmin.ID, &pb.Course{}); err == nil {
		t.Fatal(err)
	}
}

func TestGormDBGetCourse(t *testing.T) {
	course := &pb.Course{
		Name:        "Test Course",
		Code:        "DAT100",
		Year:        2017,
		Tag:         "Spring",
		Provider:    "github",
		DirectoryID: 1234,
	}

	db, cleanup := setup(t)
	defer cleanup()

	user := createFakeUser(t, db, 10)
	if err := db.CreateCourse(user.ID, course); err != nil {
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

func TestGormDBGetCourseByDirectory(t *testing.T) {
	course := &pb.Course{
		Name:        "Test Course",
		Code:        "DAT100",
		Year:        2017,
		Tag:         "Spring",
		Provider:    "github",
		DirectoryID: 1234,
	}

	db, cleanup := setup(t)
	defer cleanup()

	user := createFakeUser(t, db, 10)
	if err := db.CreateCourse(user.ID, course); err != nil {
		t.Fatal(err)
	}

	// Get the created course.
	createdCourse, err := db.GetCourseByDirectoryID(course.DirectoryID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(createdCourse, course) {
		t.Errorf("have course %+v want %+v", createdCourse, course)
	}

}

func TestGormDBGetCourseNoRecord(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetCourse(10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}

}

func TestGormDBUpdateCourse(t *testing.T) {
	var (
		course = &pb.Course{
			Name:        "Test Course",
			Code:        "DAT100",
			Year:        2017,
			Tag:         "Spring",
			Provider:    "github",
			DirectoryID: 1234,
		}
		updates = &pb.Course{
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

	user := createFakeUser(t, db, 10)
	if err := db.CreateCourse(user.ID, course); err != nil {
		t.Fatal(err)
	}

	updates.ID = course.ID
	if err := db.UpdateCourse(updates); err != nil {
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

func TestGormDBGetSubmissionForUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetSubmissionForUser(10, 10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetSubmissionByID(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if sub, err := db.GetSubmissionsByID(100); err != gorm.ErrRecordNotFound {
		t.Errorf("got submission %v", sub)
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetNonExsistingSubmissions(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetSubmissions(10, 10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBInsertSubmissions(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       1,
	}); err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 10)
	// create a course and an assignment
	var course pb.Course
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	assigment := pb.Assignment{
		CourseID: course.ID,
		Order:    1,
	}
	if err := db.CreateAssignment(&assigment); err != nil {
		t.Fatal(err)
	}

	// create a submission for the assignment; should fail
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assigment.ID,
		UserID:       2,
	}); err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 11)
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// create another submission for the assignment; now it should succeed
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assigment.ID,
		UserID:       user.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// confirm that the submission is in the database
	submissions, err := db.GetSubmissions(course.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}
	want := &pb.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assigment.ID,
		UserID:       user.ID,
	}
	if !reflect.DeepEqual(submissions[0], want) {
		t.Errorf("have %#v want %#v", submissions[0], want)
	}
}

func TestGormDBGetInsertSubmissions(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 10)
	// Create course c1 and c2
	c1 := pb.Course{DirectoryID: 1}
	if err := db.CreateCourse(teacher.ID, &c1); err != nil {
		t.Fatal(err)
	}
	c2 := pb.Course{DirectoryID: 2}
	if err := db.CreateCourse(teacher.ID, &c2); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 11)

	// enroll student in course c1
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: c1.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, c1.ID); err != nil {
		t.Fatal(err)
	}

	// Create some assignments
	assignment1 := pb.Assignment{
		Order:    1,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := pb.Assignment{
		Order:    2,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := pb.Assignment{
		Order:    1,
		CourseID: c2.ID,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// Create some submissions
	submission1 := pb.Submission{
		UserID:       user.ID,
		AssignmentID: assignment1.ID,
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := pb.Submission{
		UserID:       user.ID,
		AssignmentID: assignment1.ID,
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := pb.Submission{
		UserID:       user.ID,
		AssignmentID: assignment2.ID,
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}

	// Even if there is three submission, only the latest for each assignment should be returned

	submissions, err := db.GetSubmissions(c1.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := []*pb.Submission{&submission2, &submission3}
	if !reflect.DeepEqual(submissions, want) {
		for _, s := range submissions {
			fmt.Printf("%+v\n", s)
		}
		t.Errorf("have %#v want %#v", submissions, want)
	}
	data, err := db.GetSubmissions(c1.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	// Since there is no submissions, but the course and user exist, an empty array should be returned
	data, err = db.GetSubmissions(c2.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 0, len(data))
	}
}

var createGroupTests = []struct {
	name        string
	getGroup    func(uint64, ...uint64) *pb.Group
	enrollments []uint
	err         error
}{
	// Should fail with ErrRecordNotFound as we cannot create a group that
	// is not connected to a course.
	{
		name: "course id not set",
		getGroup: func(uint64, ...uint64) *pb.Group {
			return &pb.Group{}
		},
		err: gorm.ErrRecordNotFound,
	},
	// Should fail with ErrRecordNotFound as we cannot create a group that
	// is not connected to a course.
	{
		name: "course not found",
		getGroup: func(uint64, ...uint64) *pb.Group {
			return &pb.Group{CourseID: 999}
		},
		err: gorm.ErrRecordNotFound,
	},
	// Should pass as long as it's desirable to create a group without any
	// users.
	// TODO: This is probably fine, but there needs to be a len(users) > 1
	// check in the web handler.
	{
		name: "course found",
		getGroup: func(cid uint64, _ ...uint64) *pb.Group {
			return &pb.Group{CourseID: cid}
		},
	},
	// Should fail with ErrRecordNotFound as we cannot create a group with
	// users that doesn't exist.
	{
		name: "with non existing users",
		getGroup: func(cid uint64, _ ...uint64) *pb.Group {
			return &pb.Group{
				CourseID: cid,
				Users: []*pb.User{
					{ID: 101},
					{ID: 102},
				},
			}
		},

		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         gorm.ErrRecordNotFound,
	},
	// Should fail with ErrRecordNotFound as we cannot create a group with
	// users that's not enrolled in the course.
	{
		name: "with users but without enrollments",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         gorm.ErrRecordNotFound,
	},
	// Should fail with ErrRecordNotFound as we cannot create a group with
	// users that's not enrolled in the course.
	{
		name: "with users and pending enrollments",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         gorm.ErrRecordNotFound,
	},
	// Should fail with ErrRecordNotFound as we cannot create a group with
	// users that's not enrolled in the course.
	{
		name: "with users and rejected enrollments",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_REJECTED), uint(pb.Enrollment_REJECTED)},
		err:         gorm.ErrRecordNotFound,
	},
	// Should pass as the user exists and is enrolled in the course.
	{
		name: "with user and accepted enrollment",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_STUDENT)},
	},
	// Should pass as the users exists and are enrolled in the course.
	{
		name: "with users and accepted enrollments",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_STUDENT), uint(pb.Enrollment_STUDENT)},
	},
}

func TestGormDBCreateAndGetGroup(t *testing.T) {
	for _, test := range createGroupTests {
		t.Run(test.name, func(t *testing.T) {
			db, cleanup := setup(t)

			teacher := createFakeUser(t, db, 10)
			var course pb.Course
			if err := db.CreateCourse(teacher.ID, &course); err != nil {
				t.Fatal(err)
			}
			var uids []uint64
			// create as many users as the desired number of enrollments
			for i := 0; i < len(test.enrollments); i++ {
				user := createFakeUser(t, db, uint64(i))
				uids = append(uids, user.ID)
			}
			// enroll users in course
			//TODO(meling) this loop and the one above can be merged, I think
			for i := 0; i < len(uids); i++ {
				if test.enrollments[i] == uint(pb.Enrollment_PENDING) {
					continue
				}
				if err := db.CreateEnrollment(&pb.Enrollment{
					CourseID: course.ID,
					UserID:   uids[i],
				}); err != nil {
					t.Fatal(err)
				}
				err := errors.New("enrollment status not implemented")
				switch test.enrollments[i] {
				case uint(pb.Enrollment_REJECTED):
					err = db.RejectEnrollment(uids[i], course.ID)
				case uint(pb.Enrollment_STUDENT):
					err = db.EnrollStudent(uids[i], course.ID)
				}
				if err != nil {
					t.Fatal(err)
				}
			}

			// Test.
			group := test.getGroup(course.ID, uids...)
			if err := db.CreateGroup(group); err != test.err {
				t.Errorf("have error '%v' want '%v'", err, test.err)
			}
			if test.err != nil {
				return
			}

			// Verify.
			enrollments, err := db.GetEnrollmentsByCourse(course.ID, pb.Enrollment_STUDENT)
			if err != nil {
				t.Fatal(err)
			}
			if len(group.Users) > 0 && len(enrollments) != len(group.Users) {
				t.Errorf("have %d enrollments want %d", len(enrollments), len(group.Users))
			}
			sorted := make(map[uint64]*pb.Enrollment)
			for _, enrollment := range enrollments {
				sorted[enrollment.UserID] = enrollment
			}
			for _, user := range group.Users {
				if _, ok := sorted[user.ID]; !ok {
					t.Errorf("have no enrollment for user %d", user.ID)
				}
			}

			have, err := db.GetGroup(group.ID)
			if err != nil {
				t.Fatal(err)
			}
			if len(uids) > 0 {
				group.Users, err = db.GetUsers(uids...)
				if err != nil {
					t.Fatal(err)
				}
			}
			group.Enrollments = enrollments
			if !reflect.DeepEqual(have, group) {
				t.Errorf("have %#v want %#v", have, group)
			}
			cleanup()
		})
	}
}

func TestGormDBCreateGroupTwice(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 10)
	var course pb.Course
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	var users []*pb.User
	enrollments := []pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := createFakeUser(t, db, uint64(i))
		users = append(users, user)
	}
	// enroll users in course
	for i := 0; i < len(users); i++ {
		if enrollments[i] == pb.Enrollment_PENDING {
			continue
		}
		if err := db.CreateEnrollment(&pb.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		switch enrollments[i] {
		case pb.Enrollment_STUDENT:
			err = db.EnrollStudent(users[i].ID, course.ID)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	// Try to create two identical groups. The first should succeed while
	// further attempts should fail with ErrDuplicateGroup.
	identical := &pb.Group{
		Name:     "SameNameGroup",
		CourseID: course.ID,
		Users:    users,
	}
	if err := db.CreateGroup(identical); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateGroup(identical); err != database.ErrDuplicateGroup {
		t.Fatalf("expected error '%v' have '%v'", database.ErrDuplicateGroup, err)
	}
}

func TestGormDBGetEmptyRepo(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()
	if _, err := db.GetRepository(10); err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}
}

// createFakeUser is a test helper to create a user in the database
// with the given remote id and the fake scm provider.
func createFakeUser(t *testing.T, db database.Database, remoteID uint64) *pb.User {
	var user pb.User
	err := db.CreateUserFromRemoteIdentity(&user,
		&pb.RemoteIdentity{
			Provider: "fake",
			RemoteID: remoteID,
		})
	if err != nil {
		t.Fatal(err)
	}
	return &user
}

func TestGormDBGetSingleRepoWithUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	user := createFakeUser(t, db, 10)
	repo := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 100,
		UserID:       user.ID,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetRepository(repo.RepositoryID); err != nil {
		t.Fatal(err)
	}
}

func TestGormDBCreateSingleRepoWithMissingUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	repo := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 100,
		UserID:       20,
	}
	if err := db.CreateRepository(&repo); err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}
}

func TestGormDBGetCourseRepoType(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	repo := pb.Repository{
		DirectoryID:  120,
		RepositoryID: 100,
		RepoType:     pb.Repository_COURSEINFO,
		// Name:         "Name",
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	gotRepo, err := db.GetRepository(repo.RepositoryID)
	if err != nil {
		t.Fatal(err)
	}
	if !gotRepo.RepoType.IsCourseRepo() {
		t.Fatalf("Expected course info repo (%v), but got: %v", pb.Repository_COURSEINFO, gotRepo.RepoType)
	}
}

func TestGormDBGetGroupSubmissions(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if sub, err := db.GetGroupSubmissions(10, 10); err != gorm.ErrRecordNotFound {
		t.Errorf("got submission %v", sub)
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetInsertGroupSubmissions(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 10)
	course := pb.Course{DirectoryID: 1}
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	courseTwo := pb.Course{DirectoryID: 2}
	if err := db.CreateCourse(teacher.ID, &courseTwo); err != nil {
		t.Fatal(err)
	}

	var users []*pb.User
	enrollments := []pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := createFakeUser(t, db, uint64(i))
		users = append(users, user)
	}
	// enroll users in course
	for i := 0; i < len(users); i++ {
		if enrollments[i] == pb.Enrollment_PENDING {
			continue
		}
		if err := db.CreateEnrollment(&pb.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		switch enrollments[i] {
		case pb.Enrollment_STUDENT:
			err = db.EnrollStudent(users[i].ID, course.ID)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	// Creating Group
	group := &pb.Group{
		Name:     "SameNameGroup",
		CourseID: course.ID,
		Users:    users,
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	// Create Assignments
	assignment1 := pb.Assignment{
		Order:      1,
		CourseID:   course.ID,
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := pb.Assignment{
		Order:      2,
		CourseID:   course.ID,
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := pb.Assignment{
		Order:      1,
		CourseID:   courseTwo.ID,
		IsGroupLab: false,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// Create some submissions
	submission1 := pb.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment1.ID,
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := pb.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment1.ID,
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := pb.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment2.ID,
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}
	submission4 := pb.Submission{
		UserID:       users[0].ID,
		AssignmentID: assignment3.ID,
	}
	if err := db.CreateSubmission(&submission4); err != nil {
		t.Fatal(err)
	}

	// Even if there is three submission, only the latest for each assignment should be returned

	submissions, err := db.GetGroupSubmissions(course.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := []*pb.Submission{&submission2, &submission3}
	if !reflect.DeepEqual(submissions, want) {
		for _, s := range submissions {
			fmt.Printf("%+v\n", s)
		}
		t.Errorf("have %#v want %#v", submissions, want)
	}
	data, err := db.GetGroupSubmissions(course.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	// Since there is no submissions, but the course and user exist, an empty array should be returned
	data, err = db.GetGroupSubmissions(courseTwo.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 0, len(data))
	}
}

func TestGetRepositoriesByDirectory(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := &pb.Course{
		Name:        "Test Course",
		Code:        "DAT100",
		Year:        2017,
		Tag:         "Spring",
		Provider:    "github",
		DirectoryID: 1234,
	}

	teacher := createFakeUser(t, db, 10)
	if err := db.CreateCourse(teacher.ID, course); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 100,
		UserID:       user.ID,
		RepoType:     pb.Repository_COURSEINFO,
		HTMLURL:      "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating solution
	repoSolution := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 101,
		UserID:       user.ID,
		RepoType:     pb.Repository_SOLUTIONS,
		HTMLURL:      "http://repoSolution.com/",
	}
	if err := db.CreateRepository(&repoSolution); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 102,
		UserID:       user.ID,
		RepoType:     pb.Repository_ASSIGNMENTS,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	want := []*pb.Repository{&repoCourseInfo, &repoSolution, &repoAssignment}

	gotRepo, err := db.GetRepositoriesByDirectory(120)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotRepo, want) {
		for _, s := range gotRepo {
			fmt.Printf("have %+v\n", s)
		}
		fmt.Println("")
		for _, s := range want {
			fmt.Printf("want %+v\n", s)
		}
		t.Errorf("Failed")
	}
}

func TestDeleteGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 10)
	var course pb.Course
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	var users []*pb.User
	enrollments := []pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := createFakeUser(t, db, uint64(i))
		users = append(users, user)
	}
	// enroll users in course
	for i := 0; i < len(users); i++ {
		if enrollments[i] == pb.Enrollment_PENDING {
			continue
		}
		if err := db.CreateEnrollment(&pb.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		switch enrollments[i] {
		case pb.Enrollment_STUDENT:
			err = db.EnrollStudent(users[i].ID, course.ID)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	group := &pb.Group{
		Name:     "SameNameGroup",
		CourseID: course.ID,
		Users:    users,
		ID:       1,
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	err := db.DeleteGroup(1)
	if err != nil {
		t.Fatal(err)
	}

	gotModels, _ := db.GetGroup(group.ID)
	if gotModels != nil {
		t.Errorf("Got %+v wanted None", gotModels)
	}
}

func TestGetRepositoriesByCourseIdAndType(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := &pb.Course{
		Name:        "Test Course",
		Code:        "DAT100",
		Year:        2017,
		Tag:         "Spring",
		Provider:    "github",
		DirectoryID: 1234,
		ID:          1,
	}

	teacher := createFakeUser(t, db, 10)
	if err := db.CreateCourse(teacher.ID, course); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := pb.Repository{
		DirectoryID: 1234,
		// Name:         "Name",
		RepositoryID: 100,
		UserID:       user.ID,
		RepoType:     pb.Repository_COURSEINFO,
		HTMLURL:      "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating solution
	repoSolution := pb.Repository{
		DirectoryID: 1234,
		// Name:         "Name",
		RepositoryID: 101,
		UserID:       user.ID,
		RepoType:     pb.Repository_SOLUTIONS,
		HTMLURL:      "http://repoSolution.com/",
	}
	if err := db.CreateRepository(&repoSolution); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := pb.Repository{
		DirectoryID: 1234,
		// Name:         "Name",
		RepositoryID: 102,
		UserID:       user.ID,
		RepoType:     pb.Repository_ASSIGNMENTS,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	want := []*pb.Repository{&repoCourseInfo}

	gotRepo, err := db.GetRepositoriesByCourseAndType(course.ID, pb.Repository_COURSEINFO)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotRepo, want) {
		//t.Errorf("Failed")
		for _, s := range gotRepo {
			t.Logf("have %+v\n", s)
		}
		for _, s := range want {
			t.Logf("want %+v\n", s)
		}
		t.Errorf("Failed")
	}
}

func TestGetRepoByCourseIdUserIdandType(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := &pb.Course{
		ID:          1234,
		Name:        "Test Course",
		Code:        "DAT100",
		Year:        2017,
		Tag:         "Spring",
		Provider:    "github",
		DirectoryID: 120,
	}

	teacher := createFakeUser(t, db, 1)
	if err := db.CreateCourse(teacher.ID, course); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 10)
	userTwo := createFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 100,
		UserID:       user.ID,
		RepoType:     pb.Repository_COURSEINFO,
		HTMLURL:      "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating solution
	repoSolution := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 101,
		UserID:       user.ID,
		RepoType:     pb.Repository_SOLUTIONS,
		HTMLURL:      "http://repoSolution.com/",
	}
	if err := db.CreateRepository(&repoSolution); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 102,
		UserID:       user.ID,
		RepoType:     pb.Repository_ASSIGNMENTS,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for user
	repoUser := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 103,
		UserID:       user.ID,
		RepoType:     pb.Repository_USER,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUser); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for userTwo
	repoUserTwo := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 104,
		UserID:       userTwo.ID,
		RepoType:     pb.Repository_USER,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUserTwo); err != nil {
		t.Fatal(err)
	}

	want := &repoUserTwo

	gotRepo, err := db.GetRepositoryByCourseUserType(course.ID, userTwo.ID, pb.Repository_USER)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotRepo, want) {
		fmt.Printf("have %+v want %+v\n", gotRepo, want)
		t.Errorf("Failed")
	}
}

func TestGetRepositoriesByCourseIdandUserId(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := &pb.Course{
		ID:          1234,
		Name:        "Test Course",
		Code:        "DAT100",
		Year:        2017,
		Tag:         "Spring",
		Provider:    "github",
		DirectoryID: 120,
	}

	teacher := createFakeUser(t, db, 1)
	if err := db.CreateCourse(teacher.ID, course); err != nil {
		t.Fatal(err)
	}
	user := createFakeUser(t, db, 10)
	userTwo := createFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 100,
		UserID:       user.ID,
		RepoType:     pb.Repository_COURSEINFO,
		HTMLURL:      "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating solution
	repoSolution := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 101,
		UserID:       user.ID,
		RepoType:     pb.Repository_SOLUTIONS,
		HTMLURL:      "http://repoSolution.com/",
	}
	if err := db.CreateRepository(&repoSolution); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 102,
		UserID:       user.ID,
		RepoType:     pb.Repository_ASSIGNMENTS,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for user
	repoUser := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 103,
		UserID:       user.ID,
		RepoType:     pb.Repository_USER,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUser); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for userTwo
	repoUserTwo := pb.Repository{
		DirectoryID: 120,
		// Name:         "Name",
		RepositoryID: 104,
		UserID:       userTwo.ID,
		RepoType:     pb.Repository_USER,
		HTMLURL:      "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUserTwo); err != nil {
		t.Fatal(err)
	}

	want := &repoUserTwo

	gotRepo, err := db.GetRepositoryByCourseUserType(course.ID, userTwo.ID, pb.Repository_USER)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotRepo, want) {
		t.Errorf("\nhave %+v\nwant %+v\n", gotRepo, want)
	}
}

func envSet(env string) database.GormLogger {
	if os.Getenv(env) != "" {
		return database.Logger{Logger: logrus.New()}
	}
	return nil
}
