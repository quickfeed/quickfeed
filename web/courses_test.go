package web_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/quickfeed/quickfeed/ag"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

var allCourses = []*pb.Course{
	{
		Name:            "Distributed Systems",
		CourseCreatorID: 1,
		Code:            "DAT520",
		Year:            2018,
		Tag:             "Spring",
		Provider:        "fake",
		OrganizationID:  1,
	},
	{
		Name:            "Operating Systems",
		CourseCreatorID: 1,
		Code:            "DAT320",
		Year:            2017,
		Tag:             "Fall",
		Provider:        "fake",
		OrganizationID:  2,
	},
	{
		Name:            "New Systems",
		CourseCreatorID: 1,
		Code:            "DATx20",
		Year:            2019,
		Tag:             "Fall",
		Provider:        "fake",
		OrganizationID:  3,
	},
	{
		Name:            "Hyped Systems",
		CourseCreatorID: 1,
		Code:            "DATx20",
		Year:            2020,
		Tag:             "Fall",
		Provider:        "fake",
		OrganizationID:  4,
	},
}

func TestGetCourses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	_, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	var wantCourses []*pb.Course
	for _, course := range allCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
		wantCourses = append(wantCourses, course)
	}

	foundCourses, err := ags.GetCourses(context.Background(), &pb.Void{})
	if err != nil {
		t.Fatal(err)
	}
	gotCourses := foundCourses.Courses
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetCourses() mismatch (-wantCourses +gotCourses):\n%s", diff)
	}
}

func TestNewCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	ctx := qtest.WithUserContext(context.Background(), admin)
	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	for _, wantCourse := range allCourses {
		// each course needs a separate directory
		_, err := fakeProvider.CreateOrganization(ctx, &scm.OrganizationOptions{Path: "path", Name: "name"})
		if err != nil {
			t.Fatal(err)
		}

		gotCourse, err := ags.CreateCourse(ctx, wantCourse)
		if err != nil {
			t.Fatal(err)
		}
		wantCourse.ID = gotCourse.ID
		if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
			t.Errorf("ags.CreateCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
		}

		// check that the database also has the course
		gotCourse, err = db.GetCourse(wantCourse.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
			t.Errorf("db.GetCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
		}
	}
}

func TestNewCourseExistingRepos(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	ctx := qtest.WithUserContext(context.Background(), admin)
	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	directory, _ := fakeProvider.CreateOrganization(ctx, &scm.OrganizationOptions{Path: "path", Name: "name"})
	for path, private := range web.RepoPaths {
		repoOptions := &scm.CreateRepositoryOptions{Path: path, Organization: directory, Private: private}
		_, err := fakeProvider.CreateRepository(ctx, repoOptions)
		if err != nil {
			t.Fatal(err)
		}
	}

	course, err := ags.CreateCourse(ctx, allCourses[0])
	if course != nil {
		t.Fatal("expected CreateCourse to fail with AlreadyExists")
	}
	if err != nil && status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected CreateCourse to fail with AlreadyExists, but got: %v", err)
	}
}

