package web_test

import (
	"context"
	"os"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestGetCourses(t *testing.T) {
	db, cleanup, _, qfService := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)

	var wantCourses []*qf.Course
	for _, course := range qtest.MockCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
		wantCourses = append(wantCourses, course)
	}

	foundCourses, err := qfService.GetCourses(context.Background(), connect.NewRequest(&qf.Void{}))
	if err != nil {
		t.Fatal(err)
	}
	gotCourses := foundCourses.Msg.Courses
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetCourses() mismatch (-wantCourses +gotCourses):\n%s", diff)
	}
}

func TestNewCourse(t *testing.T) {
	db, cleanup, _, qfService := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateAdminUser(t, db, "fake")
	ctx := auth.WithUserContext(context.Background(), admin)

	for _, wantCourse := range qtest.MockCourses {
		gotCourse, err := qfService.CreateCourse(ctx, connect.NewRequest(wantCourse))
		if err != nil {
			t.Fatal(err)
		}
		wantCourse.ID = gotCourse.Msg.ID
		if diff := cmp.Diff(wantCourse, gotCourse.Msg, protocmp.Transform()); diff != "" {
			t.Errorf("ags.CreateCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
		}

		// check that the database also has the course
		gotCourse.Msg, err = db.GetCourse(wantCourse.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(wantCourse, gotCourse.Msg, protocmp.Transform()); diff != "" {
			t.Errorf("db.GetCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
		}
	}
}

func TestNewCourseExistingRepos(t *testing.T) {
	db, cleanup, mockSCM, qfService := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	ctx := auth.WithUserContext(context.Background(), admin)

	organization, err := mockSCM.GetOrganization(ctx, &scm.GetOrgOptions{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	for path, private := range web.RepoPaths {
		repoOptions := &scm.CreateRepositoryOptions{Path: path, Organization: organization.Name, Private: private}
		_, err := mockSCM.CreateRepository(ctx, repoOptions)
		if err != nil {
			t.Fatal(err)
		}
	}

	course, err := qfService.CreateCourse(ctx, connect.NewRequest(qtest.MockCourses[0]))
	if course != nil {
		t.Fatal("expected CreateCourse to fail with AlreadyExists")
	}
	if err != nil && connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("expected CreateCourse to fail with AlreadyExists, but got: %v", err)
	}
}

func TestEnrollmentProcess(t *testing.T) {
	if os.Getenv("TODO") == "" {
		t.Skip("See TODO description")
	}
	// TODO(meling): This test no longer passes since the enrollment process includes accepting invitations on behalf of the user.
	// A fix would probably be to implement a fake SCMInvite that behaves appropriately.
	// We should add manual SCM_TEST for the actual AcceptRepositoryInvites using qf101.
	db, cleanup, _, qfService := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	ctx := auth.WithUserContext(context.Background(), admin)

	course, err := qfService.CreateCourse(ctx, connect.NewRequest(qtest.MockCourses[0]))
	if err != nil {
		t.Fatal(err)
	}

	stud1 := qtest.CreateFakeUser(t, db, 2)
	enrollStud1 := &qf.Enrollment{CourseID: course.Msg.ID, UserID: stud1.ID}
	if _, err = qfService.CreateEnrollment(ctx, connect.NewRequest(enrollStud1)); err != nil {
		t.Fatal(err)
	}

	// verify that a pending enrollment was indeed created.
	pendingEnrollment, err := db.GetEnrollmentByCourseAndUser(course.Msg.ID, stud1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment := &qf.Enrollment{
		ID:           pendingEnrollment.ID,
		CourseID:     course.Msg.ID,
		UserID:       stud1.ID,
		Status:       qf.Enrollment_PENDING,
		State:        qf.Enrollment_VISIBLE,
		Course:       course.Msg,
		User:         stud1,
		UsedSlipDays: []*qf.UsedSlipDays{},
	}
	// can't use: wantEnrollment.User.RemoveRemoteID()
	wantEnrollment.User.RemoteIdentities = nil
	if diff := cmp.Diff(wantEnrollment, pendingEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +pendingEnrollment):\n%s", diff)
	}

	enrollStud1.Status = qf.Enrollment_STUDENT
	if _, err = qfService.UpdateEnrollments(ctx, connect.NewRequest(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{enrollStud1},
	})); err != nil {
		t.Fatal(err)
	}

	// verify that the enrollment was updated to student status.
	gotEnrollment, err := db.GetEnrollmentByCourseAndUser(course.Msg.ID, stud1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Status = qf.Enrollment_STUDENT
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// create another user and enroll as student

	stud2 := qtest.CreateFakeUser(t, db, 3)
	enrollStud2 := &qf.Enrollment{CourseID: course.Msg.ID, UserID: stud2.ID}
	if _, err = qfService.CreateEnrollment(ctx, connect.NewRequest(enrollStud2)); err != nil {
		t.Fatal(err)
	}
	enrollStud2.Status = qf.Enrollment_STUDENT
	if _, err = qfService.UpdateEnrollments(ctx, connect.NewRequest(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			enrollStud2,
		},
	})); err != nil {
		t.Fatal(err)
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
	wantEnrollment.User.RemoteIdentities = nil
	if diff := cmp.Diff(wantEnrollment, gotEnrollment, protocmp.Transform()); diff != "" {
		t.Errorf("EnrollmentProcess mismatch (-wantEnrollment +gotEnrollment):\n%s", diff)
	}

	// promote stud2 to teaching assistant

	enrollStud2.Status = qf.Enrollment_TEACHER
	if _, err = qfService.UpdateEnrollments(ctx, connect.NewRequest(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			enrollStud2,
		},
	})); err != nil {
		t.Fatal(err)
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
	db, cleanup, _, qfService := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	user := qtest.CreateFakeUser(t, db, 2)

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
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
		Status:   qf.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	courses_request := &qf.EnrollmentStatusRequest{UserID: user.ID}
	courses, err := qfService.GetCoursesByUser(context.Background(), connect.NewRequest(courses_request))
	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*qf.Course{
		{ID: testCourses[0].ID, Enrolled: qf.Enrollment_PENDING},
		{ID: testCourses[1].ID, Enrolled: qf.Enrollment_NONE},
		{ID: testCourses[2].ID, Enrolled: qf.Enrollment_STUDENT},
		{ID: testCourses[3].ID, Enrolled: qf.Enrollment_NONE},
	}
	for i, course := range courses.Msg.Courses {
		if course.ID != wantCourses[i].ID {
			t.Errorf("have course %+v want %+v", course.ID, wantCourses[i].ID)
		}
		if course.Enrolled != wantCourses[i].Enrolled {
			t.Errorf("have course %+v want %+v", course.Enrolled, wantCourses[i].Enrolled)
		}
	}
}

func TestListCoursesWithEnrollmentStatuses(t *testing.T) {
	db, cleanup, _, qfService := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	var testCourses []*qf.Course
	for _, course := range qtest.MockCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, course)
	}

	user := qtest.CreateFakeUser(t, db, 2)

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
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}); err != nil {
		t.Fatal(err)
	}

	// user enrollment is rejected for course 1 and enrolled for course 2, still pending for course 0
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
		Status:   qf.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	stats := make([]qf.Enrollment_UserStatus, 0)
	stats = append(stats, qf.Enrollment_STUDENT)
	course_req := &qf.EnrollmentStatusRequest{UserID: user.ID, Statuses: stats}
	courses, err := qfService.GetCoursesByUser(context.Background(), connect.NewRequest(course_req))
	if err != nil {
		t.Fatal(err)
	}
	wantCourses, err := db.GetCoursesByUser(user.ID, qf.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	gotCourses := courses.Msg.Courses
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCoursesByUser() mismatch (-wantCourses +gotCourses):\n%s", diff)
	}
}

