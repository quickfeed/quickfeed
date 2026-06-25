package database_test

// go test ./database/... -run TestBun

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestBunDBGetUser(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if _, err := db.GetUser(10); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("have error '%v' wanted '%v'", err, sql.ErrNoRows)
	}
}

func TestBunDBGetUsers(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if _, err := db.GetUsers(); err != nil {
		t.Errorf("have error '%v' wanted '%v'", err, nil)
	}
}

func TestBunDBGetUserWithEnrollments(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	student := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student, course)

	admin.Enrollments = append(admin.GetEnrollments(), &qf.Enrollment{
		ID:           1,
		CourseID:     course.GetID(),
		UserID:       admin.GetID(),
		Status:       qf.Enrollment_TEACHER,
		State:        qf.Enrollment_VISIBLE,
		Course:       course,
		UsedSlipDays: []*qf.UsedSlipDays{},
	})

	student.Enrollments = append(student.GetEnrollments(), &qf.Enrollment{
		ID:           2,
		CourseID:     course.GetID(),
		UserID:       student.GetID(),
		Status:       qf.Enrollment_STUDENT,
		State:        qf.Enrollment_VISIBLE,
		Course:       course,
		UsedSlipDays: []*qf.UsedSlipDays{},
	})

	gotTeacher, err := db.GetUserWithEnrollments(admin.GetID())
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(admin, gotTeacher, protocmp.Transform()); diff != "" {
		t.Errorf("enrollment mismatch (-teacher +gotTeacher):\n%s", diff)
	}
	gotStudent, err := db.GetUserWithEnrollments(student.GetID())
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(student, gotStudent, protocmp.Transform()); diff != "" {
		t.Errorf("enrollment mismatch (-student +gotStudent):\n%s", diff)
	}
}

func TestBunDBUpdateUser(t *testing.T) {
	const (
		userID   = 1
		secret   = "123"
		remoteID = 10
	)
	var (
		wantUser = &qf.User{
			ID:           userID,
			IsAdmin:      true,
			Name:         "Scrooge McDuck",
			StudentID:    "22",
			Email:        "scrooge@mc.duck",
			AvatarURL:    "https://github.com",
			ScmRemoteID:  remoteID,
			RefreshToken: secret,
		}
		updatedUser = &qf.User{
			ID:           userID,
			IsAdmin:      true,
			Name:         "Scrooge McDuck",
			StudentID:    "22",
			Email:        "scrooge@mc.duck",
			AvatarURL:    "https://github.com",
			ScmRemoteID:  remoteID,
			RefreshToken: secret,
		}
	)

	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	var user qf.User
	if err := db.CreateUser(&user); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateUser(updatedUser); err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(user.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}

	updatedUser.IsAdmin = false
	wantUser.IsAdmin = false
	if err := db.UpdateUser(updatedUser); err != nil {
		t.Fatal(err)
	}
	gotUser, err = db.GetUser(user.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantUser, +gotUser):\n%s", diff)
	}
}

func TestBunDBCreateEnrollmentNoRecord(t *testing.T) {
	const (
		userId   = 1
		courseId = 1
	)

	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   userId,
		CourseID: courseId,
	}); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected error '%v' have '%v'", sql.ErrNoRows, err)
	}
}

func TestBunDBCreateEnrollment(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db)
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: course.GetID(),
	}); err != nil {
		t.Error(err)
	}

	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: course.GetID(),
	}); err == nil {
		t.Fatal("expected duplicate enrollment creation to fail")
	}
}

func TestBunDBCreateEnrollmentIncompleteUser(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	tests := []struct {
		name        string
		user        *qf.User
		errContains string
	}{
		{
			name: "user without name",
			user: &qf.User{
				Email:       "test1@example.com",
				StudentID:   "12345",
				ScmRemoteID: 101,
			},
			errContains: "user must have name, email, and student ID set before enrolling",
		},
		{
			name: "user without email",
			user: &qf.User{
				Name:        "Test User",
				StudentID:   "12346",
				ScmRemoteID: 102,
			},
			errContains: "user must have name, email, and student ID set before enrolling",
		},
		{
			name: "user without student ID",
			user: &qf.User{
				Name:        "Test User",
				Email:       "test2@example.com",
				ScmRemoteID: 103,
			},
			errContains: "user must have name, email, and student ID set before enrolling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.CreateUser(tt.user); err != nil {
				t.Fatal(err)
			}
			err := db.CreateEnrollment(&qf.Enrollment{
				UserID:   tt.user.GetID(),
				CourseID: course.GetID(),
			})
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestBunDBAcceptRejectEnrollment(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db)
	query := &qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: course.GetID(),
	}
	if err := db.CreateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	pendingEnrollments, err := db.GetEnrollmentsByCourse(course.GetID(), qf.Enrollment_PENDING)
	if err != nil {
		t.Fatal(err)
	}

	if len(pendingEnrollments) != 1 && pendingEnrollments[0].GetStatus() == qf.Enrollment_PENDING {
		t.Fatalf("have %v want 1 pending enrollment", pendingEnrollments)
	}

	query.Status = qf.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	acceptedEnrollments, err := db.GetEnrollmentsByCourse(course.GetID(), qf.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}

	if len(acceptedEnrollments) != 1 && acceptedEnrollments[0].GetStatus() == qf.Enrollment_STUDENT {
		t.Fatalf("have %v want 1 accepted enrollment", acceptedEnrollments)
	}

	if err := db.RejectEnrollment(user.GetID(), course.GetID()); err != nil {
		t.Fatal(err)
	}

	allEnrollments, err := db.GetEnrollmentsByCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}

	for _, enrol := range allEnrollments {
		if enrol.GetUserID() == user.GetID() && enrol.GetCourseID() == course.GetID() {
			t.Fatalf("Enrollment %+v must have been deleted on rejection, but still found in the database", enrol)
		}
	}
}

