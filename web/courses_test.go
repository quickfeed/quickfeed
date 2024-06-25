package web_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateAndGetCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	cookie := Cookie(t, tm, admin)

	wantCourse := qtest.MockCourses[0]
	createdCourse, err := client.CreateCourse(context.Background(), qtest.RequestWithCookie(wantCourse, cookie))
	if err != nil {
		t.Fatal(err)
	}

	gotCourse, err := client.GetCourse(context.Background(), qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: createdCourse.Msg.ID,
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	wantCourse.ID = createdCourse.Msg.ID
	if diff := cmp.Diff(wantCourse, gotCourse.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
	}
	if diff := cmp.Diff(createdCourse.Msg, gotCourse.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-createdCourse +gotCourse):\n%s", diff)
	}
}

func TestGetCourseWithoutDockerfileDigest(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	cookie := Cookie(t, tm, admin)

	wantCourse := qtest.MockCourses[0]
	createdCourse, err := client.CreateCourse(context.Background(), qtest.RequestWithCookie(wantCourse, cookie))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.GetCourse(context.Background(), qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: createdCourse.Msg.ID,
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	course := resp.Msg
	if course.DockerfileDigest != "" {
		t.Errorf("expected empty DockerfileDigest, got %s", course.DockerfileDigest)
	}
	dockerfile := "FROM golang:latest"
	want := true
	got := course.UpdateDockerfile(dockerfile)
	if got != want {
		t.Errorf("UpdateDockerfile(%q) = %t, want %t", dockerfile, got, want)
	}
	// Update the course's DockerfileDigest in the database
	// To simulate the behavior in assignments.UpdateFromTestsRepo()
	if err := db.UpdateCourse(course); err != nil {
		t.Error(err)
	}

	// GetCourse again to check that the digest is not returned in the response.
	resp, err = client.GetCourse(context.Background(), qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: createdCourse.Msg.ID,
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	// Check that the digest is not returned in the response.
	course = resp.Msg
	if course.DockerfileDigest != "" {
		t.Errorf("expected DockerfileDigest to be removed, got %s", course.DockerfileDigest)
	}
}

func TestCreateAndGetCourses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	cookie := Cookie(t, tm, admin)

	for _, wantCourse := range qtest.MockCourses {
		gotCourse, err := client.CreateCourse(context.Background(), qtest.RequestWithCookie(wantCourse, cookie))
		if err != nil {
			t.Error(err)
		}
		// copy the ID from the created course to the expected course
		wantCourse.ID = gotCourse.Msg.ID
		if diff := cmp.Diff(wantCourse, gotCourse.Msg, protocmp.Transform()); diff != "" {
			t.Errorf("CreateCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
		}
	}

	wantCourses := qtest.MockCourses
	foundCourses, err := client.GetCourses(context.Background(), qtest.RequestWithCookie(&qf.Void{}, cookie))
	if err != nil {
		t.Error(err)
	}
	gotCourses := foundCourses.Msg.Courses
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses +gotCourses):\n%s", diff)
	}
}

func TestNewCourseExistingRepos(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockCourses())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	cookie := Cookie(t, tm, admin)

	ctx := context.Background()
	course, err := client.CreateCourse(ctx, qtest.RequestWithCookie(qtest.MockCourses[0], cookie))
	if course != nil {
		t.Fatal("expected CreateCourse to fail with AlreadyExists")
	}
	if err != nil && connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("expected CreateCourse to fail with AlreadyExists, but got: %v", err)
	}
}

