package interceptor_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	BufSize = 1024 * 1024
)

type accessTests []struct {
	name     string
	ctx      context.Context
	userID   uint64
	courseID uint64
	groupID  uint64
	access   bool
}

func TestAccessControlMethods(t *testing.T) {
	service := reflect.TypeOf(qfconnect.UnimplementedQuickFeedServiceHandler{})
	methods := make([]string, 0, service.NumMethod())
	for i := 0; i < service.NumMethod(); i++ {
		methods = append(methods, service.Method(i).Name)
	}
	if err := web.VerifyAccessControlMethods(methods); err != nil {
		t.Error(err)
	}
}

func TestAccessControl(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, scm.TestSCMManager(), web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db, "test")
	if err != nil {
		t.Fatal(err)
	}
	interceptors := connect.WithInterceptors(
		interceptor.AccessControl(tm),
	)

	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(ags, interceptors))
	muxServer := &http.Server{
		Handler: h2c.NewHandler(router, &http2.Server{}),
		Addr:    "127.0.0.1:8081",
	}

	go func() {
		if err := muxServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				t.Errorf("Server exited with unexpected error: %v", err)
			}
			return
		}
	}()

	client := qtest.QuickFeedClient("")

	courseAdmin := qtest.CreateAdminUser(t, db, "fake")
	groupStudent := qtest.CreateNamedUser(t, db, 2, "group student")
	student := qtest.CreateNamedUser(t, db, 3, "student")
	user := qtest.CreateNamedUser(t, db, 4, "user")
	studentCourseAdmin := qtest.CreateFakeUser(t, db, 5)
	admin := qtest.CreateFakeUser(t, db, 6)
	admin.IsAdmin = true
	studentCourseAdmin.IsAdmin = true
	if err := db.UpdateUser(admin); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateUser(studentCourseAdmin); err != nil {
		t.Fatal(err)
	}

	course := &qf.Course{
		Code:             "test101",
		Year:             2022,
		Provider:         "fake",
		OrganizationID:   1,
		OrganizationPath: "testorg",
		CourseCreatorID:  courseAdmin.ID,
	}
	if err := db.CreateCourse(courseAdmin.ID, course); err != nil {
		t.Fatal(err)
	}
	qtest.EnrollStudent(t, db, groupStudent, course)
	qtest.EnrollStudent(t, db, student, course)
	qtest.EnrollStudent(t, db, studentCourseAdmin, course)
	group := &qf.Group{
		CourseID: course.ID,
		Name:     "Test",
		Users:    []*qf.User{groupStudent},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	assignment := &qf.Assignment{
		CourseID: course.ID,
		Name:     "Test Assignment",
		Order:    1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       groupStudent.ID,
	}); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	f := func(t *testing.T, id uint64) context.Context {
		token, err := tm.NewAuthCookie(id)
		if err != nil {
			t.Fatal(err)
		}
		return qtest.WithAuthCookie(ctx, token)
	}
	courseAdminContext := f(t, courseAdmin.ID)
	groupStudentContext := f(t, groupStudent.ID)
	studentContext := f(t, student.ID)
	userContext := f(t, user.ID)
	studentAdminContext := f(t, studentCourseAdmin.ID)
	adminContext := f(t, admin.ID)

	freeAccessTest := accessTests{
		{"admin", courseAdminContext, 0, course.ID, 0, true},
		{"student", studentContext, 0, course.ID, 0, true},
		{"student", groupStudentContext, 0, course.ID, 0, true},
		{"user", userContext, 0, course.ID, 0, true},
		{"non-teacher admin", adminContext, 0, course.ID, 0, true},
		{"empty context", ctx, 0, course.ID, 0, false},
	}
	for _, tt := range freeAccessTest {
		t.Run("UnrestrictedAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetUser(tt.ctx, connect.NewRequest(&qf.Void{}))
			checkAccess(t, err, tt.access, "GetUser")
			_, err = client.GetCourse(tt.ctx, connect.NewRequest(&qf.CourseRequest{CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "GetCourse")
			_, err = client.GetCourses(tt.ctx, connect.NewRequest(&qf.Void{}))
			checkAccess(t, err, tt.access, "GetCourses")
		})
	}

	userAccessTests := accessTests{
		{"correct user ID", userContext, user.ID, course.ID, 0, true},
		{"incorrect user ID", userContext, groupStudent.ID, course.ID, 0, false},
	}
	for _, tt := range userAccessTests {
		t.Run("UserAccess/"+tt.name, func(t *testing.T) {
			enrol := &qf.Enrollment{
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}
			enrolRequest := &qf.EnrollmentStatusRequest{
				UserID: tt.userID,
			}
			_, err := client.CreateEnrollment(tt.ctx, connect.NewRequest(enrol))
			checkAccess(t, err, tt.access, "CreateEnrollment")
			_, err = client.UpdateCourseVisibility(tt.ctx, connect.NewRequest(enrol))
			checkAccess(t, err, tt.access, "UpdateCourseVisibility")
			_, err = client.GetCoursesByUser(tt.ctx, connect.NewRequest(enrolRequest))
			checkAccess(t, err, tt.access, "GetCoursesByUser")
			_, err = client.UpdateUser(tt.ctx, connect.NewRequest(&qf.User{ID: tt.userID}))
			checkAccess(t, err, tt.access, "UpdateUser")
			_, err = client.GetEnrollmentsByUser(tt.ctx, connect.NewRequest(enrolRequest))
			checkAccess(t, err, tt.access, "GetEnrollmentsByCourse")
			_, err = client.UpdateUser(tt.ctx, connect.NewRequest(&qf.User{ID: tt.userID}))
			checkAccess(t, err, tt.access, "UpdateUser")
		})
	}

	studentAccessTests := accessTests{
		{"course admin", courseAdminContext, courseAdmin.ID, course.ID, 0, true},
		{"admin, not enrolled in a course", adminContext, admin.ID, course.ID, 0, false},
		{"user, not enrolled in the course", userContext, user.ID, course.ID, 0, false},
		{"student", studentContext, student.ID, course.ID, 0, true},
		{"student of another course", studentContext, student.ID, 123, 0, false},
	}
	for _, tt := range studentAccessTests {
		t.Run("StudentAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetSubmissions(tt.ctx, connect.NewRequest(&qf.SubmissionRequest{
				UserID:   tt.userID,
				CourseID: tt.courseID,
			}))
			checkAccess(t, err, tt.access, "GetSubmissions")
			_, err = client.GetAssignments(tt.ctx, connect.NewRequest(&qf.CourseRequest{CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "GetAssignments")
			_, err = client.GetEnrollmentsByCourse(tt.ctx, connect.NewRequest(&qf.EnrollmentRequest{CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "GetEnrollmentsByCourse")
			_, err = client.GetRepositories(tt.ctx, connect.NewRequest(&qf.URLRequest{CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "GetRepositories")
			_, err = client.GetGroupByUserAndCourse(tt.ctx, connect.NewRequest(&qf.GroupRequest{
				CourseID: tt.courseID,
				UserID:   tt.userID,
				GroupID:  0,
			}))
			checkAccess(t, err, tt.access, "GetGroupByUserAndCourse")
		})
	}

	groupAccessTests := accessTests{
		{"student in a group", groupStudentContext, groupStudent.ID, course.ID, group.ID, true},
		{"student, not in a group", studentContext, student.ID, course.ID, group.ID, false},
		{"student in a group, wrong group ID in request", studentContext, student.ID, course.ID, 123, false},
	}
	for _, tt := range groupAccessTests {
		t.Run("GroupAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetGroup(tt.ctx, connect.NewRequest(&qf.GetGroupRequest{GroupID: tt.groupID}))
			checkAccess(t, err, tt.access, "GetGroup")
		})
	}

	teacherAccessTests := accessTests{
		{"course teacher", courseAdminContext, groupStudent.ID, course.ID, group.ID, true},
		{"student", studentContext, student.ID, course.ID, group.ID, false},
		{"admin, not enrolled in the course", adminContext, admin.ID, course.ID, group.ID, false},
	}
	for _, tt := range teacherAccessTests {
		t.Run("TeacherAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetGroup(tt.ctx, connect.NewRequest(&qf.GetGroupRequest{GroupID: tt.groupID}))
			checkAccess(t, err, tt.access, "GetGroup")
			_, err = client.DeleteGroup(tt.ctx, connect.NewRequest(&qf.GroupRequest{
				GroupID:  tt.groupID,
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}))
			checkAccess(t, err, tt.access, "DeleteGroup")
			_, err = client.UpdateGroup(tt.ctx, connect.NewRequest(&qf.Group{CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "UpdateGroup")
			_, err = client.UpdateCourse(tt.ctx, connect.NewRequest(course))
			checkAccess(t, err, tt.access, "UpdateCourse")
			_, err = client.UpdateEnrollments(tt.ctx, connect.NewRequest(&qf.Enrollments{
				Enrollments: []*qf.Enrollment{{ID: 1, CourseID: tt.courseID}},
			}))
			checkAccess(t, err, tt.access, "UpdateEnrollments")
			_, err = client.UpdateAssignments(tt.ctx, connect.NewRequest(&qf.CourseRequest{CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "UpdateAssignments")
			_, err = client.UpdateSubmission(tt.ctx, connect.NewRequest(&qf.UpdateSubmissionRequest{SubmissionID: 1, CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "UpdateSubmission")
			_, err = client.UpdateSubmissions(tt.ctx, connect.NewRequest(&qf.UpdateSubmissionsRequest{AssignmentID: 1, CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "UpdateSubmissions")
			_, err = client.RebuildSubmissions(tt.ctx, connect.NewRequest(&qf.RebuildRequest{
				AssignmentID: 1,
				RebuildType: &qf.RebuildRequest_CourseID{
					CourseID: tt.courseID,
				},
			}))
			checkAccess(t, err, tt.access, "RebuildSubmissions")
			_, err = client.CreateBenchmark(tt.ctx, connect.NewRequest(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}))
			checkAccess(t, err, tt.access, "CreateBenchmark")
			_, err = client.UpdateBenchmark(tt.ctx, connect.NewRequest(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}))
			checkAccess(t, err, tt.access, "UpdateBenchmark")
			_, err = client.DeleteBenchmark(tt.ctx, connect.NewRequest(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}))
			checkAccess(t, err, tt.access, "DeleteBenchmark")
			_, err = client.CreateCriterion(tt.ctx, connect.NewRequest(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}))
			checkAccess(t, err, tt.access, "CreateCriterion")
			_, err = client.UpdateCriterion(tt.ctx, connect.NewRequest(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}))
			checkAccess(t, err, tt.access, "UpdateCriterion")
			_, err = client.DeleteCriterion(tt.ctx, connect.NewRequest(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}))
			checkAccess(t, err, tt.access, "DeleteCriterion")
			_, err = client.CreateReview(tt.ctx, connect.NewRequest(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}))
			checkAccess(t, err, tt.access, "CreateReview")
			_, err = client.UpdateReview(tt.ctx, connect.NewRequest(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}))
			checkAccess(t, err, tt.access, "UpdateReview")
			_, err = client.GetReviewers(tt.ctx, connect.NewRequest(&qf.SubmissionReviewersRequest{
				CourseID:     tt.courseID,
				SubmissionID: 1,
			}))
			checkAccess(t, err, tt.access, "GetReviewers")
			_, err = client.IsEmptyRepo(tt.ctx, connect.NewRequest(&qf.RepositoryRequest{CourseID: tt.courseID}))
			checkAccess(t, err, tt.access, "IsEmptyRepo")
		})
	}

	courseAdminTests := accessTests{
		{"admin, not enrolled", adminContext, 0, course.ID, 0, false},
		{"course admin, not a teacher", studentAdminContext, 0, course.ID, 0, true},
		{"course admin, wrong course in request", studentAdminContext, 0, 123, 0, false},
	}
	for _, tt := range courseAdminTests {
		t.Run("CourseAdminAccess/"+tt.name, func(t *testing.T) {
			_, err = client.GetSubmissionsByCourse(tt.ctx, connect.NewRequest(&qf.SubmissionsForCourseRequest{
				CourseID: tt.courseID,
			}))
			checkAccess(t, err, tt.access, "GetSubmissionsByCourse")
		})
	}

	adminAccessTests := accessTests{
		{"admin (accessing own info)", courseAdminContext, courseAdmin.ID, course.ID, group.ID, true},
		{"admin (accessing other user's info)", courseAdminContext, user.ID, course.ID, group.ID, true},
		{"non admin (accessing admin's info)", studentContext, courseAdmin.ID, course.ID, group.ID, false},
		{"non admin (accessing other user's info)", studentContext, user.ID, course.ID, group.ID, false},
	}
	for _, tt := range adminAccessTests {
		t.Run("AdminAccess/"+tt.name, func(t *testing.T) {
			_, err := client.UpdateUser(tt.ctx, connect.NewRequest(&qf.User{ID: tt.userID}))
			checkAccess(t, err, tt.access, "UpdateUser")
			_, err = client.GetEnrollmentsByUser(tt.ctx, connect.NewRequest(&qf.EnrollmentStatusRequest{UserID: tt.userID}))
			checkAccess(t, err, tt.access, "GetEnrollmentsByUser")
			_, err = client.GetUsers(tt.ctx, connect.NewRequest(&qf.Void{}))
			checkAccess(t, err, tt.access, "GetUsers")
			_, err = client.GetOrganization(tt.ctx, connect.NewRequest(&qf.OrgRequest{OrgName: "testorg"}))
			checkAccess(t, err, tt.access, "GetOrganization")
			_, err = client.CreateCourse(tt.ctx, connect.NewRequest(course))
			checkAccess(t, err, tt.access, "CreateCourse")
			_, err = client.GetUserByCourse(tt.ctx, connect.NewRequest(&qf.CourseUserRequest{
				CourseCode: course.Code,
				CourseYear: course.Year,
				UserLogin:  "student",
			}))
			checkAccess(t, err, tt.access, "GetUserByCourse")
		})
	}

	createGroupTests := []struct {
		name   string
		ctx    context.Context
		group  *qf.Group
		access bool
	}{
		{"valid student, not in the request group", studentContext, &qf.Group{
			CourseID: course.ID,
		}, false},
		{"valid student", studentContext, &qf.Group{
			Name:     "test",
			CourseID: course.ID,
			Users:    []*qf.User{student},
		}, true},
		{"course teacher", courseAdminContext, &qf.Group{
			CourseID: course.ID,
			Users:    []*qf.User{courseAdmin},
		}, true},
		{"admin, not a teacher", adminContext, &qf.Group{
			CourseID: course.ID,
		}, false},
	}

	for _, tt := range createGroupTests {
		t.Run("CreateGroupAccess/"+tt.name, func(t *testing.T) {
			_, err := client.CreateGroup(tt.ctx, connect.NewRequest(tt.group))
			checkAccess(t, err, tt.access, "CreateGroup")
		})
	}

	adminStatusChangeTests := []struct {
		name   string
		ctx    context.Context
		user   *qf.User
		access bool
	}{
		{"admin demoting a user", courseAdminContext, &qf.User{
			ID:      admin.ID,
			IsAdmin: false,
		}, true},
		{"admin promoting a user", courseAdminContext, &qf.User{
			ID:      admin.ID,
			IsAdmin: true,
		}, true},
		{"admin demoting self", courseAdminContext, &qf.User{
			ID:      courseAdmin.ID,
			IsAdmin: false,
		}, true},
		{"user promoting another user", userContext, &qf.User{
			ID:      groupStudent.ID,
			IsAdmin: true,
		}, false},
		{"user promoting self", userContext, &qf.User{
			ID:      user.ID,
			IsAdmin: true,
		}, false},
	}

	for _, tt := range adminStatusChangeTests {
		t.Run("AdminStatusChange/"+tt.name, func(t *testing.T) {
			_, err := client.UpdateUser(tt.ctx, connect.NewRequest(tt.user))
			checkAccess(t, err, tt.access, "UpdateUser")
		})
	}

	if err = muxServer.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}

func checkAccess(t *testing.T, err error, wantAccess bool, method string) {
	t.Helper()
	if connErr, ok := err.(*connect.Error); ok {
		gotCode := connErr.Code()
		wantCode := connect.CodePermissionDenied
		gotAccess := gotCode == wantCode
		if gotAccess == wantAccess {
			t.Errorf("%23s: (%v == %v) = %t, want %t", method, gotCode, wantCode, gotAccess, !wantAccess)
		}
	} else if err != nil && wantAccess {
		// got error and want access; expected non-error or not access
		t.Errorf("%23s: got %v (%t), want <nil> (%t)", method, err, wantAccess, !wantAccess)
	}
}
