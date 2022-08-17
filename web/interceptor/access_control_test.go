package interceptor_test

import (
	"context"
	"errors"
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
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, scm.TestSCMManager(), web.BaseHookOptions{}, &ci.Local{})

	s := grpc.NewServer() // skipcq: GO-S0902
	qf.RegisterQuickFeedServiceServer(s, ags)
	if err := web.VerifyAccessControlMethods(s); err != nil {
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

	lis := bufconn.Listen(BufSize)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	opt := grpc.ChainUnaryInterceptor(
		interceptor.UnaryUserVerifier(logger, tm),
		interceptor.AccessControl(logger, tm),
	)
	s := grpc.NewServer(opt) // skipcq: GO-S0902
	qf.RegisterQuickFeedServiceServer(s, ags)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
			return
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

	f := func(t *testing.T, id uint64) context.Context {
		token, err := tm.NewAuthCookie(id)
		if err != nil {
			t.Fatal(err)
		}
		return qtest.WithAuthCookie(ctx, interceptor.AuthTokenString(token.Value))
	}
	courseAdminContext := f(t, courseAdmin.ID)
	groupStudentContext := f(t, groupStudent.ID)
	studentContext := f(t, student.ID)
	userContext := f(t, user.ID)
	adminContext := f(t, admin.ID)

	freeAccessTest := accessTests{
		{"admin", courseAdminContext, 0, course.ID, 0, true},
		{"student", studentContext, 0, course.ID, 0, true},
		{"student", groupStudentContext, 0, course.ID, 0, true},
		{"user", userContext, 0, course.ID, 0, true},
		{"non-teacher admin", adminContext, 0, course.ID, 0, true},
	}
	testUnrestrictedAccess(client, t, freeAccessTest)

	userAccessTests := accessTests{
		{"correct user ID", userContext, user.ID, course.ID, 0, true},
		{"incorrect user ID", userContext, groupStudent.ID, course.ID, 0, false},
	}
	testUserAccess(client, t, userAccessTests)

	studentAccessTests := accessTests{
		{"course admin", courseAdminContext, courseAdmin.ID, course.ID, 0, true},
		{"admin, not enrolled in a course", adminContext, admin.ID, course.ID, 0, false},
		{"user, not enrolled in the course", userContext, user.ID, course.ID, 0, false},
		{"student", studentContext, student.ID, course.ID, 0, true},
		{"student of another course", studentContext, student.ID, 123, 0, false},
	}
	testStudentAccess(client, t, studentAccessTests)

	groupAccessTests := accessTests{
		{"student in a group", groupStudentContext, groupStudent.ID, course.ID, group.ID, true},
		{"student, not in a group", studentContext, student.ID, course.ID, group.ID, false},
		{"student in a group, wrong group ID in request", studentContext, student.ID, course.ID, 123, false},
	}
	testGroupAccess(client, t, groupAccessTests)

	teacherAccessTests := accessTests{
		{"course teacher", courseAdminContext, groupStudent.ID, course.ID, group.ID, true},
		{"student", studentContext, student.ID, course.ID, group.ID, false},
		{"admin, not enrolled in the course", adminContext, admin.ID, course.ID, group.ID, false},
	}
	testTeacherAccess(client, t, teacherAccessTests, course)

	adminAccessTests := accessTests{
		{"admin (accessing own info)", courseAdminContext, courseAdmin.ID, course.ID, group.ID, true},
		{"admin (accessing other user's info)", courseAdminContext, user.ID, course.ID, group.ID, true},
		{"non admin (accessing admin's info)", studentContext, courseAdmin.ID, course.ID, group.ID, false},
		{"non admin (accessing other user's info)", studentContext, user.ID, course.ID, group.ID, false},
	}
	testAdminAccess(client, t, course, adminAccessTests)

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

	for _, testCase := range createGroupTests {
		_, err := client.CreateGroup(testCase.ctx, testCase.group)
		verifyAccess(t, err, testCase.access, "CreateGroup", testCase.name)
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

	for _, testCase := range adminStatusChangeTests {
		_, err := client.UpdateUser(testCase.ctx, testCase.user)
		verifyAccess(t, err, testCase.access, "UpdateUser", testCase.name)
	}
}

func testUnrestrictedAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
		_, err := client.GetUser(testCase.ctx, &qf.Void{})
		verifyAccess(t, err, testCase.access, "GetUser", testCase.name)
		_, err = client.GetCourse(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "GetCourse", testCase.name)
		_, err = client.GetCourses(testCase.ctx, &qf.Void{})
		verifyAccess(t, err, testCase.access, "GetCourses", testCase.name)
	}
}

func testUserAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
		enrol := &qf.Enrollment{
			CourseID: testCase.courseID,
			UserID:   testCase.userID,
		}
		enrolRequest := &qf.EnrollmentStatusRequest{
			UserID: testCase.userID,
		}
		_, err := client.CreateEnrollment(testCase.ctx, enrol)
		verifyAccess(t, err, testCase.access, "CreateEnrollment", testCase.name)
		_, err = client.UpdateCourseVisibility(testCase.ctx, enrol)
		verifyAccess(t, err, testCase.access, "UpdateCourseVisibility", testCase.name)
		_, err = client.GetCoursesByUser(testCase.ctx, enrolRequest)
		verifyAccess(t, err, testCase.access, "GetCoursesByUser", testCase.name)
		_, err = client.UpdateUser(testCase.ctx, &qf.User{ID: testCase.userID})
		verifyAccess(t, err, testCase.access, "UpdateUser", testCase.name)
		_, err = client.GetEnrollmentsByUser(testCase.ctx, enrolRequest)
		verifyAccess(t, err, testCase.access, "GetEnrollmentsByCourse", testCase.name)
		_, err = client.UpdateUser(testCase.ctx, &qf.User{ID: testCase.userID})
		verifyAccess(t, err, testCase.access, "UpdateUser", testCase.name)
	}
}

func testStudentAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
		_, err := client.GetSubmissions(testCase.ctx, &qf.SubmissionRequest{
			UserID:   testCase.userID,
			CourseID: testCase.courseID,
		})
		verifyAccess(t, err, testCase.access, "GetSubmissions", testCase.name)
		_, err = client.GetAssignments(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "GetAssignments", testCase.name)
		_, err = client.GetEnrollmentsByCourse(testCase.ctx, &qf.EnrollmentRequest{CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "GetEnrollmentsByCourse", testCase.name)
		_, err = client.GetRepositories(testCase.ctx, &qf.URLRequest{CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "GetRepositories", testCase.name)
	}
}

func testGroupAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests) {
	for _, testCase := range tests {
		_, err := client.GetGroupByUserAndCourse(testCase.ctx, &qf.GroupRequest{
			CourseID: testCase.courseID,
			UserID:   testCase.userID,
			GroupID:  testCase.groupID,
		})
		verifyAccess(t, err, testCase.access, "GetGroupByUserAndCourse", testCase.name)
		_, err = client.GetGroup(testCase.ctx, &qf.GetGroupRequest{GroupID: testCase.groupID})
		verifyAccess(t, err, testCase.access, "GetGroup", testCase.name)
	}
}