func TestEnrollmentProcess(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	client, tm := MockClientWithOption(t, db, scm.WithMockOrgs())

	ctx := context.Background()
	course, err := client.CreateCourse(ctx, qtest.RequestWithCookie(qtest.MockCourses[0], Cookie(t, tm, admin)))
	if err != nil {
		t.Fatal(err)
	}

	stud1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student1", Login: "student1"})
	enrollStud1 := &qf.Enrollment{CourseID: course.Msg.ID, UserID: stud1.ID}
	if _, err = client.CreateEnrollment(ctx, qtest.RequestWithCookie(enrollStud1, Cookie(t, tm, stud1))); err != nil {
		t.Error(err)
	}

	// verify that a pending enrollment was indeed created for the user
	enrollStatusReq := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_UserID{
			UserID: stud1.ID,
		},
		Statuses: []qf.Enrollment_UserStatus{
			qf.Enrollment_PENDING,
		},
	}
	userEnrollments, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(enrollStatusReq, Cookie(t, tm, stud1)))
	if err != nil {
		t.Fatal(err)
	}
	var pendingUserEnrollment *qf.Enrollment
	for _, enrollment := range userEnrollments.Msg.Enrollments {
		if enrollment.CourseID == course.Msg.ID {
			if enrollment.Status == qf.Enrollment_PENDING {
				pendingUserEnrollment = enrollment
			} else {
				t.Errorf("expected student %d to have pending enrollment in course %d", stud1.ID, course.Msg.ID)
			}
		}
	}

	// verify that a pending enrollment was indeed created for the course.
	enrollReq := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_CourseID{
			CourseID: course.Msg.ID,
		},
	}
	courseEnrollments, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(enrollReq, Cookie(t, tm, admin)))
	if err != nil {
		t.Error(err)
	}
	var pendingCourseEnrollment *qf.Enrollment
	for _, enrollment := range courseEnrollments.Msg.Enrollments {
		if enrollment.UserID == stud1.ID {
			if enrollment.Status == qf.Enrollment_PENDING {
				pendingCourseEnrollment = enrollment
			} else {
				t.Errorf("expected student %d to have pending enrollment in course %d", stud1.ID, course.Msg.ID)
			}
		}
	}
	if diff := cmp.Diff(pendingUserEnrollment, pendingCourseEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-pendingUserEnrollment +pendingCourseEnrollment):\n%s", diff)
	}

	wantEnrollment := &qf.Enrollment{
		ID:           pendingCourseEnrollment.ID,
		CourseID:     course.Msg.ID,
		UserID:       stud1.ID,
		Status:       qf.Enrollment_PENDING,
		State:        qf.Enrollment_VISIBLE,
		Course:       course.Msg,
		User:         stud1,
		UsedSlipDays: []*qf.UsedSlipDays{},
	}
	if diff := cmp.Diff(wantEnrollment, pendingCourseEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +pendingEnrollment):\n%s", diff)
	}

	enrollStud1.Status = qf.Enrollment_STUDENT
	enrollStud1.Course = course.Msg
	if _, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{enrollStud1},
	}, Cookie(t, tm, admin))); err != nil {
		t.Error(err)
	}

	// verify that the enrollment was updated to student status.
	gotEnrollment, err := db.GetEnrollmentByCourseAndUser(course.Msg.ID, stud1.ID)
	if err != nil {
		t.Error(err)
	}
	wantEnrollment.Status = qf.Enrollment_STUDENT
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// create another user and enroll as student

	stud2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student2", Login: "student2"})
	enrollStud2 := &qf.Enrollment{CourseID: course.Msg.ID, UserID: stud2.ID}
	if _, err = client.CreateEnrollment(ctx, qtest.RequestWithCookie(enrollStud2, Cookie(t, tm, stud2))); err != nil {
		t.Error(err)
	}
	enrollStud2.Status = qf.Enrollment_STUDENT
	if _, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			enrollStud2,
		},
	}, Cookie(t, tm, admin))); err != nil {
		t.Error(err)
	}
	// verify that the stud2 was enrolled with student status.
	gotEnrollment, err = db.GetEnrollmentByCourseAndUser(course.Msg.ID, stud2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = gotEnrollment.ID
	wantEnrollment.Status = qf.Enrollment_STUDENT
	wantEnrollment.UserID = stud2.ID
	wantEnrollment.User = stud2
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// promote stud2 to teaching assistant

	enrollStud2.Status = qf.Enrollment_TEACHER
	if _, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			enrollStud2,
		},
	}, Cookie(t, tm, admin))); err != nil {
		t.Error(err)
	}
	// verify that the stud2 was promoted to teacher status.
	gotEnrollment, err = db.GetEnrollmentByCourseAndUser(course.Msg.ID, stud2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = gotEnrollment.ID
	wantEnrollment.Status = qf.Enrollment_TEACHER
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}
}

