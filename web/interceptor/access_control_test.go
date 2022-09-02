package interceptor_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

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
	cookie   string
	userID   uint64
	courseID uint64
	groupID  uint64
	access   bool
}

// TestAccessControlMethods checks that all QuickFeedService methods have an entry
// in the access control list.
func TestAccessControlMethods(t *testing.T) {
	service := reflect.TypeOf(qfconnect.UnimplementedQuickFeedServiceHandler{})
	serviceMethods := make(map[string]bool)
	for i := 0; i < service.NumMethod(); i++ {
		serviceMethods[service.Method(i).Name] = true
	}
	if err := interceptor.CheckAccessMethods(serviceMethods); err != nil {
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
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		Addr:              "127.0.0.1:8081",
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
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
	admin := qtest.CreateFakeUser(t, db, 6)
	admin.IsAdmin = true
	if err := db.UpdateUser(admin); err != nil {
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
	f := func(t *testing.T, id uint64) string {
		cookie, err := tm.NewAuthCookie(id)
		if err != nil {
			t.Fatal(err)
		}
		return auth.CookieString(cookie)
	}
	courseAdminCookie := f(t, courseAdmin.ID)
	groupStudentCookie := f(t, groupStudent.ID)
	studentCookie := f(t, student.ID)
	userCookie := f(t, user.ID)
	adminCookie := f(t, admin.ID)

	freeAccessTest := accessTests{
		{"admin", courseAdminCookie, 0, course.ID, 0, true},
		{"student", studentCookie, 0, course.ID, 0, true},
		{"student", groupStudentCookie, 0, course.ID, 0, true},
		{"user", userCookie, 0, course.ID, 0, true},
		{"non-teacher admin", adminCookie, 0, course.ID, 0, true},
		{"empty context", "", 0, course.ID, 0, false},
	}
	for _, tt := range freeAccessTest {
		t.Run("UnrestrictedAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetUser(ctx, requestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, err, tt.access, "GetUser")
			_, err = client.GetCourse(ctx, requestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "GetCourse")
			_, err = client.GetCourses(ctx, requestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, err, tt.access, "GetCourses")
		})
	}

	userAccessTests := accessTests{
		{"correct user ID", userCookie, user.ID, course.ID, 0, true},
		{"incorrect user ID", userCookie, groupStudent.ID, course.ID, 0, false},
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
			_, err := client.CreateEnrollment(ctx, requestWithCookie(enrol, tt.cookie))
			checkAccess(t, err, tt.access, "CreateEnrollment")
			_, err = client.UpdateCourseVisibility(ctx, requestWithCookie(enrol, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateCourseVisibility")
			_, err = client.GetCoursesByUser(ctx, requestWithCookie(enrolRequest, tt.cookie))
			checkAccess(t, err, tt.access, "GetCoursesByUser")
			_, err = client.UpdateUser(ctx, requestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateUser")
			_, err = client.GetEnrollmentsByUser(ctx, requestWithCookie(enrolRequest, tt.cookie))
			checkAccess(t, err, tt.access, "GetEnrollmentsByCourse")
			_, err = client.UpdateUser(ctx, requestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateUser")
		})
	}

	studentAccessTests := accessTests{
		{"course admin", courseAdminCookie, courseAdmin.ID, course.ID, 0, true},
		{"admin, not enrolled in a course", adminCookie, admin.ID, course.ID, 0, false},
		{"user, not enrolled in the course", userCookie, user.ID, course.ID, 0, false},
		{"student", studentCookie, student.ID, course.ID, 0, true},
		{"student of another course", studentCookie, student.ID, 123, 0, false},
	}
	for _, tt := range studentAccessTests {
		t.Run("StudentAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetSubmissions(ctx, requestWithCookie(&qf.SubmissionRequest{
				UserID:   tt.userID,
				CourseID: tt.courseID,
			}, tt.cookie))
			checkAccess(t, err, tt.access, "GetSubmissions")
			_, err = client.GetAssignments(ctx, requestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "GetAssignments")
			_, err = client.GetEnrollmentsByCourse(ctx, requestWithCookie(&qf.EnrollmentRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "GetEnrollmentsByCourse")
			_, err = client.GetRepositories(ctx, requestWithCookie(&qf.URLRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "GetRepositories")
			_, err = client.GetGroupByUserAndCourse(ctx, requestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				UserID:   tt.userID,
				GroupID:  0,
			}, tt.cookie))
			checkAccess(t, err, tt.access, "GetGroupByUserAndCourse")
		})
	}

	groupAccessTests := accessTests{
		{"student in a group", groupStudentCookie, groupStudent.ID, course.ID, group.ID, true},
		{"student, not in a group", studentCookie, student.ID, course.ID, group.ID, false},
		{"student in a group, wrong group ID in request", studentCookie, student.ID, course.ID, 123, false},
	}
	for _, tt := range groupAccessTests {
		t.Run("GroupAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetGroup(ctx, requestWithCookie(&qf.GetGroupRequest{GroupID: tt.groupID}, tt.cookie))
			checkAccess(t, err, tt.access, "GetGroup")
		})
	}

	teacherAccessTests := accessTests{
		{"course teacher", courseAdminCookie, groupStudent.ID, course.ID, group.ID, true},
		{"student", studentCookie, student.ID, course.ID, group.ID, false},
		{"admin, not enrolled in the course", adminCookie, admin.ID, course.ID, group.ID, false},
	}
	for _, tt := range teacherAccessTests {
		t.Run("TeacherAccess/"+tt.name, func(t *testing.T) {
			_, err := client.GetGroup(ctx, requestWithCookie(&qf.GetGroupRequest{GroupID: tt.groupID}, tt.cookie))
			checkAccess(t, err, tt.access, "GetGroup")
			_, err = client.DeleteGroup(ctx, requestWithCookie(&qf.GroupRequest{
				GroupID:  tt.groupID,
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}, tt.cookie))
			checkAccess(t, err, tt.access, "DeleteGroup")
			_, err = client.UpdateGroup(ctx, requestWithCookie(&qf.Group{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateGroup")
			_, err = client.UpdateCourse(ctx, requestWithCookie(course, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateCourse")
			_, err = client.UpdateEnrollments(ctx, requestWithCookie(&qf.Enrollments{
				Enrollments: []*qf.Enrollment{{ID: 1, CourseID: tt.courseID}},
			}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateEnrollments")
			_, err = client.UpdateAssignments(ctx, requestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateAssignments")
			_, err = client.UpdateSubmission(ctx, requestWithCookie(&qf.UpdateSubmissionRequest{SubmissionID: 1, CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateSubmission")
			_, err = client.UpdateSubmissions(ctx, requestWithCookie(&qf.UpdateSubmissionsRequest{AssignmentID: 1, CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateSubmissions")
			_, err = client.RebuildSubmissions(ctx, requestWithCookie(&qf.RebuildRequest{
				AssignmentID: 1,
				CourseID:     tt.courseID,
			}, tt.cookie))
			checkAccess(t, err, tt.access, "RebuildSubmissions")
			_, err = client.CreateBenchmark(ctx, requestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, err, tt.access, "CreateBenchmark")
			_, err = client.UpdateBenchmark(ctx, requestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateBenchmark")
			_, err = client.DeleteBenchmark(ctx, requestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, err, tt.access, "DeleteBenchmark")
			_, err = client.CreateCriterion(ctx, requestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, err, tt.access, "CreateCriterion")
			_, err = client.UpdateCriterion(ctx, requestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateCriterion")
			_, err = client.DeleteCriterion(ctx, requestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, err, tt.access, "DeleteCriterion")
			_, err = client.CreateReview(ctx, requestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, err, tt.access, "CreateReview")
			_, err = client.UpdateReview(ctx, requestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateReview")
			_, err = client.GetReviewers(ctx, requestWithCookie(&qf.SubmissionReviewersRequest{
				CourseID:     tt.courseID,
				SubmissionID: 1,
			}, tt.cookie))
			checkAccess(t, err, tt.access, "GetReviewers")
			_, err = client.IsEmptyRepo(ctx, requestWithCookie(&qf.RepositoryRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, err, tt.access, "IsEmptyRepo")
		})
	}

	courseAdminTests := accessTests{
		{"admin, not enrolled", adminCookie, 0, course.ID, 0, false},
	}
	for _, tt := range courseAdminTests {
		t.Run("CourseAdminAccess/"+tt.name, func(t *testing.T) {
			_, err = client.GetSubmissionsByCourse(ctx, requestWithCookie(&qf.SubmissionsForCourseRequest{
				CourseID: tt.courseID,
			}, tt.cookie))
			checkAccess(t, err, tt.access, "GetSubmissionsByCourse")
		})
	}

	adminAccessTests := accessTests{
		{"admin (accessing own info)", courseAdminCookie, courseAdmin.ID, course.ID, group.ID, true},
		{"admin (accessing other user's info)", courseAdminCookie, user.ID, course.ID, group.ID, true},
		{"non admin (accessing admin's info)", studentCookie, courseAdmin.ID, course.ID, group.ID, false},
		{"non admin (accessing other user's info)", studentCookie, user.ID, course.ID, group.ID, false},
	}
	for _, tt := range adminAccessTests {
		t.Run("AdminAccess/"+tt.name, func(t *testing.T) {
			_, err := client.UpdateUser(ctx, requestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateUser")
			_, err = client.GetEnrollmentsByUser(ctx, requestWithCookie(&qf.EnrollmentStatusRequest{UserID: tt.userID}, tt.cookie))
			checkAccess(t, err, tt.access, "GetEnrollmentsByUser")
			_, err = client.GetUsers(ctx, requestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, err, tt.access, "GetUsers")
			_, err = client.GetOrganization(ctx, requestWithCookie(&qf.OrgRequest{OrgName: "testorg"}, tt.cookie))
			checkAccess(t, err, tt.access, "GetOrganization")
			_, err = client.CreateCourse(ctx, requestWithCookie(course, tt.cookie))
			checkAccess(t, err, tt.access, "CreateCourse")
			_, err = client.GetUserByCourse(ctx, requestWithCookie(&qf.CourseUserRequest{
				CourseCode: course.Code,
				CourseYear: course.Year,
				UserLogin:  "student",
			}, tt.cookie))
			checkAccess(t, err, tt.access, "GetUserByCourse")
		})
	}

	createGroupTests := []struct {
		name   string
		cookie string
		group  *qf.Group
		access bool
	}{
		{"valid student, not in the request group", studentCookie, &qf.Group{
			CourseID: course.ID,
		}, false},
		{"valid student", studentCookie, &qf.Group{
			Name:     "test",
			CourseID: course.ID,
			Users:    []*qf.User{student},
		}, true},
		{"course teacher", courseAdminCookie, &qf.Group{
			CourseID: course.ID,
			Users:    []*qf.User{courseAdmin},
		}, true},
		{"admin, not a teacher", adminCookie, &qf.Group{
			CourseID: course.ID,
		}, false},
	}

	for _, tt := range createGroupTests {
		t.Run("CreateGroupAccess/"+tt.name, func(t *testing.T) {
			_, err := client.CreateGroup(ctx, requestWithCookie(tt.group, tt.cookie))
			checkAccess(t, err, tt.access, "CreateGroup")
		})
	}

	adminStatusChangeTests := []struct {
		name   string
		cookie string
		user   *qf.User
		access bool
	}{
		{"admin demoting a user", courseAdminCookie, &qf.User{
			ID:      admin.ID,
			IsAdmin: false,
		}, true},
		{"admin promoting a user", courseAdminCookie, &qf.User{
			ID:      admin.ID,
			IsAdmin: true,
		}, true},
		{"admin demoting self", courseAdminCookie, &qf.User{
			ID:      courseAdmin.ID,
			IsAdmin: false,
		}, true},
		{"user promoting another user", userCookie, &qf.User{
			ID:      groupStudent.ID,
			IsAdmin: true,
		}, false},
		{"user promoting self", userCookie, &qf.User{
			ID:      user.ID,
			IsAdmin: true,
		}, false},
	}

	for _, tt := range adminStatusChangeTests {
		t.Run("AdminStatusChange/"+tt.name, func(t *testing.T) {
			_, err := client.UpdateUser(ctx, requestWithCookie(tt.user, tt.cookie))
			checkAccess(t, err, tt.access, "UpdateUser")
		})
	}

	if err = muxServer.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}

func requestWithCookie[T any](message *T, cookie string) *connect.Request[T] {
	request := connect.NewRequest(message)
	request.Header().Set(auth.Cookie, cookie)
	return request
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