func testTeacherAccess(client qf.QuickFeedServiceClient, t *testing.T, tests accessTests, course *qf.Course) {
	for _, testCase := range tests {
		_, err := client.GetGroupByUserAndCourse(testCase.ctx, &qf.GroupRequest{
			UserID:   testCase.userID,
			CourseID: testCase.courseID,
		})
		verifyAccess(t, err, testCase.access, "GetGroupByUserAndCourse", testCase.name)
		_, err = client.GetGroup(testCase.ctx, &qf.GetGroupRequest{GroupID: testCase.groupID})
		verifyAccess(t, err, testCase.access, "GetGroup", testCase.name)
		_, err = client.DeleteGroup(testCase.ctx, &qf.GroupRequest{
			GroupID:  testCase.groupID,
			CourseID: testCase.courseID,
			UserID:   testCase.userID,
		})
		verifyAccess(t, err, testCase.access, "DeleteGroup", testCase.name)
		_, err = client.UpdateGroup(testCase.ctx, &qf.Group{CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "UpdateGroup", testCase.name)
		_, err = client.UpdateCourse(testCase.ctx, course)
		verifyAccess(t, err, testCase.access, "UpdateCourse", testCase.name)
		_, err = client.UpdateEnrollments(testCase.ctx, &qf.Enrollments{
			Enrollments: []*qf.Enrollment{{ID: 1, CourseID: testCase.courseID}},
		})
		verifyAccess(t, err, testCase.access, "UpdateEnrollments", testCase.name)
		_, err = client.UpdateAssignments(testCase.ctx, &qf.CourseRequest{CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "UpdateAssignments", testCase.name)
		_, err = client.UpdateSubmission(testCase.ctx, &qf.UpdateSubmissionRequest{SubmissionID: 1, CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "UpdateSubmission", testCase.name)
		_, err = client.UpdateSubmissions(testCase.ctx, &qf.UpdateSubmissionsRequest{AssignmentID: 1, CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "UpdateSubmissions", testCase.name)
		_, err = client.RebuildSubmissions(testCase.ctx, &qf.RebuildRequest{
			AssignmentID: 1,
			RebuildType: &qf.RebuildRequest_CourseID{
				CourseID: testCase.courseID,
			},
		})
		verifyAccess(t, err, testCase.access, "RebuildSubmissions", testCase.name)
		_, err = client.CreateBenchmark(testCase.ctx, &qf.GradingBenchmark{CourseID: testCase.courseID, AssignmentID: 1})
		verifyAccess(t, err, testCase.access, "CreateBenchmark", testCase.name)
		_, err = client.UpdateBenchmark(testCase.ctx, &qf.GradingBenchmark{CourseID: testCase.courseID, AssignmentID: 1})
		verifyAccess(t, err, testCase.access, "UpdateBenchmark", testCase.name)
		_, err = client.DeleteBenchmark(testCase.ctx, &qf.GradingBenchmark{CourseID: testCase.courseID, AssignmentID: 1})
		verifyAccess(t, err, testCase.access, "DeleteBenchmark", testCase.name)
		_, err = client.CreateCriterion(testCase.ctx, &qf.GradingCriterion{CourseID: testCase.courseID, BenchmarkID: 1})
		verifyAccess(t, err, testCase.access, "CreateCriterion", testCase.name)
		_, err = client.UpdateCriterion(testCase.ctx, &qf.GradingCriterion{CourseID: testCase.courseID, BenchmarkID: 1})
		verifyAccess(t, err, testCase.access, "UpdateCriterion", testCase.name)
		_, err = client.DeleteCriterion(testCase.ctx, &qf.GradingCriterion{CourseID: testCase.courseID, BenchmarkID: 1})
		verifyAccess(t, err, testCase.access, "DeleteCriterion", testCase.name)
		_, err = client.CreateReview(testCase.ctx, &qf.ReviewRequest{
			CourseID: testCase.courseID,
			Review: &qf.Review{
				SubmissionID: 1,
				ReviewerID:   1,
			},
		})
		verifyAccess(t, err, testCase.access, "CreateReview", testCase.name)
		_, err = client.UpdateReview(testCase.ctx, &qf.ReviewRequest{
			CourseID: testCase.courseID,
			Review: &qf.Review{
				SubmissionID: 1,
				ReviewerID:   1,
			},
		})
		verifyAccess(t, err, testCase.access, "UpdateReview", testCase.name)
		_, err = client.GetReviewers(testCase.ctx, &qf.SubmissionReviewersRequest{
			CourseID:     testCase.courseID,
			SubmissionID: 1,
		})
		verifyAccess(t, err, testCase.access, "GetReviewers", testCase.name)
		_, err = client.IsEmptyRepo(testCase.ctx, &qf.RepositoryRequest{CourseID: testCase.courseID})
		verifyAccess(t, err, testCase.access, "IsEmptyRepo", testCase.name)
	}
}

func testAdminAccess(client qf.QuickFeedServiceClient, t *testing.T, course *qf.Course, tests accessTests) {
	for _, testCase := range tests {
		_, err := client.UpdateUser(testCase.ctx, &qf.User{ID: testCase.userID})
		verifyAccess(t, err, testCase.access, "UpdateUser", testCase.name)
		_, err = client.GetEnrollmentsByUser(testCase.ctx, &qf.EnrollmentStatusRequest{UserID: testCase.userID})
		verifyAccess(t, err, testCase.access, "GetEnrollmentsByUser", testCase.name)
		_, err = client.GetUsers(testCase.ctx, &qf.Void{})
		verifyAccess(t, err, testCase.access, "GetUsers", testCase.name)
		_, err = client.GetOrganization(testCase.ctx, &qf.OrgRequest{OrgName: "testorg"})
		verifyAccess(t, err, testCase.access, "GetOrganization", testCase.name)
		_, err = client.CreateCourse(testCase.ctx, course)
		verifyAccess(t, err, testCase.access, "CreateCourse", testCase.name)
		_, err = client.GetUserByCourse(testCase.ctx, &qf.CourseUserRequest{
			CourseCode: course.Code,
			CourseYear: course.Year,
			UserLogin:  "student",
		})
		verifyAccess(t, err, testCase.access, "GetUserByCourse", testCase.name)
	}
}

func verifyAccess(t *testing.T, err error, expected bool, method, name string) {
	if errors.Is(err, interceptor.ErrAccessDenied) == expected {
		t.Errorf("unexpected access control response for %s (%s): expected access: %v, got error: %v", name, method, expected, err)
	}
}