func TestGetCourse(t *testing.T) {
	db, cleanup, _, qfService := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	wantCourse := qtest.MockCourses[0]
	err := db.CreateCourse(admin.ID, wantCourse)
	if err != nil {
		t.Fatal(err)
	}

	gotCourse, err := qfService.GetCourse(context.Background(), connect.NewRequest(&qf.CourseRequest{
		CourseID: wantCourse.ID,
	}))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetCourse() mismatch (-wantCourse +gotCourse):\n%s", diff)
	}
}

func TestPromoteDemoteRejectTeacher(t *testing.T) {
	db, cleanup, mockSCM, qfService := testQuickFeedService(t)
	defer cleanup()

	teacher := qtest.CreateAdminUser(t, db, "fake")
	student1 := qtest.CreateNamedUser(t, db, 11, "student1")
	student2 := qtest.CreateNamedUser(t, db, 12, "student2")
	ta := qtest.CreateNamedUser(t, db, 13, "TA")

	course := qtest.MockCourses[0]
	err := db.CreateCourse(teacher.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   ta.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	query.UserID = student1.ID
	query.Status = qf.Enrollment_STUDENT
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

	student1Enrollment := &qf.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}
	student2Enrollment := &qf.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}
	taEnrollment := &qf.Enrollment{
		UserID:   ta.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}
	teacherEnrollment := &qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}

	request := &qf.Enrollments{}

	// teacher promotes students to teachers, must succeed
	ctx := auth.WithUserContext(context.Background(), teacher)
	// Need course teams to update enrollments.
	if _, err := mockSCM.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: qtest.MockOrg,
		TeamName:     "allstudents",
	}); err != nil {
		t.Error(err)
	}
	if _, err := mockSCM.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: qtest.MockOrg,
		TeamName:     "allteachers",
	}); err != nil {
		t.Error(err)
	}

	request.Enrollments = []*qf.Enrollment{student1Enrollment, student2Enrollment, taEnrollment}
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err != nil {
		t.Fatal(err)
	}

	// TA attempts to demote self, must succeed
	taEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{taEnrollment}
	ctx = auth.WithUserContext(context.Background(), ta)
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err != nil {
		t.Fatal(err)
	}

	// student2 attempts to demote course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{teacherEnrollment}
	ctx = auth.WithUserContext(context.Background(), student2)
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err == nil {
		t.Error("expected error: 'only course creator can change status of other teachers'", err)
	}

	// student2 attempts to reject course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_NONE
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err == nil {
		t.Error("expected error: 'only course creator can change status of other teachers'")
	}

	// teacher demotes student1, must succeed
	student1Enrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{student1Enrollment}
	ctx = auth.WithUserContext(context.Background(), teacher)
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err != nil {
		t.Fatal(err)
	}

	// check that student1 is now enrolled as student
	enrol, err := db.GetEnrollmentByCourseAndUser(course.ID, student1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if enrol.Status != qf.Enrollment_STUDENT {
		t.Errorf("expected status %s, got %s", qf.Enrollment_STUDENT, enrol.Status)
	}

	// teacher rejects student2, must succeed
	student2Enrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{student2Enrollment}
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err != nil {
		t.Fatal(err)
	}
	student2Enrollment.Status = qf.Enrollment_NONE
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err != nil {
		t.Fatal(err)
	}

	// ensure that student2 is no longer enrolled in the course
	if _, err := db.GetEnrollmentByCourseAndUser(course.ID, student2.ID); err == nil {
		t.Error("expected error 'record not found'")
	}

	// justice is served

	// course creator attempts to demote himself, must fail as well
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	request.Enrollments = []*qf.Enrollment{teacherEnrollment}
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err == nil {
		t.Error("expected error 'course creator cannot be demoted'")
	}

	// same when rejecting
	teacherEnrollment.Status = qf.Enrollment_NONE
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err == nil {
		t.Error("expected error 'course creator cannot be demoted'")
	}

	// ta attempts to demote course creator, must fail
	teacherEnrollment.Status = qf.Enrollment_STUDENT
	ctx = auth.WithUserContext(context.Background(), ta)
	if _, err := qfService.UpdateEnrollments(ctx, connect.NewRequest(request)); err == nil {
		t.Error("expected error 'ta cannot be demoted course creator'")
	}
}

