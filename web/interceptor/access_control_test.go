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

const (
	bufSize = 1024 * 1024
)

type accessTests []struct {
	name     string
	ctx      context.Context
	userID   uint64
	courseID uint64
	groupID  uint64
	err      error
}

func TestAccessControlMethods(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, &scm.Manager{}, web.BaseHookOptions{}, &ci.Local{})

	s := grpc.NewServer()
	qf.RegisterQuickFeedServiceServer(s, ags)
	if err := web.VerifyAccessControlMethods(s); err != nil {
		t.Error(err)
	}
}

func TestAccessControl(t *testing.T) {
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

	courseAdmin := qtest.CreateAdminUser(t, db, "fake")
	groupStudent := qtest.CreateNamedUser(t, db, 2, "group student")
	student := qtest.CreateNamedUser(t, db, 3, "student")
	user := qtest.CreateNamedUser(t, db, 4, "user")
	admin := qtest.CreateFakeUser(t, db, 5)
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

	courseAdminToken, err := tm.NewAuthCookie(courseAdmin.ID)
	if err != nil {
		t.Fatal(err)
	}
	groupStudentToken, err := tm.NewAuthCookie(groupStudent.ID)
	if err != nil {
		t.Fatal(err)
	}
	studentToken, err := tm.NewAuthCookie(student.ID)
	if err != nil {
		t.Fatal(err)
	}
	userToken, err := tm.NewAuthCookie(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	adminToken, err := tm.NewAuthCookie(admin.ID)
	if err != nil {
		t.Fatal(err)
	}

	courseAdminContext := qtest.WithAuthCookie(ctx, courseAdminToken.Value)
	groupStudentContext := qtest.WithAuthCookie(ctx, groupStudentToken.Value)
	studentContext := qtest.WithAuthCookie(ctx, studentToken.Value)
	userContext := qtest.WithAuthCookie(ctx, userToken.Value)
	adminContext := qtest.WithAuthCookie(ctx, adminToken.Value)

	freeAccessTest := accessTests{
		{"admin", courseAdminContext, 0, course.ID, 0, nil},
		{"student", studentContext, 0, course.ID, 0, nil},
		{"student", groupStudentContext, 0, course.ID, 0, nil},
		{"user", userContext, 0, course.ID, 0, nil},
		{"non-teacher admin", adminContext, 0, course.ID, 0, nil},
	}
	testUnrestrictedAccess(client, t, freeAccessTest)

	userAccessTests := accessTests{
		{"correct user ID", userContext, user.ID, course.ID, 0, nil},
		{"incorrect user ID", userContext, groupStudent.ID, course.ID, 0, interceptor.ErrAccessDenied},
	}
	testUserAccess(client, t, course, userAccessTests)

	studentAccessTests := accessTests{
		{"course admin", courseAdminContext, courseAdmin.ID, course.ID, 0, nil},
		{"admin, not enrolled in a course", adminContext, admin.ID, course.ID, 0, interceptor.ErrAccessDenied},
		{"user, not enrolled in the course", userContext, user.ID, course.ID, 0, interceptor.ErrAccessDenied},
		{"student", studentContext, student.ID, course.ID, 0, nil},
		{"student of another course", studentContext, student.ID, 123, 0, interceptor.ErrAccessDenied},
	}
	testStudentAccess(client, t, studentAccessTests)

	groupAccessTests := accessTests{
		{"student in a group", groupStudentContext, groupStudent.ID, course.ID, group.ID, nil},
		{"student, not in a group", studentContext, student.ID, course.ID, group.ID, interceptor.ErrAccessDenied},
		{"student in a group, wrong group ID in request", studentContext, student.ID, course.ID, 123, interceptor.ErrAccessDenied},
	}
	testGroupAccess(client, t, groupAccessTests)

	adminAccessTests := accessTests{
		{"admin (accessing own info)", courseAdminContext, courseAdmin.ID, course.ID, group.ID, nil},
		{"admin (accessing other user's info)", courseAdminContext, user.ID, course.ID, group.ID, nil},
		{"non admin (accessing admin's info)", studentContext, courseAdmin.ID, course.ID, group.ID, interceptor.ErrAccessDenied},
		{"non admin (accessing other user's info)", studentContext, user.ID, course.ID, group.ID, interceptor.ErrAccessDenied},
	}
	testAdminAccess(client, t, course, adminAccessTests)

	teacherAccessTests := accessTests{
		{"course teacher", courseAdminContext, courseAdmin.ID, course.ID, group.ID, nil},
		{"student", studentContext, student.ID, course.ID, group.ID, interceptor.ErrAccessDenied},
		{"admin, not enrolled in the course", adminContext, admin.ID, course.ID, group.ID, interceptor.ErrAccessDenied},
	}
	testTeacherAccess(client, t, teacherAccessTests)

}

func testTeacherAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
		if _, err := client.GetGroupByUserAndCourse(testCase.ctx, &qf.GroupRequest{
			GroupID:  testCase.groupID,
			CourseID: testCase.courseID,
		}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetGroupByUserAndCourse", err, testCase.err)
		}
		if _, err := client.GetGroup(testCase.ctx, &qf.GetGroupRequest{GroupID: testCase.groupID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetGroup", err, testCase.err)
		}
		if _, err := client.GetAssignments(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetAssignments", err, testCase.err)
		}
		if _, err := client.GetEnrollmentsByCourse(testCase.ctx, &qf.EnrollmentRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetEnrollmentsByCourse", err, testCase.err)
		}
		if _, err := client.GetRepositories(testCase.ctx, &qf.URLRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetRepositories", err, testCase.err)
		}
		if _, err := client.UpdateGroup(testCase.ctx, &qf.Group{ID: testCase.groupID, CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateGroup", err, testCase.err)
		}
		if _, err := client.DeleteGroup(testCase.ctx, &qf.GroupRequest{GroupID: testCase.groupID, CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "DeleteGroup", err, testCase.err)
		}
		if _, err := client.GetGroupsByCourse(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetGroupsByCourse", err, testCase.err)
		}
		if _, err := client.UpdateCourse(testCase.ctx, &qf.Course{ID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateCourse", err, testCase.err)
		}
		if _, err := client.UpdateEnrollments(testCase.ctx, &qf.Enrollments{
			Enrollments: []*qf.Enrollment{{ID: 1}},
		}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateEnrollments", err, testCase.err)
		}
		if _, err := client.UpdateAssignments(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateAssignments", err, testCase.err)
		}
		if _, err := client.UpdateSubmission(testCase.ctx, &qf.UpdateSubmissionRequest{SubmissionID: 1, CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateSubmission", err, testCase.err)
		}
		if _, err := client.UpdateSubmissions(testCase.ctx, &qf.UpdateSubmissionsRequest{AssignmentID: 1, CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateSubmissions", err, testCase.err)
		}
		if _, err := client.RebuildSubmissions(testCase.ctx, &qf.RebuildRequest{
			AssignmentID: 1,
			RebuildType: &qf.RebuildRequest_CourseID{
				CourseID: testCase.courseID,
			}}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "RebuildSubmissions", err, testCase.err)
		}
		if _, err := client.CreateBenchmark(testCase.ctx, &qf.GradingBenchmark{AssignmentID: 1}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "CreateBenchmark", err, testCase.err)
		}
		if _, err := client.UpdateBenchmark(testCase.ctx, &qf.GradingBenchmark{AssignmentID: 1}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateBenchmark", err, testCase.err)
		}
		if _, err := client.DeleteBenchmark(testCase.ctx, &qf.GradingBenchmark{AssignmentID: 1}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "DeleteBenchmark", err, testCase.err)
		}
		if _, err := client.CreateCriterion(testCase.ctx, &qf.GradingCriterion{BenchmarkID: 1}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "CreateCriterion", err, testCase.err)
		}
		if _, err := client.UpdateCriterion(testCase.ctx, &qf.GradingCriterion{BenchmarkID: 1}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateCriterion", err, testCase.err)
		}
		if _, err := client.DeleteCriterion(testCase.ctx, &qf.GradingCriterion{BenchmarkID: 1}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "DeleteCriterion", err, testCase.err)
		}
		if _, err := client.CreateReview(testCase.ctx, &qf.ReviewRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "CreateReview", err, testCase.err)
		}
		if _, err := client.UpdateReview(testCase.ctx, &qf.ReviewRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "UpdateReview", err, testCase.err)
		}
		if _, err := client.GetReviewers(testCase.ctx, &qf.SubmissionReviewersRequest{
			CourseID:     testCase.courseID,
			SubmissionID: 1,
		}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetReviewers", err, testCase.err)
		}
		if _, err := client.IsEmptyRepo(testCase.ctx, &qf.RepositoryRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "IsEmptyRepo", err, testCase.err)
		}
	}
}

func testGroupAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
		if _, err := client.GetGroupByUserAndCourse(testCase.ctx, &qf.GroupRequest{
			CourseID: testCase.courseID,
			UserID:   testCase.userID,
			GroupID:  testCase.groupID,
		}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetGroupByUserAndCourse", err, testCase.err)
		}
		if _, err := client.GetGroup(testCase.ctx, &qf.GetGroupRequest{GroupID: testCase.groupID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetGroup", err, testCase.err)
		}

		// TODO(vera): CreateGroup needs a separate set of test cases
		// to account for backend errors (user who already in a group will pass
		// access control but the request will result in error because already in a group)
		// if _, err := client.CreateGroup(testCase.ctx, &qf.Group{
		// 	CourseID: testCase.courseID,
		// 	Users: []*qf.User{
		// 		{
		// 			ID: testCase.userID,
		// 		},
		// 	},
		// }); !errors.Is(err, testCase.err) {
		// 	logError(t, testCase.name, "CreateGroup", err, testCase.err)
		//}
	}
}

func testStudentAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
		if _, err := client.GetSubmissions(testCase.ctx, &qf.SubmissionRequest{
			UserID:   testCase.userID,
			CourseID: testCase.courseID,
		}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetSubmissions", err, testCase.err)
		}
		if _, err := client.GetAssignments(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetAssignments", err, testCase.err)
		}
		if _, err := client.GetEnrollmentsByCourse(testCase.ctx, &qf.EnrollmentRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetEnrollmentsByCourse", err, testCase.err)
		}
		if _, err := client.GetRepositories(testCase.ctx, &qf.URLRequest{CourseID: testCase.courseID}); !errors.Is(err, testCase.err) {
			logError(t, testCase.name, "GetRepositories", err, testCase.err)
		}
	}
}

func testUnrestrictedAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
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
}

func testUserAccess(client qf.QuickFeedServiceClient, t *testing.T, course *qf.Course, tests accessTests) {
	for _, testCase := range tests {
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
}

func testAdminAccess(client qf.QuickFeedServiceClient, t *testing.T, course *qf.Course, tests accessTests) {
	for _, testCase := range tests {
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
		if _, err := client.GetUserByCourse(testCase.ctx, &qf.CourseUserRequest{
			CourseCode: course.Code,
			CourseYear: course.Year,
			UserLogin:  "student",
		}); !errors.Is(err, interceptor.ErrAccessDenied) {
			logError(t, testCase.name, "GetUserByCourse", err, testCase.err)
		}
	}
}

func logError(t *testing.T, name, method string, got, want error) {
	t.Errorf("unexpected error for %s calling %s: expected %v, got %v", name, method, want, got)
}