func TestBunDBDuplicateIdentity(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if err := db.CreateUser(&qf.User{StudentID: "test1", ScmRemoteID: 1}); err != nil {
		t.Fatal(err)
	}
	// Bun correctly enforces UNIQUE constraints, so duplicate ScmRemoteID should fail
	if err := db.CreateUser(&qf.User{StudentID: "test2", ScmRemoteID: 1}); err == nil {
		t.Fatal("expected duplicate remote identity creation to fail")
	}
}

func TestBunDBSetAdminNoRecord(t *testing.T) {
	const id = 1

	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if err := db.UpdateUser(&qf.User{ID: id, IsAdmin: true}); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("have error '%v' wanted '%v'", err, sql.ErrNoRows)
	}
}

func TestBunDBSetAdmin(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	// Create first user with unique values to avoid UNIQUE constraint violations
	if err := db.CreateUser(&qf.User{StudentID: "admin1", ScmRemoteID: 100}); err != nil {
		t.Fatal(err)
	}

	user := qf.User{StudentID: "user2", ScmRemoteID: 200}
	if err := db.CreateUser(&user); err != nil {
		t.Fatal(err)
	}

	if user.GetIsAdmin() {
		t.Error("user should not yet be an administrator")
	}

	if err := db.UpdateUser(&qf.User{ID: user.GetID(), IsAdmin: true}); err != nil {
		t.Error(err)
	}

	admin, err := db.GetUser(user.GetID())
	if err != nil {
		t.Fatal(err)
	}

	if !admin.GetIsAdmin() {
		t.Error("user should be an administrator")
	}
}

func TestBunDBGetGroupSubmissions(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if sub, err := db.GetLastSubmissions(10, &qf.Submission{GroupID: 10}); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got submission %v", sub)
		t.Errorf("have error '%v' wanted '%v'", err, sql.ErrNoRows)
	}
}

func TestBunDBGetInsertGroupSubmissions(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	c1 := &qf.Course{ScmOrganizationID: 1, Code: "DAT101", Year: 1}
	c2 := &qf.Course{ScmOrganizationID: 2, Code: "DAT101", Year: 2}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)

	var users []*qf.User
	enrollments := []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT}
	for range enrollments {
		user := qtest.CreateFakeUser(t, db)
		users = append(users, user)
	}
	for i := 0; i < len(users); i++ {
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}
		query := &qf.Enrollment{
			CourseID: c1.GetID(),
			UserID:   users[i].GetID(),
		}
		if err := db.CreateEnrollment(query); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query.Status = qf.Enrollment_STUDENT
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	group := &qf.Group{
		Name:     "SameNameGroup",
		CourseID: c1.GetID(),
		Users:    users,
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	assignment1 := qf.Assignment{
		Order:      1,
		CourseID:   c1.GetID(),
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := qf.Assignment{
		Order:      2,
		CourseID:   c1.GetID(),
		IsGroupLab: true,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := qf.Assignment{
		Order:      1,
		CourseID:   c2.GetID(),
		IsGroupLab: false,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	submission1 := qf.Submission{
		GroupID:      group.GetID(),
		AssignmentID: assignment1.GetID(),
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := qf.Submission{
		GroupID:      group.GetID(),
		AssignmentID: assignment1.GetID(),
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := qf.Submission{
		GroupID:      group.GetID(),
		AssignmentID: assignment2.GetID(),
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}
	submission4 := qf.Submission{
		UserID:       users[0].GetID(),
		AssignmentID: assignment3.GetID(),
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission4); err != nil {
		t.Fatal(err)
	}

	submissions, err := db.GetLastSubmissions(c1.GetID(), &qf.Submission{GroupID: group.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	want := []*qf.Submission{&submission2, &submission3}
	if diff := cmp.Diff(submissions, want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submissions, but got (-sub +want):\n%s", diff)
	}
	data, err := db.GetLastSubmissions(c1.GetID(), &qf.Submission{GroupID: group.GetID()})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	data, err = db.GetLastSubmissions(c2.GetID(), &qf.Submission{GroupID: group.GetID()})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 0, len(data))
	}
}

func TestBunDeleteGroup(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	var users []*qf.User
	enrollments := []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT}
	for range enrollments {
		user := qtest.CreateFakeUser(t, db)
		users = append(users, user)
	}
	for i := 0; i < len(users); i++ {
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}
		query := &qf.Enrollment{
			CourseID: course.GetID(),
			UserID:   users[i].GetID(),
		}
		if err := db.CreateEnrollment(query); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query.Status = qf.Enrollment_STUDENT
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	group := &qf.Group{
		Name:     "SameNameGroup",
		CourseID: course.GetID(),
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

	gotModels, _ := db.GetGroup(group.GetID())
	if gotModels != nil {
		t.Errorf("Got %+v wanted None", gotModels)
	}
}
