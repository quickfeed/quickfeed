package database_test

import (
	"errors"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGormDBGetUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if _, err := db.GetUser(10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetUsers(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if _, err := db.GetUsers(); err != nil {
		t.Errorf("have error '%v' wanted '%v'", err, nil)
	}
}

func TestGormDBGetUserWithEnrollments(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 11)
	course := &pb.Course{}
	qtest.CreateCourse(t, db, admin, course)

	student := qtest.CreateFakeUser(t, db, 13)
	if err := db.CreateEnrollment(&pb.Enrollment{
		CourseID: course.ID,
		UserID:   student.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &pb.Enrollment{
		UserID:   student.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	// user entries from the database will have to be enrolled as
	// teacher and student respectively
	admin.Enrollments = append(admin.Enrollments, &pb.Enrollment{
		ID:           1,
		CourseID:     course.ID,
		UserID:       admin.ID,
		Status:       pb.Enrollment_TEACHER,
		State:        pb.Enrollment_VISIBLE,
		UsedSlipDays: []*pb.UsedSlipDays{},
	})
	admin.RemoteIdentities = nil

	student.Enrollments = append(student.Enrollments, &pb.Enrollment{
		ID:           2,
		CourseID:     course.ID,
		UserID:       student.ID,
		Status:       pb.Enrollment_STUDENT,
		State:        pb.Enrollment_VISIBLE,
		UsedSlipDays: []*pb.UsedSlipDays{},
	})
	student.RemoteIdentities = nil

	gotTeacher, err := db.GetUserWithEnrollments(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(admin, gotTeacher, protocmp.Transform()); diff != "" {
		t.Errorf("enrollment mismatch (-teacher +gotTeacher):\n%s", diff)
	}
	gotStudent, err := db.GetUserWithEnrollments(student.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(student, gotStudent, protocmp.Transform()); diff != "" {
		t.Errorf("enrollment mismatch (-student +gotStudent):\n%s", diff)
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
			IsAdmin:   true, // have to set IsAdmin or will be switched back to false
		}
	)

	db, cleanup := qtest.TestDB(t)
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

	gotUser, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	gotUser.Enrollments = nil
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}

	// check that admin role can be revoked
	updates.IsAdmin = false
	wantUser.IsAdmin = false
	if err := db.UpdateUser(updates); err != nil {
		t.Fatal(err)
	}
	gotUser, err = db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	gotUser.Enrollments = nil
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}
}

func TestGormDBGetCourses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	c1 := &pb.Course{OrganizationID: 1}
	c2 := &pb.Course{OrganizationID: 2}
	c3 := &pb.Course{OrganizationID: 3}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)
	qtest.CreateCourse(t, db, admin, c3)

	gotCourses, err := db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}
	wantCourses := []*pb.Course{c1, c2, c3}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	// An empty list should return the same as no argument, it makes no
	// sense to ask the database to return no courses.
	gotCourses, err = db.GetCourses([]uint64{}...)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	gotCourses, err = db.GetCourses(c1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses = []*pb.Course{c1}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	gotCourses, err = db.GetCourses(c1.ID, c2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses = []*pb.Course{c1, c2}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}
}

func TestGormDBCreateEnrollmentNoRecord(t *testing.T) {
	const (
		userId   = 1
		courseId = 1
	)

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   userId,
		CourseID: courseId,
	}); err != gorm.ErrRecordNotFound {
		t.Errorf("expected error '%v' have '%v'", gorm.ErrRecordNotFound, err)
	}
}

func TestGormDBCreateEnrollment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	course := &pb.Course{}
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 10)
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
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	course := &pb.Course{}
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 10)
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
	query := &pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
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

	// Get all enrollments.
	allEnrollments, err := db.GetEnrollmentsByCourse(course.ID)
	if err != nil {
		t.Fatal(err)
	}

	for _, enrol := range allEnrollments {
		if enrol.UserID == user.ID && enrol.CourseID == course.ID {
			t.Fatalf("Enrollment %+v must have been deleted on rejection, but still found in the database", enrol)
		}
	}
}

func TestGormDBGetCoursesByUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	c1 := &pb.Course{OrganizationID: 1}
	c2 := &pb.Course{OrganizationID: 2}
	c3 := &pb.Course{OrganizationID: 3}
	c4 := &pb.Course{OrganizationID: 4}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)
	qtest.CreateCourse(t, db, admin, c3)
	qtest.CreateCourse(t, db, admin, c4)

	user := qtest.CreateFakeUser(t, db, 10)
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
	query := &pb.Enrollment{
		UserID:   user.ID,
		CourseID: c3.ID,
		Status:   pb.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	gotCourses, err := db.GetCoursesByUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*pb.Course{
		{ID: c1.ID, OrganizationID: 1, CourseCreatorID: admin.ID, Provider: "fake", Enrolled: pb.Enrollment_PENDING},
		{ID: c2.ID, OrganizationID: 2, CourseCreatorID: admin.ID, Provider: "fake", Enrolled: pb.Enrollment_NONE},
		{ID: c3.ID, OrganizationID: 3, CourseCreatorID: admin.ID, Provider: "fake", Enrolled: pb.Enrollment_STUDENT},
		{ID: c4.ID, OrganizationID: 4, CourseCreatorID: admin.ID, Provider: "fake", Enrolled: pb.Enrollment_NONE},
	}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCoursesByUser() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}
}

