package interceptor_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestAccessControlMethods(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, &scm.Manager{}, web.BaseHookOptions{}, &ci.Local{})

	s := grpc.NewServer()
	qf.RegisterQuickFeedServiceServer(s, ags)

	access := interceptor.GetAccessTable()
	qfServiceInfo, ok := s.GetServiceInfo()[web.QuickFeedServiceName]
	if !ok {
		t.Fatalf("failed to read service info (%s)", web.QuickFeedServiceName)
	}

	for _, method := range qfServiceInfo.Methods {
		_, ok := access[method.Name]
		if !ok {
			t.Errorf("access control table missing method %s", method.Name)
		}
	}
}

func TestAccessControl(t *testing.T) {
	// TODO(vera): refactor grpc and database setup and each group of test cases into separate methods.
	const (
		bufSize = 1024 * 1024
	)

	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, &scm.Manager{}, web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db, "test")
	if err != nil {
		t.Fatal(err)
	}

	lis := bufconn.Listen(bufSize)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	opt := grpc.ChainUnaryInterceptor(
		interceptor.UnaryUserVerifier(logger, tm),
		interceptor.AccessControl(logger, tm),
	)
	s := grpc.NewServer(opt)
	qf.RegisterQuickFeedServiceServer(s, ags)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := qf.NewQuickFeedServiceClient(conn)

	admin := qtest.CreateAdminUser(t, db, "fake")
	groupStudent := qtest.CreateNamedUser(t, db, 2, "group student")
	student := qtest.CreateNamedUser(t, db, 3, "student")
	user := qtest.CreateNamedUser(t, db, 4, "user")

	course := &qf.Course{
		Code:             "test101",
		Year:             2022,
		Provider:         "fake",
		OrganizationID:   1,
		OrganizationPath: "test",
	}
	if err := db.CreateCourse(admin.ID, course); err != nil {
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

	adminToken, err := tm.NewAuthCookie(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	studentToken, err := tm.NewAuthCookie(groupStudent.ID)
	if err != nil {
		t.Fatal(err)
	}
	userToken, err := tm.NewAuthCookie(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	adminContext := qtest.WithAuthCookie(context.Background(), adminToken.Value)
	studentContext := qtest.WithAuthCookie(context.Background(), studentToken.Value)
	userContext := qtest.WithAuthCookie(context.Background(), userToken.Value)

	type accessTests []struct {
		name     string
		ctx      context.Context
		userID   uint64
		courseID uint64
		groupID  uint64
		err      error
	}

	freeAccessTest := accessTests{
		{"admin", adminContext, 0, course.ID, 0, nil},
		{"student", studentContext, 0, course.ID, 0, nil},
		{"user", userContext, 0, course.ID, 0, nil},
	}

	for _, testCase := range freeAccessTest {
		if _, err := client.GetUser(testCase.ctx, &qf.Void{}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetUser", err, testCase.err)
		}
		if _, err := client.GetCourse(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetCourse", err, testCase.err)
		}
		if _, err := client.GetCourses(testCase.ctx, &qf.Void{}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetCourses", err, testCase.err)
		}
	}

	userAccessTests := accessTests{
		{"correct user ID", userContext, user.ID, course.ID, 0, nil},
		{"incorrect user ID", userContext, groupStudent.ID, course.ID, 0, interceptor.ErrAccessDenied},
	}
	for _, testCase := range userAccessTests {
		enrol := &qf.Enrollment{
			CourseID: course.ID,
			UserID:   testCase.userID,
		}
		enrolRequest := &qf.EnrollmentStatusRequest{
			UserID: testCase.userID,
		}
		if _, err := client.CreateEnrollment(testCase.ctx, enrol); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "CreateEnrollment", err, testCase.err)
		}
		if _, err := client.UpdateCourseVisibility(testCase.ctx, enrol); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateCourseVisibility", err, testCase.err)
		}
		if _, err := client.GetCoursesByUser(testCase.ctx, enrolRequest); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetCoursesByUser", err, testCase.err)
		}
		if _, err := client.UpdateUser(testCase.ctx, &qf.User{ID: testCase.userID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateUser", err, testCase.err)
		}
		if _, err := client.GetEnrollmentsByUser(testCase.ctx, enrolRequest); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetEnrollmentsByUser", err, testCase.err)
		}
	}

	adminAccessTests := accessTests{
		{"admin (accessing own info)", adminContext, admin.ID, course.ID, group.ID, nil},
		{"admin (accessing other user's info)", adminContext, user.ID, course.ID, group.ID, nil},
		{"non admin (accessing admin's info)", studentContext, admin.ID, course.ID, group.ID, interceptor.ErrAccessDenied},
		{"non admin (accessing other user's info)", studentContext, user.ID, course.ID, group.ID, interceptor.ErrAccessDenied},
	}

	for _, testCase := range adminAccessTests {
		if _, err := client.UpdateUser(testCase.ctx, &qf.User{ID: testCase.userID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateUser", err, testCase.err)
		}
		if _, err := client.GetEnrollmentsByUser(testCase.ctx, &qf.EnrollmentStatusRequest{UserID: testCase.userID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetEnrollmentsByUser", err, testCase.err)
		}
		if _, err := client.GetUsers(testCase.ctx, &qf.Void{}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetUsers", err, testCase.err)
		}
		if _, err := client.GetOrganization(testCase.ctx, &qf.OrgRequest{OrgName: scm.GetTestOrganization(t)}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetOrganization", err, testCase.err)
		}
		if _, err := client.CreateCourse(testCase.ctx, course); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateUser", err, testCase.err)
		}
		if _, err := client.GetUserByCourse(adminContext, &qf.CourseUserRequest{
			CourseCode: course.Code,
			CourseYear: course.Year,
			UserLogin:  admin.Login,
		}); !errors.Is(err, interceptor.ErrAccessDenied) {
			logError(t, testCase.name, "GetUserByCourse", err, testCase.err)
		}
		if _, err := client.GetUserByCourse(adminContext, &qf.CourseUserRequest{
			CourseCode: course.Code,
			CourseYear: course.Year,
			UserLogin:  student.Login,
		}); !errors.Is(err, interceptor.ErrAccessDenied) {
			logError(t, testCase.name, "GetUserByCourse", err, testCase.err)
		}
	}
}

func logError(t *testing.T, name, method string, got, want error) {
	t.Errorf("unexpected error for %s calling %s: expected %v, got %v", name, method, want, got)
}