func TestUpdateCourseVisibility(t *testing.T) {
	db, cleanup, _, _ := testQuickFeedService(t)
	defer cleanup()

	logger := qtest.Logger(t)

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

	interceptors := connect.WithInterceptors(
		interceptor.NewMetricsInterceptor(),
		interceptor.NewValidationInterceptor(logger),
		interceptor.NewUserInterceptor(logger, tm),
		interceptor.NewAccessControlInterceptor(tm),
		interceptor.NewTokenInterceptor(tm),
	)
	shutdown, client := MockQuickFeedClient(t, db, interceptors)

	ctx := context.Background()
	defer shutdown(ctx)

	teacher := qtest.CreateAdminUser(t, db, "fake")

	user := qtest.CreateFakeUser(t, db, 2)
	userCookie, err := tm.NewAuthCookie(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	cookie := userCookie.String()
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

	req := &qf.EnrollmentStatusRequest{
		UserID: user.ID,
	}
	enrollments, err := client.GetEnrollmentsByUser(auth.WithUserContext(ctx, user), qtest.RequestWithCookie(req, cookie))
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

	if _, err := client.UpdateCourseVisibility(auth.WithUserContext(ctx, user), qtest.RequestWithCookie(enrollment, cookie)); err != nil {
		t.Error(err)
	}

	gotEnrollments, err := client.GetEnrollmentsByUser(auth.WithUserContext(ctx, user), qtest.RequestWithCookie(req, cookie))
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
