package interceptor_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

// TestLoggingInterceptor tests that the logging interceptor logs the correct information.
// Must run with LOG=1 to see the logs:
//
//	LOG=1 go test -v -run TestLoggingInterceptor
func TestLoggingInterceptor(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	client := web.MockClient(t, db, connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
		interceptor.NewAccessControlInterceptor(tm),
		interceptor.NewContextLoggingInterceptor(logger.Desugar(), db),
	))
	ctx := context.Background()

	courseAdmin := qtest.CreateNamedUser(t, db, 1, "course admin")
	groupStudent := qtest.CreateNamedUser(t, db, 2, "group student")
	student := qtest.CreateNamedUser(t, db, 3, "student")
	user := qtest.CreateNamedUser(t, db, 4, "user")
	admin := qtest.CreateNamedUser(t, db, 6, "admin")
	admin.IsAdmin = true
	if err := db.UpdateUser(admin); err != nil {
		t.Fatal(err)
	}

	course := &qf.Course{
		Code:                "DAT101",
		Year:                2024,
		ScmOrganizationID:   1,
		ScmOrganizationName: "test",
		CourseCreatorID:     courseAdmin.ID,
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
		})
	}
}
