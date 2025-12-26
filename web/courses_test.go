package web_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestUpdateCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	user := qtest.CreateFakeUser(t, db)
	dat520 := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, user, dat520)
	cookie := client.Cookie(t, user)

	// Update the course name
	wantCourse := proto.CloneOf(dat520)
	wantCourse.Name = "Updated Course Name"
	if _, err := client.UpdateCourse(context.Background(), qtest.RequestWithCookie(wantCourse, cookie)); err != nil {
		t.Error(err)
	}
	gotCourse := qtest.GetCourse(t, db, dat520.GetID())
	qtest.Diff(t, "UpdateCourse() mismatch", gotCourse, wantCourse, protocmp.Transform())
}

func TestUpdateCourseCommunityLink(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	user := qtest.CreateFakeUser(t, db)
	dat520 := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, user, dat520)
	cookie := client.Cookie(t, user)

	// Update the course with community link
	wantCourse := proto.CloneOf(dat520)
	wantCourse.CommunityLink = "https://discord.gg/example"
	wantCourse.CommunityLinkDescription = "Join our Discord server"
	if _, err := client.UpdateCourse(context.Background(), qtest.RequestWithCookie(wantCourse, cookie)); err != nil {
		t.Error(err)
	}
	gotCourse := qtest.GetCourse(t, db, dat520.GetID())
	qtest.Diff(t, "UpdateCourse() community link mismatch", gotCourse, wantCourse, protocmp.Transform())
}

func TestGetCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	cookie := client.Cookie(t, admin)

	wantCourse := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, wantCourse)

	gotCourse, err := client.GetCourse(context.Background(), qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: wantCourse.GetID(),
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
	}
}

func TestGetCourseWithoutDockerfileDigest(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	cookie := client.Cookie(t, admin)

	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	resp, err := client.GetCourse(context.Background(), qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: course.GetID(),
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	course = resp.Msg
	if course.GetDockerfileDigest() != "" {
		t.Errorf("expected empty DockerfileDigest, got %s", course.GetDockerfileDigest())
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
		CourseID: course.GetID(),
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	// Check that the digest is not returned in the response.
	course = resp.Msg
	if course.GetDockerfileDigest() != "" {
		t.Errorf("expected DockerfileDigest to be removed, got %s", course.GetDockerfileDigest())
	}
}

func TestGetCourses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	cookie := client.Cookie(t, admin)

	for _, wantCourse := range qtest.MockCourses {
		qtest.CreateCourse(t, db, admin, wantCourse)
	}

	wantCourses := qtest.MockCourses
	foundCourses, err := client.GetCourses(context.Background(), qtest.RequestWithCookie(&qf.Void{}, cookie))
	if err != nil {
		t.Error(err)
	}
	gotCourses := foundCourses.Msg.GetCourses()
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses +gotCourses):\n%s", diff)
	}
}

