package interceptor_test

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

type accessTest struct {
	cookie     string
	userID     uint64
	courseID   uint64
	groupID    uint64
	wantAccess bool
	wantCode   connect.Code
}

func TestAccessControl(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	serveFn, shutdown := web.MockQuickFeedServer(t, logger, db, connect.WithInterceptors(
		interceptor.UnaryUserVerifier(logger, tm),
		interceptor.AccessControl(tm),
	))
	go serveFn()

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
		OrganizationPath: "test",
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
		return cookie.String()
	}
	courseAdminCookie := f(t, courseAdmin.ID)
	groupStudentCookie := f(t, groupStudent.ID)
	studentCookie := f(t, student.ID)
	userCookie := f(t, user.ID)
	adminCookie := f(t, admin.ID)

	freeAccessTest := map[string]accessTest{
		"admin":             {cookie: courseAdminCookie, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"student":           {cookie: studentCookie, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"group student":     {cookie: groupStudentCookie, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"user":              {cookie: userCookie, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"non-teacher admin": {cookie: adminCookie, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"empty context":     {wantAccess: false, wantCode: connect.CodeUnauthenticated},
	}
	for name, tt := range freeAccessTest {
		t.Run("UnrestrictedAccess/"+name, func(t *testing.T) {
			_, err := client.GetUser(ctx, requestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourse(ctx, requestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourses(ctx, requestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetCourses", err, tt.wantCode, tt.wantAccess)
		})
	}

	userAccessTests := map[string]accessTest{
		"correct user ID":   {cookie: userCookie, userID: user.ID, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"incorrect user ID": {cookie: userCookie, groupID: groupStudent.ID, courseID: course.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range userAccessTests {
		t.Run("UserAccess/"+name, func(t *testing.T) {
			enrol := &qf.Enrollment{
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}
			enrolRequest := &qf.EnrollmentStatusRequest{
				UserID: tt.userID,
			}
			_, err := client.CreateEnrollment(ctx, requestWithCookie(enrol, tt.cookie))
			checkAccess(t, "CreateEnrollment", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourseVisibility(ctx, requestWithCookie(enrol, tt.cookie))
			checkAccess(t, "UpdateCourseVisibility", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCoursesByUser(ctx, requestWithCookie(enrolRequest, tt.cookie))
			checkAccess(t, "GetCoursesByUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(ctx, requestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollmentsByUser(ctx, requestWithCookie(enrolRequest, tt.cookie))
			checkAccess(t, "GetEnrollmentsByCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(ctx, requestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
		})
	}

	studentAccessTests := map[string]accessTest{
		"course admin":                     {cookie: courseAdminCookie, userID: courseAdmin.ID, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"admin, not enrolled in a course":  {cookie: adminCookie, userID: admin.ID, courseID: course.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"user, not enrolled in the course": {cookie: userCookie, userID: user.ID, courseID: course.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"student":                          {cookie: studentCookie, userID: student.ID, courseID: course.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"student of another course":        {cookie: studentCookie, userID: student.ID, courseID: 123, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range studentAccessTests {
		t.Run("StudentAccess/"+name, func(t *testing.T) {
			_, err := client.GetSubmissions(ctx, requestWithCookie(&qf.SubmissionRequest{
				UserID:   tt.userID,
				CourseID: tt.courseID,
			}, tt.cookie))
			checkAccess(t, "GetSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetAssignments(ctx, requestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollmentsByCourse(ctx, requestWithCookie(&qf.EnrollmentRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetEnrollmentsByCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetRepositories(ctx, requestWithCookie(&qf.URLRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetRepositories", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetGroupByUserAndCourse(ctx, requestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				UserID:   tt.userID,
				GroupID:  0,
			}, tt.cookie))
			checkAccess(t, "GetGroupByUserAndCourse", err, tt.wantCode, tt.wantAccess)
		})
	}

	groupAccessTests := map[string]accessTest{
		"student in a group":                            {cookie: groupStudentCookie, userID: groupStudent.ID, courseID: course.ID, groupID: group.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"student, not in a group":                       {cookie: studentCookie, userID: student.ID, courseID: course.ID, groupID: group.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"student in a group, wrong group ID in request": {cookie: studentCookie, userID: student.ID, courseID: course.ID, groupID: 123, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range groupAccessTests {
		t.Run("GroupAccess/"+name, func(t *testing.T) {
			_, err := client.GetGroup(ctx, requestWithCookie(&qf.GetGroupRequest{GroupID: tt.groupID}, tt.cookie))
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
		})
	}

	teacherAccessTests := map[string]accessTest{
		"course teacher":                    {cookie: courseAdminCookie, userID: groupStudent.ID, courseID: course.ID, groupID: group.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"student":                           {cookie: studentCookie, userID: student.ID, courseID: course.ID, groupID: group.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"admin, not enrolled in the course": {cookie: adminCookie, userID: admin.ID, courseID: course.ID, groupID: group.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range teacherAccessTests {
		t.Run("TeacherAccess/"+name, func(t *testing.T) {
			_, err := client.GetGroup(ctx, requestWithCookie(&qf.GetGroupRequest{GroupID: tt.groupID}, tt.cookie))
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteGroup(ctx, requestWithCookie(&qf.GroupRequest{
				GroupID:  tt.groupID,
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}, tt.cookie))
			checkAccess(t, "DeleteGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateGroup(ctx, requestWithCookie(&qf.Group{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourse(ctx, requestWithCookie(course, tt.cookie))
			checkAccess(t, "UpdateCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateEnrollments(ctx, requestWithCookie(&qf.Enrollments{
				Enrollments: []*qf.Enrollment{{ID: 1, CourseID: tt.courseID}},
			}, tt.cookie))
			checkAccess(t, "UpdateEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateAssignments(ctx, requestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateSubmission(ctx, requestWithCookie(&qf.UpdateSubmissionRequest{SubmissionID: 1, CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateSubmission", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateSubmissions(ctx, requestWithCookie(&qf.UpdateSubmissionsRequest{AssignmentID: 1, CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.RebuildSubmissions(ctx, requestWithCookie(&qf.RebuildRequest{
				AssignmentID: 1,
				CourseID:     tt.courseID,
			}, tt.cookie))
			checkAccess(t, "RebuildSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateBenchmark(ctx, requestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, "CreateBenchmark", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateBenchmark(ctx, requestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, "UpdateBenchmark", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteBenchmark(ctx, requestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, "DeleteBenchmark", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateCriterion(ctx, requestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, "CreateCriterion", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCriterion(ctx, requestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, "UpdateCriterion", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteCriterion(ctx, requestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, "DeleteCriterion", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateReview(ctx, requestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, "CreateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateReview(ctx, requestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, "UpdateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetReviewers(ctx, requestWithCookie(&qf.SubmissionReviewersRequest{
				CourseID:     tt.courseID,
				SubmissionID: 1,
			}, tt.cookie))
			checkAccess(t, "GetReviewers", err, tt.wantCode, tt.wantAccess)
			_, err = client.IsEmptyRepo(ctx, requestWithCookie(&qf.RepositoryRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "IsEmptyRepo", err, tt.wantCode, tt.wantAccess)
		})
	}

	courseAdminTests := map[string]accessTest{
		"admin, not enrolled": {cookie: adminCookie, courseID: course.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range courseAdminTests {
		t.Run("CourseAdminAccess/"+name, func(t *testing.T) {
			_, err = client.GetSubmissionsByCourse(ctx, requestWithCookie(&qf.SubmissionsForCourseRequest{
				CourseID: tt.courseID,
			}, tt.cookie))
			checkAccess(t, "GetSubmissionsByCourse", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminAccessTests := map[string]accessTest{
		"admin (accessing own info)":              {cookie: courseAdminCookie, userID: courseAdmin.ID, courseID: course.ID, groupID: group.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"admin (accessing other user's info)":     {cookie: courseAdminCookie, userID: user.ID, courseID: course.ID, groupID: group.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"non admin (accessing admin's info)":      {cookie: studentCookie, userID: courseAdmin.ID, courseID: course.ID, groupID: group.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"non admin (accessing other user's info)": {cookie: studentCookie, userID: user.ID, courseID: course.ID, groupID: group.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range adminAccessTests {
		t.Run("AdminAccess/"+name, func(t *testing.T) {
			_, err := client.UpdateUser(ctx, requestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollmentsByUser(ctx, requestWithCookie(&qf.EnrollmentStatusRequest{UserID: tt.userID}, tt.cookie))
			checkAccess(t, "GetEnrollmentsByUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetUsers(ctx, requestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetUsers", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetOrganization(ctx, requestWithCookie(&qf.OrgRequest{OrgName: "test"}, tt.cookie))
			checkAccess(t, "GetOrganization", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateCourse(ctx, requestWithCookie(course, tt.cookie))
			checkAccess(t, "CreateCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetUserByCourse(ctx, requestWithCookie(&qf.CourseUserRequest{
				CourseCode: course.Code,
				CourseYear: course.Year,
				UserLogin:  "student",
			}, tt.cookie))
			checkAccess(t, "GetUserByCourse", err, tt.wantCode, tt.wantAccess)
		})
	}

	createGroupTests := map[string]struct {
		cookie     string
		group      *qf.Group
		wantAccess bool
		wantCode   connect.Code
	}{
		"valid student, not in the request group": {cookie: studentCookie, group: &qf.Group{
			CourseID: course.ID,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"valid student": {cookie: studentCookie, group: &qf.Group{
			Name:     "test",
			CourseID: course.ID,
			Users:    []*qf.User{student},
		}, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"course teacher": {cookie: courseAdminCookie, group: &qf.Group{
			CourseID: course.ID,
			Users:    []*qf.User{courseAdmin},
		}, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"admin, not a teacher": {cookie: adminCookie, group: &qf.Group{
			CourseID: course.ID,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range createGroupTests {
		t.Run("CreateGroupAccess/"+name, func(t *testing.T) {
			_, err := client.CreateGroup(ctx, requestWithCookie(tt.group, tt.cookie))
			checkAccess(t, "CreateGroup", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminStatusChangeTests := map[string]struct {
		cookie     string
		user       *qf.User
		wantAccess bool
		wantCode   connect.Code
	}{
		"admin demoting a user": {cookie: courseAdminCookie, user: &qf.User{
			ID:      admin.ID,
			IsAdmin: false,
		}, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"admin promoting a user": {cookie: courseAdminCookie, user: &qf.User{
			ID:      admin.ID,
			IsAdmin: true,
		}, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"admin demoting self": {cookie: courseAdminCookie, user: &qf.User{
			ID:      courseAdmin.ID,
			IsAdmin: false,
		}, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"user promoting another user": {cookie: userCookie, user: &qf.User{
			ID:      groupStudent.ID,
			IsAdmin: true,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"user promoting self": {cookie: userCookie, user: &qf.User{
			ID:      user.ID,
			IsAdmin: true,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range adminStatusChangeTests {
		t.Run("AdminStatusChange/"+name, func(t *testing.T) {
			_, err := client.UpdateUser(ctx, requestWithCookie(tt.user, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
		})
	}
	shutdown(ctx)
}

func requestWithCookie[T any](message *T, cookie string) *connect.Request[T] {
	request := connect.NewRequest(message)
	request.Header().Set(auth.Cookie, cookie)
	return request
}

func checkAccess(t *testing.T, method string, err error, wantCode connect.Code, wantAccess bool) {
	t.Helper()
	if connErr, ok := err.(*connect.Error); ok {
		gotCode := connErr.Code()
		gotAccess := gotCode == wantCode
		if gotAccess == wantAccess {
			t.Errorf("%23s: (%v == %v) = %t, want %t", method, gotCode, wantCode, gotAccess, !wantAccess)
		}
	} else if err != nil && wantAccess {
		// got error and want access; expected non-error or not access
		t.Errorf("%23s: got %v (%t), want <nil> (%t)", method, err, wantAccess, !wantAccess)
	}
}