func TestListCoursesWithEnrollment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	user := qtest.CreateFakeUser(t, db)

	var testCourses []*qf.Course
	for _, course := range qtest.MockCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, course)
	}

	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}
	if err := db.CreateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	query.Status = qf.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	gotUser, err := client.GetUser(context.Background(), qtest.RequestWithCookie(&qf.Void{}, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}

	wantCourses := map[uint64]qf.Enrollment_UserStatus{
		testCourses[0].ID: qf.Enrollment_PENDING,
		testCourses[1].ID: qf.Enrollment_NONE,
		testCourses[2].ID: qf.Enrollment_STUDENT,
		testCourses[3].ID: qf.Enrollment_NONE,
	}
	for _, enrollment := range gotUser.Msg.GetEnrollments() {
		course := enrollment.Course
		wantStatus, ok := wantCourses[course.ID]
		if !ok {
			t.Errorf("unexpected course: %+v", course.ID)
		}
		if enrollment.Status != wantStatus {
			t.Errorf("have course %+v want %+v", enrollment.Status, wantStatus)
		}
	}
}

func TestListCoursesWithEnrollmentStatuses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	var testCourses []*qf.Course
	for _, course := range qtest.MockCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, course)
	}

	user := qtest.CreateFakeUser(t, db)

	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}
	if err := db.CreateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	// user enrollment is rejected for course 1 and enrolled for course 2, still pending for course 0
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	query.Status = qf.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	gotUser, err := client.GetUser(context.Background(), qtest.RequestWithCookie(&qf.Void{}, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}
	gotCourses := make([]*qf.Course, 0)
	for _, enrollment := range gotUser.Msg.GetEnrollments() {
		// since GetUser returns all enrollments, we only keep the student enrollments
		if enrollment.Status == qf.Enrollment_STUDENT {
			course := enrollment.Course
			course.Enrolled = enrollment.Status
			gotCourses = append(gotCourses, course)
		}
	}

	stats := make([]qf.Enrollment_UserStatus, 0)
	stats = append(stats, qf.Enrollment_STUDENT)
	course_req := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_UserID{
			UserID: user.ID,
		},
		Statuses: stats,
	}
	enrollments, err := client.GetEnrollments(context.Background(), qtest.RequestWithCookie(course_req, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}
	gotCourses2 := make([]*qf.Course, 0)
	for _, enrollment := range enrollments.Msg.GetEnrollments() {
		// since GetEnrollmentsByUser returns only student enrollments
		course := enrollment.Course
		course.Enrolled = enrollment.Status
		gotCourses2 = append(gotCourses2, course)
	}

	wantCourses, err := db.GetCoursesByUser(user.ID, qf.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetUser() mismatch (-wantCourses +gotCourses):\n%s", diff)
	}
	if diff := cmp.Diff(wantCourses, gotCourses2, protocmp.Transform()); diff != "" {
		t.Errorf("GetEnrollmentsByUser() mismatch (-wantCourses +gotCourses2):\n%s", diff)
	}
}