func TestEnrollmentProcess(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin", ScmRemoteID: 1})
	client := web.NewMockClient(t, db, scm.WithMockOptions(scm.WithMockCourses(), scm.WithMockOrgs("admin", "student1", "student2")), web.WithInterceptors())

	ctx := context.Background()
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	stud1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student1", Login: "student1", ScmRemoteID: 2})
	enrollStud1 := &qf.Enrollment{CourseID: course.GetID(), UserID: stud1.GetID()}
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(enrollStud1, client.Cookie(t, stud1))); err != nil {
		t.Error(err)
	}

	// verify that a pending enrollment was indeed created for the user
	enrollStatusReq := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_UserID{
			UserID: stud1.GetID(),
		},
		Statuses: []qf.Enrollment_UserStatus{
			qf.Enrollment_PENDING,
		},
	}
	userEnrollments, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(enrollStatusReq, client.Cookie(t, stud1)))
	if err != nil {
		t.Fatal(err)
	}
	var pendingUserEnrollment *qf.Enrollment
	for _, enrollment := range userEnrollments.Msg.GetEnrollments() {
		if enrollment.GetCourseID() == course.GetID() {
			if enrollment.GetStatus() == qf.Enrollment_PENDING {
				pendingUserEnrollment = enrollment
			} else {
				t.Errorf("expected student %d to have pending enrollment in course %d", stud1.GetID(), course.GetID())
			}
		}
	}

	// verify that a pending enrollment was indeed created for the course.
	enrollReq := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_CourseID{
			CourseID: course.GetID(),
		},
	}

	// enrollments fetched with FetchMode CourseID will not have the course field preloaded.
	courseEnrollments, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(enrollReq, client.Cookie(t, admin)))
	if err != nil {
		t.Error(err)
	}
	var pendingCourseEnrollment *qf.Enrollment
	for _, enrollment := range courseEnrollments.Msg.GetEnrollments() {
		if enrollment.GetUserID() == stud1.GetID() {
			if enrollment.GetStatus() == qf.Enrollment_PENDING {
				pendingCourseEnrollment = enrollment
			} else {
				t.Errorf("expected student %d to have pending enrollment in course %d", stud1.GetID(), course.GetID())
			}
		}
	}
	if diff := cmp.Diff(pendingUserEnrollment, pendingCourseEnrollment,
		qtest.UserDiffOptions(),
		protocmp.IgnoreFields(&qf.Enrollment{}, "course"),
	); diff != "" {
		t.Errorf("%v, %v", pendingUserEnrollment, pendingCourseEnrollment)
		t.Errorf("EnrollmentProcess mismatch (-pendingUserEnrollment +pendingCourseEnrollment):\n%s", diff)
	}

	wantEnrollment := &qf.Enrollment{
		ID:           pendingCourseEnrollment.GetID(),
		CourseID:     course.GetID(),
		UserID:       stud1.GetID(),
		Status:       qf.Enrollment_PENDING,
		State:        qf.Enrollment_VISIBLE,
		Course:       course,
		User:         stud1,
		UsedSlipDays: []*qf.UsedSlipDays{},
	}
	if diff := cmp.Diff(wantEnrollment, pendingCourseEnrollment, cmp.Options{
		qtest.UserDiffOptions(),
		protocmp.IgnoreFields(&qf.Enrollment{}, "course"),
	}); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +pendingEnrollment):\n%s", diff)
	}

	enrollStud1.Status = qf.Enrollment_STUDENT
	enrollStud1.Course = course
	if _, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{enrollStud1},
	}, client.Cookie(t, admin))); err != nil {
		t.Error(err)
	}

	// verify that the enrollment was updated to student status.
	gotEnrollment, err := db.GetEnrollmentByCourseAndUser(course.GetID(), stud1.GetID())
	if err != nil {
		t.Error(err)
	}
	wantEnrollment.Status = qf.Enrollment_STUDENT
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, qtest.UserDiffOptions()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// create another user and enroll as student

	stud2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student2", Login: "student2", ScmRemoteID: 3})
	enrollStud2 := &qf.Enrollment{CourseID: course.GetID(), UserID: stud2.GetID()}
	if _, err = client.CreateEnrollment(ctx, qtest.RequestWithCookie(enrollStud2, client.Cookie(t, stud2))); err != nil {
		t.Error(err)
	}
	enrollStud2.Status = qf.Enrollment_STUDENT
	if _, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			enrollStud2,
		},
	}, client.Cookie(t, admin))); err != nil {
		t.Error(err)
	}
	// verify that the stud2 was enrolled with student status.
	gotEnrollment, err = db.GetEnrollmentByCourseAndUser(course.GetID(), stud2.GetID())
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = gotEnrollment.GetID()
	wantEnrollment.Status = qf.Enrollment_STUDENT
	wantEnrollment.UserID = stud2.GetID()
	wantEnrollment.User = stud2
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, qtest.UserDiffOptions()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// promote stud2 to teaching assistant

	enrollStud2.Status = qf.Enrollment_TEACHER
	if _, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			enrollStud2,
		},
	}, client.Cookie(t, admin))); err != nil {
		t.Error(err)
	}
	// verify that the stud2 was promoted to teacher status.
	gotEnrollment, err = db.GetEnrollmentByCourseAndUser(course.GetID(), stud2.GetID())
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = gotEnrollment.GetID()
	wantEnrollment.Status = qf.Enrollment_TEACHER
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, qtest.UserDiffOptions()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}
}