func TestEnrollmentProcess(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	ctx := qtest.WithUserContext(context.Background(), admin)
	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	_, err := fakeProvider.CreateOrganization(ctx, &scm.OrganizationOptions{Path: "path", Name: "name"})
	if err != nil {
		t.Fatal(err)
	}

	course, err := ags.CreateCourse(ctx, allCourses[0])
	if err != nil {
		t.Fatal(err)
	}

	stud1 := qtest.CreateFakeUser(t, db, 2)
	enrollStud1 := &pb.Enrollment{CourseID: course.ID, UserID: stud1.ID}
	if _, err = ags.CreateEnrollment(ctx, enrollStud1); err != nil {
		t.Fatal(err)
	}

	// verify that a pending enrollment was indeed created.
	pendingEnrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, stud1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment := &pb.Enrollment{
		ID:           pendingEnrollment.ID,
		CourseID:     course.ID,
		UserID:       stud1.ID,
		Status:       pb.Enrollment_PENDING,
		State:        pb.Enrollment_VISIBLE,
		Course:       course,
		User:         stud1,
		UsedSlipDays: []*pb.UsedSlipDays{},
	}
	// can't use: wantEnrollment.User.RemoveRemoteID()
	wantEnrollment.User.RemoteIdentities = nil
	if diff := cmp.Diff(wantEnrollment, pendingEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +pendingEnrollment):\n%s", diff)
	}

	enrollStud1.Status = pb.Enrollment_STUDENT
	if _, err = ags.UpdateEnrollments(ctx, &pb.Enrollments{Enrollments: []*pb.Enrollment{enrollStud1}}); err != nil {
		t.Fatal(err)
	}

	// verify that the enrollment was updated to student status.
	gotEnrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, stud1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Status = pb.Enrollment_STUDENT
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// create another user and enroll as student

	stud2 := qtest.CreateFakeUser(t, db, 3)
	enrollStud2 := &pb.Enrollment{CourseID: course.ID, UserID: stud2.ID}
	if _, err = ags.CreateEnrollment(ctx, enrollStud2); err != nil {
		t.Fatal(err)
	}
	enrollStud2.Status = pb.Enrollment_STUDENT
	if _, err = ags.UpdateEnrollments(ctx, &pb.Enrollments{Enrollments: []*pb.Enrollment{enrollStud2}}); err != nil {
		t.Fatal(err)
	}
	// verify that the stud2 was enrolled with student status.
	gotEnrollment, err = db.GetEnrollmentByCourseAndUser(course.ID, stud2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = gotEnrollment.ID
	wantEnrollment.Status = pb.Enrollment_STUDENT
	wantEnrollment.UserID = stud2.ID
	wantEnrollment.User = stud2
	wantEnrollment.User.RemoteIdentities = nil
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// promote stud2 to teaching assistant

	enrollStud2.Status = pb.Enrollment_TEACHER
	if _, err = ags.UpdateEnrollments(ctx, &pb.Enrollments{Enrollments: []*pb.Enrollment{enrollStud2}}); err != nil {
		t.Fatal(err)
	}
	// verify that the stud2 was promoted to teacher status.
	gotEnrollment, err = db.GetEnrollmentByCourseAndUser(course.ID, stud2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = gotEnrollment.ID
	wantEnrollment.Status = pb.Enrollment_TEACHER
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}
}

func TestListCoursesWithEnrollment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	user := qtest.CreateFakeUser(t, db, 2)
	_, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	var testCourses []*pb.Course
	for _, course := range allCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, course)
	}

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	query := &pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
		Status:   pb.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	courses_request := &pb.EnrollmentStatusRequest{UserID: user.ID}
	courses, err := ags.GetCoursesByUser(context.Background(), courses_request)
	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*pb.Course{
		{ID: testCourses[0].ID, Enrolled: pb.Enrollment_PENDING},
		{ID: testCourses[1].ID, Enrolled: pb.Enrollment_NONE},
		{ID: testCourses[2].ID, Enrolled: pb.Enrollment_STUDENT},
		{ID: testCourses[3].ID, Enrolled: pb.Enrollment_NONE},
	}
	for i, course := range courses.Courses {
		if course.ID != wantCourses[i].ID {
			t.Errorf("have course %+v want %+v", course.ID, wantCourses[i].ID)
		}
		if course.Enrolled != wantCourses[i].Enrolled {
			t.Errorf("have course %+v want %+v", course.Enrolled, wantCourses[i].Enrolled)
		}
	}
}

func TestListCoursesWithEnrollmentStatuses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	var testCourses []*pb.Course
	for _, course := range allCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, course)
	}

	user := qtest.CreateFakeUser(t, db, 2)
	_, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}); err != nil {
		t.Fatal(err)
	}

	// user enrollment is rejected for course 1 and enrolled for course 2, still pending for course 0
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	query := &pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
		Status:   pb.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	stats := make([]pb.Enrollment_UserStatus, 0)
	stats = append(stats, pb.Enrollment_STUDENT)
	course_req := &pb.EnrollmentStatusRequest{UserID: user.ID, Statuses: stats}
	courses, err := ags.GetCoursesByUser(context.Background(), course_req)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses, err := db.GetCoursesByUser(user.ID, pb.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	gotCourses := courses.Courses
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCoursesByUser() mismatch (-wantCourses +gotCourses):\n%s", diff)
	}
}

func TestGetCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	wantCourse := allCourses[0]
	err := db.CreateCourse(admin.ID, wantCourse)
	if err != nil {
		t.Fatal(err)
	}
	_, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	gotCourse, err := ags.GetCourse(context.Background(), &pb.CourseRequest{CourseID: wantCourse.ID})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
	}
}

