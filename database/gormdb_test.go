package database_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
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
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	student := qtest.CreateFakeUser(t, db, 13)
	if err := db.CreateEnrollment(&qf.Enrollment{
		CourseID: course.ID,
		UserID:   student.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   student.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	// user entries from the database will have to be enrolled as
	// teacher and student respectively
	admin.Enrollments = append(admin.Enrollments, &qf.Enrollment{
		ID:           1,
		CourseID:     course.ID,
		UserID:       admin.ID,
		Status:       qf.Enrollment_TEACHER,
		State:        qf.Enrollment_VISIBLE,
		Course:       course,
		UsedSlipDays: []*qf.UsedSlipDays{},
	})
	admin.RemoteIdentities = nil

	student.Enrollments = append(student.Enrollments, &qf.Enrollment{
		ID:           2,
		CourseID:     course.ID,
		UserID:       student.ID,
		Status:       qf.Enrollment_STUDENT,
		State:        qf.Enrollment_VISIBLE,
		Course:       course,
		UsedSlipDays: []*qf.UsedSlipDays{},
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
		wantUser = &qf.User{
			ID:        uID,
			IsAdmin:   admin, // first user is always admin
			Name:      "Scrooge McDuck",
			StudentID: "22",
			Email:     "scrooge@mc.duck",
			AvatarURL: "https://github.com",
			RemoteIdentities: []*qf.RemoteIdentity{{
				ID:          rID,
				Provider:    provider,
				RemoteID:    remoteID,
				AccessToken: secret,
				UserID:      uID,
			}},
		}
		updates = &qf.User{
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

	var user qf.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&qf.RemoteIdentity{
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

func TestGormDBCreateEnrollmentNoRecord(t *testing.T) {
	const (
		userId   = 1
		courseId = 1
	)

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if err := db.CreateEnrollment(&qf.Enrollment{
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
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 10)
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Error(err)
	}

	if err := db.CreateEnrollment(&qf.Enrollment{
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
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 10)
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// Get course's pending enrollments.
	pendingEnrollments, err := db.GetEnrollmentsByCourse(course.ID, qf.Enrollment_PENDING)
	if err != nil {
		t.Fatal(err)
	}

	if len(pendingEnrollments) != 1 && pendingEnrollments[0].Status == qf.Enrollment_PENDING {
		t.Fatalf("have %v want 1 pending enrollment", pendingEnrollments)
	}

	// Accept enrollment.
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	// Get course's accepted enrollments.
	acceptedEnrollments, err := db.GetEnrollmentsByCourse(course.ID, qf.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}

	if len(acceptedEnrollments) != 1 && acceptedEnrollments[0].Status == qf.Enrollment_STUDENT {
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

func TestGormDBDuplicateIdentity(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if err := db.CreateUserFromRemoteIdentity(
		&qf.User{}, &qf.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	if err := db.CreateUserFromRemoteIdentity(
		&qf.User{}, &qf.RemoteIdentity{},
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
		wantUser1 = &qf.User{
			ID: uID,
			RemoteIdentities: []*qf.RemoteIdentity{{
				ID:          rID1,
				Provider:    provider1,
				RemoteID:    remoteID1,
				AccessToken: secret1,
				UserID:      uID,
			}},
		}

		wantUser2 = &qf.User{
			ID: uID,
			RemoteIdentities: []*qf.RemoteIdentity{
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
		&qf.User{},
		&qf.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	gotUser1 := &qf.User{}
	if err := db.CreateUserFromRemoteIdentity(
		gotUser1,
		&qf.RemoteIdentity{
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

	if err := db.UpdateUser(&qf.User{ID: id, IsAdmin: true}); err != gorm.ErrRecordNotFound {
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
		&qf.User{},
		&qf.RemoteIdentity{
			Provider: github,
		},
	); err != nil {
		t.Fatal(err)
	}

	var user qf.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&qf.RemoteIdentity{
			Provider: gitlab,
		},
	); err != nil {
		t.Fatal(err)
	}

	if user.IsAdmin {
		t.Error("user should not yet be an administrator")
	}

	if err := db.UpdateUser(&qf.User{ID: user.ID, IsAdmin: true}); err != nil {
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

func TestGormDBGetGroupSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if sub, err := db.GetLastSubmissions(10, &qf.Submission{GroupID: 10}); err != gorm.ErrRecordNotFound {
		t.Errorf("got submission %v", sub)
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBGetInsertGroupSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	c1 := &qf.Course{OrganizationID: 1, Code: "DAT101", Year: 1}
	c2 := &qf.Course{OrganizationID: 2, Code: "DAT101", Year: 2}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)

	var users []*qf.User
	enrollments := []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db, uint64(i))
		users = append(users, user)
	}
	// enroll users in course
	for i := 0; i < len(users); i++ {
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}
		if err := db.CreateEnrollment(&qf.Enrollment{
			CourseID: c1.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query := &qf.Enrollment{
				UserID:   users[i].ID,
				CourseID: c1.ID,
				Status:   qf.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	// Creating Group
	group := &qf.Group{
		Name:     "SameNameGroup",
		CourseID: c1.ID,
		Users:    users,
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	// Create Assignments
	assignment1 := qf.Assignment{
		Order:      1,
		CourseID:   c1.ID,
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := qf.Assignment{
		Order:      2,
		CourseID:   c1.ID,
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := qf.Assignment{
		Order:      1,
		CourseID:   c2.ID,
		IsGroupLab: false,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// Create some submissions
	submission1 := qf.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := qf.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := qf.Submission{
		GroupID:      group.ID,
		AssignmentID: assignment2.ID,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}
	submission4 := qf.Submission{
		UserID:       users[0].ID,
		AssignmentID: assignment3.ID,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission4); err != nil {
		t.Fatal(err)
	}

	// Even if there is three submission, only the latest for each assignment should be returned

	submissions, err := db.GetLastSubmissions(c1.ID, &qf.Submission{GroupID: group.ID})
	if err != nil {
		t.Fatal(err)
	}
	want := []*qf.Submission{&submission2, &submission3}
	if diff := cmp.Diff(submissions, want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submissions, but got (-sub +want):\n%s", diff)
	}
	data, err := db.GetLastSubmissions(c1.ID, &qf.Submission{GroupID: group.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	// Since there is no submissions, but the course and user exist, an empty array should be returned
	data, err = db.GetLastSubmissions(c2.ID, &qf.Submission{GroupID: group.ID})
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
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	var users []*qf.User
	enrollments := []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db, uint64(i))
		users = append(users, user)
	}
	// enroll users in course
	for i := 0; i < len(users); i++ {
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}
		if err := db.CreateEnrollment(&qf.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query := &qf.Enrollment{
				UserID:   users[i].ID,
				CourseID: course.ID,
				Status:   qf.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	group := &qf.Group{
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