func TestListCoursesWithEnrollment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())

	admin := qtest.CreateFakeUser(t, db)
	user := qtest.CreateFakeUser(t, db)

	var testCourses []*qf.Course
	for _, course := range qtest.MockCourses {
		qtest.CreateCourse(t, db, admin, course)
		testCourses = append(testCourses, course)
	}

	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: testCourses[0].GetID(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: testCourses[1].GetID(),
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: testCourses[2].GetID(),
	}
	if err := db.CreateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.GetID(), testCourses[1].GetID()); err != nil {
		t.Fatal(err)
	}
	query.Status = qf.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	gotUser, err := client.GetUser(context.Background(), qtest.RequestWithCookie(&qf.Void{}, client.Cookie(t, user)))
	if err != nil {
		t.Error(err)
	}

	wantCourses := map[uint64]qf.Enrollment_UserStatus{
		testCourses[0].GetID(): qf.Enrollment_PENDING,
		testCourses[1].GetID(): qf.Enrollment_NONE,
		testCourses[2].GetID(): qf.Enrollment_STUDENT,
		testCourses[3].GetID(): qf.Enrollment_NONE,
	}
	for _, enrollment := range gotUser.Msg.GetEnrollments() {
		course := enrollment.GetCourse()
		wantStatus, ok := wantCourses[course.GetID()]
		if !ok {
			t.Errorf("unexpected course: %+v", course.GetID())
		}
		if enrollment.GetStatus() != wantStatus {
			t.Errorf("have course %+v want %+v", enrollment.GetStatus(), wantStatus)
		}
	}
}

