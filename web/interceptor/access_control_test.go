package interceptor_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
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
	client := web.MockClient(t, db, scm.WithMockOrgs(), connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
		interceptor.NewAccessControlInterceptor(tm),
	))
	ctx := context.Background()

	courseAdmin := qtest.CreateFakeUser(t, db)
	groupStudent := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "group student", Login: "group student"})
	student := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student", Login: "student"})
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "user", Login: "user"})
	admin := qtest.CreateFakeUser(t, db)
	admin.IsAdmin = true
	if err := db.UpdateUser(admin); err != nil {
		t.Fatal(err)
	}

	course := &qf.Course{
		Code:                "test101",
		Year:                2022,
		ScmOrganizationID:   1,
		ScmOrganizationName: "test",
		CourseCreatorID:     courseAdmin.ID,
	}
	qtest.CreateCourse(t, db, courseAdmin, course)
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
			_, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourse(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourses(ctx, qtest.RequestWithCookie(&qf.Void{}, tt.cookie))
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
			enrolRequest := &qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_UserID{
					UserID: tt.userID,
				},
			}
			_, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(enrol, tt.cookie))
			checkAccess(t, "CreateEnrollment", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourseVisibility(ctx, qtest.RequestWithCookie(enrol, tt.cookie))
			checkAccess(t, "UpdateCourseVisibility", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(ctx, qtest.RequestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollments(ctx, qtest.RequestWithCookie(enrolRequest, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(ctx, qtest.RequestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
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
			_, err := client.GetSubmissions(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
				CourseID: tt.courseID,
				FetchMode: &qf.SubmissionRequest_UserID{
					UserID: tt.userID,
				},
			}, tt.cookie))
			checkAccess(t, "GetSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetAssignments(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollments(ctx, qtest.RequestWithCookie(&qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_UserID{
					UserID: tt.userID,
				},
			}, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetRepositories", err, tt.wantCode, tt.wantAccess)
		})
	}

	groupAccessTests := map[string]accessTest{
		"student in a group":                            {cookie: groupStudentCookie, userID: groupStudent.ID, courseID: course.ID, groupID: group.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
		"student, not in a group":                       {cookie: studentCookie, userID: student.ID, courseID: course.ID, groupID: group.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"student in a group, wrong group ID in request": {cookie: studentCookie, userID: student.ID, courseID: course.ID, groupID: 123, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range groupAccessTests {
		t.Run("GroupAccess/"+name, func(t *testing.T) {
			_, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				GroupID:  tt.groupID,
			}, tt.cookie))
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
			_, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				GroupID:  tt.groupID,
			}, tt.cookie))
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}, tt.cookie))
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				GroupID:  tt.groupID,
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}, tt.cookie))
			checkAccess(t, "DeleteGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateGroup(ctx, qtest.RequestWithCookie(&qf.Group{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourse(ctx, qtest.RequestWithCookie(course, tt.cookie))
			checkAccess(t, "UpdateCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
				Enrollments: []*qf.Enrollment{{ID: 1, CourseID: tt.courseID}},
			}, tt.cookie))
			checkAccess(t, "UpdateEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateAssignments(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateSubmission(ctx, qtest.RequestWithCookie(&qf.UpdateSubmissionRequest{SubmissionID: 1, CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateSubmission", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateSubmissions(ctx, qtest.RequestWithCookie(&qf.UpdateSubmissionsRequest{AssignmentID: 1, CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.RebuildSubmissions(ctx, qtest.RequestWithCookie(&qf.RebuildRequest{
				AssignmentID: 1,
				CourseID:     tt.courseID,
			}, tt.cookie))
			checkAccess(t, "RebuildSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateBenchmark(ctx, qtest.RequestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, "CreateBenchmark", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateBenchmark(ctx, qtest.RequestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, "UpdateBenchmark", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteBenchmark(ctx, qtest.RequestWithCookie(&qf.GradingBenchmark{CourseID: tt.courseID, AssignmentID: 1}, tt.cookie))
			checkAccess(t, "DeleteBenchmark", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateCriterion(ctx, qtest.RequestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, "CreateCriterion", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCriterion(ctx, qtest.RequestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, "UpdateCriterion", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteCriterion(ctx, qtest.RequestWithCookie(&qf.GradingCriterion{CourseID: tt.courseID, BenchmarkID: 1}, tt.cookie))
			checkAccess(t, "DeleteCriterion", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateReview(ctx, qtest.RequestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, "CreateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateReview(ctx, qtest.RequestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, "UpdateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.IsEmptyRepo(ctx, qtest.RequestWithCookie(&qf.RepositoryRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "IsEmptyRepo", err, tt.wantCode, tt.wantAccess)
		})
	}

	courseAdminTests := map[string]accessTest{
		"admin, not enrolled": {cookie: adminCookie, courseID: course.ID, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range courseAdminTests {
		t.Run("CourseAdminAccess/"+name, func(t *testing.T) {
			_, err = client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
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
			_, err := client.UpdateUser(ctx, qtest.RequestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetUsers(ctx, qtest.RequestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetUsers", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetOrganization(ctx, qtest.RequestWithCookie(&qf.Organization{ScmOrganizationName: "test"}, tt.cookie))
			checkAccess(t, "GetOrganization", err, tt.wantCode, tt.wantAccess)
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
			_, err := client.CreateGroup(ctx, qtest.RequestWithCookie(tt.group, tt.cookie))
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
			_, err := client.UpdateUser(ctx, qtest.RequestWithCookie(tt.user, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminGetEnrollmentsTests := map[string]accessTest{
		"admin, not enrolled in the course": {cookie: adminCookie, courseID: course.ID, userID: student.ID, wantAccess: true, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range adminGetEnrollmentsTests {
		t.Run("AdminGetEnrollments/"+name, func(t *testing.T) {
			_, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(&qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_CourseID{
					CourseID: tt.courseID,
				},
			}, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollments(ctx, qtest.RequestWithCookie(&qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_UserID{
					UserID: tt.userID,
				},
			}, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
		})
	}
}

func checkAccess(t *testing.T, method string, err error, wantCode connect.Code, wantAccess bool) {
	t.Helper()
	var connErr *connect.Error
	if errors.As(err, &connErr) {
		gotCode := connErr.Code()
		gotAccess := gotCode == wantCode
		if gotAccess == wantAccess {
			t.Errorf("%23s: (%v == %v) = %t, want %t", method, gotCode, wantCode, gotAccess, !wantAccess)
			t.Log(err)
		}
	} else if err != nil && wantAccess {
		// got error and want access; expected non-error or not access
		t.Errorf("%23s: got %v (%t), want <nil> (%t)", method, err, wantAccess, !wantAccess)
	}
}