func TestGormDBDuplicateIdentity(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
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

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// Create first user (the admin).
	if err := db.CreateUserFromRemoteIdentity(
		&pb.User{},
		&pb.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	gotUser1 := &pb.User{}
	if err := db.CreateUserFromRemoteIdentity(
		gotUser1,
		&pb.RemoteIdentity{
			Provider:    provider1,
			RemoteID:    remoteID1,
			AccessToken: secret1,
		},
	); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantUser1, gotUser1, protocmp.Transform()); diff != "" {
		t.Errorf("CreateUserFromRemoteIdentity() mismatch (-wantUser1, +gotUser1):\n%s", diff)
	}

	if err := db.AssociateUserWithRemoteIdentity(gotUser1.ID, provider2, remoteID2, secret2); err != nil {
		t.Fatal(err)
	}

	gotUser2, err := db.GetUser(uID)
	if err != nil {
		t.Fatal(err)
	}
	gotUser2.Enrollments = nil

	if diff := cmp.Diff(wantUser2, gotUser2, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser2, +gotUser2):\n%s", diff)
	}

	if err := db.AssociateUserWithRemoteIdentity(gotUser1.ID, provider2, remoteID2, secret3); err != nil {
		t.Fatal(err)
	}

	gotUser3, err := db.GetUser(uID)
	if err != nil {
		t.Fatal(err)
	}
	gotUser3.Enrollments = nil
	wantUser2.RemoteIdentities[1].AccessToken = secret3

	if diff := cmp.Diff(wantUser2, gotUser3, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser2, +gotUser3):\n%s", diff)
	}
}