func TestListCoursesWithEnrollmentStatuses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())

	admin := qtest.CreateFakeUser(t, db)
	var testCourses []*qf.Course
	for _, course := range qtest.MockCourses {
		qtest.CreateCourse(t, db, admin, course)
		testCourses = append(testCourses, course)
	}

	user := qtest.CreateFakeUser(t, db)

	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: testCourses[0].GetID(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: testCourses[1].GetID(),
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: testCourses[2].GetID(),
	}
	if err := db.CreateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	// user enrollment is rejected for course 1 and enrolled for course 2, still pending for course 0
	if err := db.RejectEnrollment(user.GetID(), testCourses[1].GetID()); err != nil {
		t.Fatal(err)
	}
	query.Status = qf.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	gotUser, err := client.GetUser(context.Background(), qtest.RequestWithCookie(&qf.Void{}, client.Cookie(t, user)))
	if err != nil {
		t.Error(err)
	}
	gotCourses := make([]*qf.Course, 0)
	for _, enrollment := range gotUser.Msg.GetEnrollments() {
		// since GetUser returns all enrollments, we only keep the student enrollments
		if enrollment.GetStatus() == qf.Enrollment_STUDENT {
			course := enrollment.GetCourse()
			course.Enrolled = enrollment.GetStatus()
			gotCourses = append(gotCourses, course)
		}
	}

	stats := make([]qf.Enrollment_UserStatus, 0)
	stats = append(stats, qf.Enrollment_STUDENT)
	course_req := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_UserID{
			UserID: user.GetID(),
		},
		Statuses: stats,
	}
	enrollments, err := client.GetEnrollments(context.Background(), qtest.RequestWithCookie(course_req, client.Cookie(t, user)))
	if err != nil {
		t.Error(err)
	}
	gotCourses2 := make([]*qf.Course, 0)
	for _, enrollment := range enrollments.Msg.GetEnrollments() {
		// since GetEnrollmentsByUser returns only student enrollments
		course := enrollment.GetCourse()
		course.Enrolled = enrollment.GetStatus()
		gotCourses2 = append(gotCourses2, course)
	}

	wantCourses, err := db.GetCoursesByUser(user.GetID(), qf.Enrollment_STUDENT)
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

	client := web.NewMockClient(t, db, scm.WithMockOptions(scm.WithMockCourses(), scm.WithMockOrgs("teacher", "student1", "student2", "TA")), web.WithInterceptors())

	teacher := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "teacher", Login: "teacher", ScmRemoteID: 1})
	student1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student1", Login: "student1", ScmRemoteID: 2})
	student2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student2", Login: "student2", ScmRemoteID: 3})
	ta := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "TA", Login: "TA", ScmRemoteID: 4})

	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)

	student1Enrollment := &qf.Enrollment{UserID: student1.GetID(), CourseID: course.GetID()}
	student2Enrollment := &qf.Enrollment{UserID: student2.GetID(), CourseID: course.GetID()}
	taEnrollment := &qf.Enrollment{UserID: ta.GetID(), CourseID: course.GetID()}
	teacherEnrollment := &qf.Enrollment{UserID: teacher.GetID(), CourseID: course.GetID()}

	ctx := context.Background()

	// student1 attempts to enroll in the course, must succeed
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(student1Enrollment, client.Cookie(t, student1))); err != nil {
		t.Error(err)
	}
	// student2 attempts to enroll in the course, must succeed
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(student2Enrollment, client.Cookie(t, student2))); err != nil {
		t.Error(err)
	}
	// ta attempts to enroll in the course, must succeed
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(taEnrollment, client.Cookie(t, ta))); err != nil {
		t.Error(err)
	}

	request := &qf.Enrollments{}

	// teacher accepts pending students {student1, student2, ta} as student, must succeed
	request.Enrollments = []*qf.Enrollment{student1Enrollment, student2Enrollment, taEnrollment}
	for _, enrollment := range request.GetEnrollments() {
		enrollment.Status = qf.Enrollment_STUDENT
	}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, teacher))); err != nil {
		t.Error(err)
	}

	// teacher promotes students to teachers, must succeed
	for _, enrollment := range request.GetEnrollments() {
		enrollment.Status = qf.Enrollment_TEACHER
	}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, teacher))); err != nil {
		t.Error(err)
	}

	// TA attempts to demote self, must succeed
	taEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{taEnrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, ta))); err != nil {
		t.Error(err)
	}

	// student2 attempts to demote course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{teacherEnrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, student2))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// student2 attempts to reject course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_NONE
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, student2))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// teacher demotes student1, must succeed
	student1Enrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{student1Enrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, teacher))); err != nil {
		t.Error(err)
	}

	// check that student1 is now enrolled as student
	enrol, err := db.GetEnrollmentByCourseAndUser(course.GetID(), student1.GetID())
	if err != nil {
		t.Error(err)
	}
	if enrol.GetStatus() != qf.Enrollment_STUDENT {
		t.Errorf("expected status %s, got %s", qf.Enrollment_STUDENT, enrol.GetStatus())
	}

	// teacher rejects student2, must succeed
	student2Enrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{student2Enrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, teacher))); err != nil {
		t.Error(err)
	}
	student2Enrollment.Status = qf.Enrollment_NONE
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, teacher))); err != nil {
		t.Error(err)
	}

	// ensure that student2 is no longer enrolled in the course
	if _, err := db.GetEnrollmentByCourseAndUser(course.GetID(), student2.GetID()); err == nil {
		t.Error("expected error 'record not found'")
	}

	// course creator attempts to demote himself, must fail as well
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{teacherEnrollment}
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, teacher))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// same when rejecting
	teacherEnrollment.Status = qf.Enrollment_NONE
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, teacher))); err == nil {
		t.Errorf("expected error: 'permission_denied: course creator cannot be demoted', got: '%v'", err)
	}

	// ta attempts to demote course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(request, client.Cookie(t, ta))); err == nil {
		t.Errorf("expected error 'permission_denied: access denied for UpdateEnrollments: required roles [4] not satisfied by claims', got: '%v'", err)
	}
}

func TestUpdateCourseVisibility(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockCourses(), web.WithInterceptors())

	teacher := qtest.CreateFakeUser(t, db)
	user := qtest.CreateFakeUser(t, db)
	cookie := client.Cookie(t, user)

	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: course.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	req := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_UserID{
			UserID: user.GetID(),
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
	enrollment := enrollments.Msg.GetEnrollments()[0]
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

	gotEnrollment := gotEnrollments.Msg.GetEnrollments()[0]
	if gotEnrollment.GetState() != qf.Enrollment_FAVORITE {
		// State should have changed to favorite
		t.Errorf("expected enrollment state %s, got %s", qf.Enrollment_FAVORITE, gotEnrollment.GetState())
	}
	if gotEnrollment.GetStatus() != qf.Enrollment_PENDING {
		// Status should *not* have changed
		t.Errorf("expected enrollment status %s, got %s", qf.Enrollment_NONE, gotEnrollment.GetStatus())
	}
}