func TestPromoteDemoteRejectTeacher(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockCourses())

	teacher := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "teacher", Login: "teacher"})
	student1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student1", Login: "student1"})
	student2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student2", Login: "student2"})
	ta := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "TA", Login: "TA"})

	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)

	student1Enrollment := &qf.Enrollment{UserID: student1.ID, CourseID: course.ID}
	student2Enrollment := &qf.Enrollment{UserID: student2.ID, CourseID: course.ID}
	taEnrollment := &qf.Enrollment{UserID: ta.ID, CourseID: course.ID}
	teacherEnrollment := &qf.Enrollment{UserID: teacher.ID, CourseID: course.ID}

	ctx := context.Background()

	// student1 attempts to enroll in the course, must succeed
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(student1Enrollment, Cookie(t, tm, student1))); err != nil {
		t.Error(err)
	}
	// student2 attempts to enroll in the course, must succeed
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(student2Enrollment, Cookie(t, tm, student2))); err != nil {
		t.Error(err)
	}
	// ta attempts to enroll in the course, must succeed
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(taEnrollment, Cookie(t, tm, ta))); err != nil {
		t.Error(err)
	}

	request := &qf.Enrollments{}

	// teacher accepts pending students {student1, student2, ta} as student, must succeed
	request.Enrollments = []*qf.Enrollment{student1Enrollment, student2Enrollment, taEnrollment}
	for _, enrollment := range request.Enrollments {
		enrollment.Status = qf.Enrollment_STUDENT
	}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, teacher))); err != nil {
		t.Error(err)
	}

	// teacher promotes students to teachers, must succeed
	for _, enrollment := range request.Enrollments {
		enrollment.Status = qf.Enrollment_TEACHER
	}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, teacher))); err != nil {
		t.Error(err)
	}

	// TA attempts to demote self, must succeed
	taEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{taEnrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, ta))); err != nil {
		t.Error(err)
	}

	// student2 attempts to demote course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{teacherEnrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, student2))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// student2 attempts to reject course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_NONE
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, student2))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// teacher demotes student1, must succeed
	student1Enrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{student1Enrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, teacher))); err != nil {
		t.Error(err)
	}

	// check that student1 is now enrolled as student
	enrol, err := db.GetEnrollmentByCourseAndUser(course.ID, student1.ID)
	if err != nil {
		t.Error(err)
	}
	if enrol.Status != qf.Enrollment_STUDENT {
		t.Errorf("expected status %s, got %s", qf.Enrollment_STUDENT, enrol.Status)
	}

	// teacher rejects student2, must succeed
	student2Enrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{student2Enrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, teacher))); err != nil {
		t.Error(err)
	}
	student2Enrollment.Status = qf.Enrollment_NONE
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, teacher))); err != nil {
		t.Error(err)
	}

	// ensure that student2 is no longer enrolled in the course
	if _, err := db.GetEnrollmentByCourseAndUser(course.ID, student2.ID); err == nil {
		t.Error("expected error 'record not found'")
	}

	// course creator attempts to demote himself, must fail as well
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{teacherEnrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, teacher))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// same when rejecting
	teacherEnrollment.Status = qf.Enrollment_NONE
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, teacher))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// ta attempts to demote course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, Cookie(t, tm, ta))); err == nil {
		t.Errorf("expected error 'permission_denied: access denied for UpdateEnrollments: required roles [4] not satisfied by claims', got: '%v'", err)
	}
}

func TestUpdateCourseVisibility(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := MockClientWithOption(t, db, scm.WithMockCourses())

	teacher := qtest.CreateFakeUser(t, db)
	user := qtest.CreateFakeUser(t, db)
	cookie := Cookie(t, tm, user)

	course := qtest.MockCourses[0]
	if err := db.CreateCourse(teacher.ID, course); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	req := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_UserID{
			UserID: user.ID,
		},
	}
	enrollments, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(req, cookie))
	if err != nil {
		t.Error(err)
	}
	if len(enrollments.Msg.GetEnrollments()) != 1 {
		t.Errorf("expected 1 enrollment, got %d", len(enrollments.Msg.GetEnrollments()))
	}

	// pending enrollment should be allowed to change visibility, but not status
	enrollment := enrollments.Msg.Enrollments[0]
	enrollment.State = qf.Enrollment_FAVORITE
	enrollment.Status = qf.Enrollment_TEACHER
	if _, err := client.UpdateCourseVisibility(ctx, qtest.RequestWithCookie(enrollment, cookie)); err != nil {
		t.Error(err)
	}

	gotEnrollments, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(req, cookie))
	if err != nil {
		t.Error(err)
	}
	if len(gotEnrollments.Msg.GetEnrollments()) != 1 {
		t.Errorf("expected 1 enrollment, got %d", len(gotEnrollments.Msg.GetEnrollments()))
	}

	gotEnrollment := gotEnrollments.Msg.Enrollments[0]
	if gotEnrollment.State != qf.Enrollment_FAVORITE {
		// State should have changed to favorite
		t.Errorf("expected enrollment state %s, got %s", qf.Enrollment_FAVORITE, gotEnrollment.State)
	}
	if gotEnrollment.Status != qf.Enrollment_PENDING {
		// Status should *not* have changed
		t.Errorf("expected enrollment status %s, got %s", qf.Enrollment_NONE, gotEnrollment.Status)
	}
}