func TestGormDBSetAdminNoRecord(t *testing.T) {
	const id = 1

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if err := db.UpdateUser(&pb.User{ID: id, IsAdmin: true}); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBSetAdmin(t *testing.T) {
	const (
		github = "github"
		gitlab = "gitlab"
	)

	db, cleanup := qtest.TestDB(t)
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

	if err := db.UpdateUser(&pb.User{ID: user.ID, IsAdmin: true}); err != nil {
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
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &pb.Course{
		Name:           "name",
		Code:           "code",
		Year:           2017,
		Tag:            "tag",
		Provider:       "github",
		OrganizationID: 1,
	}

	remoteID := &pb.RemoteIdentity{Provider: course.Provider, RemoteID: 10, AccessToken: "token"}
	admin := qtest.CreateUserFromRemoteIdentity(t, db, remoteID)
	qtest.CreateCourse(t, db, admin, course)
	if course.ID == 0 {
		t.Error("expected id to be set")
	}

	// check that admin (teacher) was automatically enrolled when creating course
	enroll, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if enroll.CourseID != course.ID || enroll.UserID != admin.ID {
		t.Errorf("expected user %d to be enrolled in course %d, but got user %d and course %d", admin.ID, course.ID, enroll.UserID, enroll.CourseID)
	}
	if enroll.Status != pb.Enrollment_TEACHER || enroll.State != pb.Enrollment_VISIBLE {
		t.Errorf("expected enrolled user to be teacher and visible, but got status: %v and state: %v", enroll.Status, enroll.State)
	}

	// check that no users were enrolled as students
	enrolls, err := db.GetEnrollmentsByCourse(course.ID, pb.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	if len(enrolls) > 0 {
		t.Errorf("expected no enrollments, but got %d enrollments: %v", len(enrolls), enrolls)
	}

	// check that exactly one user was enrolled as teacher for the course
	enrolls, err = db.GetEnrollmentsByCourse(course.ID, pb.Enrollment_TEACHER)
	if err != nil {
		t.Fatal(err)
	}
	if len(enrolls) != 1 {
		t.Errorf("expected exactly one enrollment, but got %d enrollments: %v", len(enrolls), enrolls)
	}
	for _, enroll := range enrolls {
		if enroll.CourseID != course.ID || enroll.UserID != admin.ID {
			t.Errorf("expected user %d to be enrolled in course %d, but got user %d and course %d", admin.ID, course.ID, enroll.UserID, enroll.CourseID)
		}
		if enroll.Status != pb.Enrollment_TEACHER || enroll.State != pb.Enrollment_VISIBLE {
			t.Errorf("expected enrolled user to be teacher and visible, but got status: %v and state: %v", enroll.Status, enroll.State)
		}
	}
}

func TestGormDBCreateCourseNonAdmin(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	qtest.CreateCourse(t, db, admin, &pb.Course{})

	nonAdmin := qtest.CreateFakeUser(t, db, 11)
	// the following should fail to create a course
	if err := db.CreateCourse(nonAdmin.ID, &pb.Course{}); err == nil {
		t.Fatal(err)
	}
}

func TestGormDBGetCourse(t *testing.T) {
	wantCourse := &pb.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		Provider:       "github",
		OrganizationID: 1234,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	remoteID := &pb.RemoteIdentity{Provider: wantCourse.Provider, RemoteID: 10, AccessToken: "token"}
	admin := qtest.CreateUserFromRemoteIdentity(t, db, remoteID)
	qtest.CreateCourse(t, db, admin, wantCourse)

	// Get the created course.
	gotCourse, err := db.GetCourse(wantCourse.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse, +gotCourse):\n%s", diff)
	}
}

func TestGormDBGetCourseByOrganization(t *testing.T) {
	wantCourse := &pb.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		Provider:       "github",
		OrganizationID: 1234,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	remoteID := &pb.RemoteIdentity{Provider: wantCourse.Provider, RemoteID: 10, AccessToken: "token"}
	admin := qtest.CreateUserFromRemoteIdentity(t, db, remoteID)
	qtest.CreateCourse(t, db, admin, wantCourse)

	// Get the created course.
	gotCourse, err := db.GetCourseByOrganizationID(wantCourse.OrganizationID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse, +gotCourse):\n%s", diff)
	}
}

func TestGormDBGetCourseNoRecord(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if _, err := db.GetCourse(10, false); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBUpdateCourse(t *testing.T) {
	course := &pb.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		Provider:       "github",
		OrganizationID: 1234,
	}
	wantCourse := &pb.Course{
		Name:           "Test Course Edit",
		Code:           "DAT100-1",
		Year:           2018,
		Tag:            "Autumn",
		Provider:       "gitlab",
		OrganizationID: 12345,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	remoteID := &pb.RemoteIdentity{Provider: course.Provider, RemoteID: 10, AccessToken: "token"}
	admin := qtest.CreateUserFromRemoteIdentity(t, db, remoteID)
	qtest.CreateCourse(t, db, admin, course)

	wantCourse.ID = course.ID
	wantCourse.CourseCreatorID = admin.ID
	if err := db.UpdateCourse(wantCourse); err != nil {
		t.Fatal(err)
	}

	// Get the updated course.
	gotCourse, err := db.GetCourse(course.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("course mismatch (-want +got):\n%s", diff)
	}
}

func TestGormDBGetGroupSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if sub, err := db.GetLastSubmissions(10, &pb.Submission{GroupID: 10}); err != gorm.ErrRecordNotFound {
		t.Errorf("got submission %v", sub)
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetInsertGroupSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	c1 := &pb.Course{OrganizationID: 1}
	c2 := &pb.Course{OrganizationID: 2}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)

	var users []*pb.User
	enrollments := []pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db, uint64(i))
		users = append(users, user)
	}
	// enroll users in course
	for i := 0; i < len(users); i++ {
		if enrollments[i] == pb.Enrollment_PENDING {
			continue
		}
		if err := db.CreateEnrollment(&pb.Enrollment{
			CourseID: c1.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		switch enrollments[i] {
		case pb.Enrollment_STUDENT:
			query := &pb.Enrollment{
				UserID:   users[i].ID,
				CourseID: c1.ID,
				Status:   pb.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	// Creating Group
	group := &pb.Group{
		Name:     "SameNameGroup",
		CourseID: c1.ID,
		Users:    users,
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	// Create Assignments
	assignment1 := pb.Assignment{
		Order:      1,
		CourseID:   c1.ID,
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := pb.Assignment{
		Order:      2,
		CourseID:   c1.ID,
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := pb.Assignment{
		Order:      1,
		CourseID:   c2.ID,
		IsGroupLab: false,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// Create some submissions
	submission1 := pb.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := pb.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := pb.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment2.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}
	submission4 := pb.Submission{
		UserID:       users[0].ID,
		AssignmentID: assignment3.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission4); err != nil {
		t.Fatal(err)
	}

	// Even if there is three submission, only the latest for each assignment should be returned

	submissions, err := db.GetLastSubmissions(c1.ID, &pb.Submission{GroupID: group.ID})
	if err != nil {
		t.Fatal(err)
	}
	want := []*pb.Submission{&submission2, &submission3}
	if diff := cmp.Diff(submissions, want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submissions, but got (-sub +want):\n%s", diff)
	}
	data, err := db.GetLastSubmissions(c1.ID, &pb.Submission{GroupID: group.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	// Since there is no submissions, but the course and user exist, an empty array should be returned
	data, err = db.GetLastSubmissions(c2.ID, &pb.Submission{GroupID: group.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 0, len(data))
	}
}

func TestDeleteGroup(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	course := &pb.Course{}
	qtest.CreateCourse(t, db, admin, course)

	var users []*pb.User
	enrollments := []pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db, uint64(i))
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
			query := &pb.Enrollment{
				UserID:   users[i].ID,
				CourseID: course.ID,
				Status:   pb.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
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