func TestPromoteDemoteRejectTeacher(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	teacher := qtest.CreateFakeUser(t, db, 10)
	student1 := qtest.CreateFakeUser(t, db, 11)
	student2 := qtest.CreateFakeUser(t, db, 12)
	ta := qtest.CreateFakeUser(t, db, 13)

	course := allCourses[0]
	err := db.CreateCourse(teacher.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   ta.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &pb.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	query.UserID = student1.ID
	query.Status = pb.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	query.UserID = student2.ID
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	query.UserID = ta.ID
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	student1Enrollment := &pb.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}
	student2Enrollment := &pb.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}
	taEnrollment := &pb.Enrollment{
		UserID:   ta.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}
	teacherEnrollment := &pb.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}

	request := &pb.Enrollments{}

	// student1 attempts to promote student2 to teacher, must fail
	request.Enrollments = []*pb.Enrollment{student2Enrollment}
	ctx := qtest.WithUserContext(context.Background(), student1)
	if _, err := ags.UpdateEnrollments(ctx, request); err == nil {
		t.Errorf("expected error: 'only teachers can update enrollment status'")
	}

	// teacher promotes students to teachers, must succeed
	ctx = qtest.WithUserContext(context.Background(), teacher)
	_, err = fakeProvider.CreateOrganization(ctx, &scm.OrganizationOptions{Path: "path", Name: "name"})
	if err != nil {
		t.Fatal(err)
	}

	request.Enrollments = []*pb.Enrollment{student1Enrollment, student2Enrollment, taEnrollment}
	if _, err := ags.UpdateEnrollments(ctx, request); err != nil {
		t.Fatal(err)
	}

	// TA attempts to demote self, must succeed
	taEnrollment.Status = pb.Enrollment_STUDENT
	request.Enrollments = []*pb.Enrollment{taEnrollment}
	ctx = qtest.WithUserContext(context.Background(), ta)
	if _, err := ags.UpdateEnrollments(ctx, request); err != nil {
		t.Fatal(err)
	}

	// student2 attempts to demote course creator, must fail
	teacherEnrollment.Status = pb.Enrollment_STUDENT
	request.Enrollments = []*pb.Enrollment{teacherEnrollment}
	ctx = qtest.WithUserContext(context.Background(), student2)
	if _, err := ags.UpdateEnrollments(ctx, request); err == nil {
		t.Error("expected error: 'only course creator can change status of other teachers'", err)
	}

	// student2 attempts to reject course creator, must fail
	teacherEnrollment.Status = pb.Enrollment_NONE
	if _, err := ags.UpdateEnrollments(ctx, request); err == nil {
		t.Error("expected error: 'only course creator can change status of other teachers'")
	}

	// teacher demotes student1, must succeed
	student1Enrollment.Status = pb.Enrollment_STUDENT
	request.Enrollments = []*pb.Enrollment{student1Enrollment}
	ctx = qtest.WithUserContext(context.Background(), teacher)
	if _, err := ags.UpdateEnrollments(ctx, request); err != nil {
		t.Fatal(err)
	}

	// check that student1 is now enrolled as student
	enrol, err := db.GetEnrollmentByCourseAndUser(course.ID, student1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if enrol.Status != pb.Enrollment_STUDENT {
		t.Errorf("expected status %s, got %s", pb.Enrollment_STUDENT, enrol.Status)
	}

	// teacher rejects student2, must succeed
	student2Enrollment.Status = pb.Enrollment_STUDENT
	request.Enrollments = []*pb.Enrollment{student2Enrollment}
	if _, err := ags.UpdateEnrollments(ctx, request); err != nil {
		t.Fatal(err)
	}
	student2Enrollment.Status = pb.Enrollment_NONE
	if _, err := ags.UpdateEnrollments(ctx, request); err != nil {
		t.Fatal(err)
	}

	// ensure that student2 is no longer enrolled in the course
	if _, err := db.GetEnrollmentByCourseAndUser(course.ID, student2.ID); err == nil {
		t.Error("expected error 'record not found'")
	}

	// justice is served

	// course creator attempts to demote himself, must fail as well
	teacherEnrollment.Status = pb.Enrollment_STUDENT
	request.Enrollments = []*pb.Enrollment{teacherEnrollment}
	if _, err := ags.UpdateEnrollments(ctx, request); err == nil {
		t.Error("expected error 'course creator cannot be demoted'")
	}

	// same when rejecting
	teacherEnrollment.Status = pb.Enrollment_NONE
	if _, err := ags.UpdateEnrollments(ctx, request); err == nil {
		t.Error("expected error 'course creator cannot be demoted'")
	}

	// ta attempts to demote course creator, must fail
	teacherEnrollment.Status = pb.Enrollment_STUDENT
	ctx = qtest.WithUserContext(context.Background(), ta)
	if _, err := ags.UpdateEnrollments(ctx, request); err == nil {
		t.Error("expected error 'ta cannot be demoted course creator'")
	}
}
