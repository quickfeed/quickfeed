package web_test

import (
	"context"
	"reflect"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/web/grpc_service"
	"google.golang.org/grpc/metadata"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	_ "github.com/mattn/go-sqlite3"
)

var allCourses = []*pb.Course{
	{
		Name:            "Distributed Systems",
		CoursecreatorId: 1,
		Code:            "DAT520",
		Year:            2018,
		Tag:             "Spring",
		Provider:        "fake",
		DirectoryId:     1,
	},
	{
		Name:            "Operating Systems",
		CoursecreatorId: 1,
		Code:            "DAT320",
		Year:            2017,
		Tag:             "Fall",
		Provider:        "fake",
		DirectoryId:     2,
	}, {
		Name:            "New Systems",
		CoursecreatorId: 1,
		Code:            "DATx20",
		Year:            2019,
		Tag:             "Fall",
		Provider:        "fake",
		DirectoryId:     3,
	}, {
		Name:            "Hyped Systems",
		CoursecreatorId: 1,
		Code:            "DATx20",
		Year:            2019,
		Tag:             "Fall",
		Provider:        "fake",
		DirectoryId:     4,
	},
}

// createFakeUser is a test helper to create a user in the database
// with the given remote id and the fake scm provider.
func createFakeUser(t *testing.T, db database.Database, remoteId uint64) *pb.User {
	var user pb.User
	err := db.CreateUserFromRemoteIdentity(&user,
		&pb.RemoteIdentity{
			Provider:    "fake",
			RemoteId:    remoteId,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return &user
}

func TestListCourses(t *testing.T) {
	const route = "/courses"

	db, cleanup := setup(t)
	defer cleanup()

	user := createFakeUser(t, db, 1)
	var testCourses []*pb.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(user.Id, &testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	foundCourses, err := web.ListCourses(db)
	if err != nil {
		t.Fatal(err)
	}

	for i, course := range foundCourses.Courses {
		if !reflect.DeepEqual(course, testCourses[i]) {
			t.Errorf("have course %+v want %+v", course, testCourses[i])
		}
	}

}

func TestNewCourse(t *testing.T) {
	const (
		route = "/courses"
		fake  = "fake"
	)

	db, cleanup := setup(t)
	defer cleanup()

	testscms := make(map[string]scm.SCM)

	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})

	user := createFakeUser(t, db, 1)
	testCourse := *allCourses[0]

	// create metadata for user imitating contect coming from the browser
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))

	test_scm := scm.NewFakeSCMClient()
	if _, err := test_scm.CreateDirectory(cont, &scm.CreateDirectoryOptions{
		Name: testCourse.Code,
		Path: testCourse.Code,
	}); err != nil {
		t.Fatal(err)
	}
	// add the fake scm to the scm map with the fake token as key
	testscms["token"] = test_scm

	respCourse, err := test_ag.CreateCourse(cont, &testCourse)
	if err != nil {
		t.Fatal(err)
	}

	course, err := db.GetCourse(respCourse.Id)
	if err != nil {
		t.Fatal(err)
	}

	testCourse.Id = respCourse.Id
	if !reflect.DeepEqual(course, &testCourse) {
		t.Errorf("have database course %+v want %+v", course, &testCourse)
	}

	if !reflect.DeepEqual(&respCourse, course) {
		t.Errorf("have response course %+v want %+v", &respCourse, course)
	}

	enrollment, err := db.GetEnrollmentByCourseAndUser(testCourse.Id, user.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment := &pb.Enrollment{
		Id:       enrollment.Id,
		CourseId: testCourse.Id,
		UserId:   user.Id,
		Status:   pb.Enrollment_TEACHER,
	}
	if !reflect.DeepEqual(enrollment, wantEnrollment) {
		t.Errorf("have enrollment %+v want %+v", enrollment, wantEnrollment)
	}

	if len(test_scm.Hooks) != 4 {
		t.Errorf("have %d hooks want %d", len(test_scm.Hooks), 4)
	}
}

func TestEnrollmentProcess(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})

	admin := createFakeUser(t, db, 1)
	user := createFakeUser(t, db, 2)
	testCourse := *allCourses[0]
	if err := db.CreateCourse(admin.Id, &testCourse); err != nil {
		t.Fatal(err)
	}

	enroll_request := &pb.ActionRequest{CourseId: testCourse.Id, UserId: user.Id}
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))
	_, err := test_ag.CreateEnrollment(cont, enroll_request)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that an appropriate enrollment was indeed created.
	pendingEnrollment, err := db.GetEnrollmentByCourseAndUser(testCourse.Id, user.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment := &pb.Enrollment{
		Id:       pendingEnrollment.Id,
		CourseId: testCourse.Id,
		UserId:   user.Id,
	}
	if !reflect.DeepEqual(pendingEnrollment, wantEnrollment) {
		t.Errorf("have enrollment\n %+v\n want\n %+v", pendingEnrollment, wantEnrollment)
	}

	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(cont,
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	_, err = test_ag.UpdateEnrollment(cont, enroll_request)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the enrollment have been accepted.
	acceptedEnrollment, err := db.GetEnrollmentByCourseAndUser(testCourse.Id, user.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Status = pb.Enrollment_STUDENT
	if !reflect.DeepEqual(acceptedEnrollment, wantEnrollment) {
		t.Errorf("have enrollment %+v want %+v", acceptedEnrollment, wantEnrollment)
	}
}

func TestListCoursesWithEnrollment(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	user := createFakeUser(t, db, 2)

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))

	var testCourses []*pb.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(admin.Id, &testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId:   user.Id,
		CourseId: testCourses[0].Id,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId:   user.Id,
		CourseId: testCourses[1].Id,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId:   user.Id,
		CourseId: testCourses[2].Id,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.Id, testCourses[1].Id); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.Id, testCourses[2].Id); err != nil {
		t.Fatal(err)
	}

	courses_request := &pb.RecordRequest{Id: user.Id}
	courses, err := test_ag.GetCoursesWithEnrollment(cont, courses_request)

	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*pb.Course{
		{Id: testCourses[0].Id, Enrolled: pb.Enrollment_PENDING},
		{Id: testCourses[1].Id, Enrolled: pb.Enrollment_REJECTED},
		{Id: testCourses[2].Id, Enrolled: pb.Enrollment_STUDENT},
		{Id: testCourses[3].Id, Enrolled: -1},
	}
	for i, course := range courses.Courses {
		if course.Id != wantCourses[i].Id {
			t.Errorf("have course %+v want %+v", course.Id, wantCourses[i].Id)
		}
		if course.Enrolled != wantCourses[i].Enrolled {
			t.Errorf("have course %+v want %+v", course.Enrolled, wantCourses[i].Enrolled)
		}
	}
}

func TestListCoursesWithEnrollmentStatuses(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var testCourses []*pb.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(admin.Id, &testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	user := createFakeUser(t, db, 2)

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId:   user.Id,
		CourseId: testCourses[0].Id,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId:   user.Id,
		CourseId: testCourses[1].Id,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId:   user.Id,
		CourseId: testCourses[2].Id,
	}); err != nil {
		t.Fatal(err)
	}

	// user enrollment is rejected for course 1 and enrolled for course 2, still pending for course 0
	if err := db.RejectEnrollment(user.Id, testCourses[1].Id); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.Id, testCourses[2].Id); err != nil {
		t.Fatal(err)
	}

	stats := make([]pb.Enrollment_UserStatus, 0)
	stats = append(stats, pb.Enrollment_REJECTED, pb.Enrollment_STUDENT)
	course_req := &pb.RecordRequest{Id: user.Id, Statuses: stats}
	courses, err := test_ag.GetCoursesWithEnrollment(cont, course_req)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses, err := db.GetCoursesByUser(user.Id, pb.Enrollment_REJECTED, pb.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(courses, wantCourses) {
		t.Errorf("have course %+v want %+v", courses, wantCourses)
	}

}

func TestGetCourse(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	course := *allCourses[0]
	err := db.CreateCourse(admin.Id, &course)
	if err != nil {
		t.Fatal(err)
	}
	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(admin.Id))

	foundCourse, err := test_ag.GetCourse(cont, &pb.RecordRequest{Id: course.Id})

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(&foundCourse, &course) {
		t.Errorf("have course %+v want %+v", &foundCourse, &course)
	}
}

func TestNewGroup(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryId = 1
	if err := db.CreateCourse(admin.Id, &course); err != nil {
		t.Fatal(err)
	}
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserId: user.Id, CourseId: course.Id}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.Id, course.Id); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))

	// Prepare provider
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(cont,
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{Id: user.Id})
	group_req := &pb.Group{Name: "Hein's Group", CourseId: course.Id, Users: users}

	respGroup, err := test_ag.CreateGroup(cont, group_req)
	if err != nil {
		t.Fatal(err)
	}

	group, err := db.GetGroup(respGroup.Id)
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	//group.Enrollments = nil
	if !reflect.DeepEqual(&respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}
}

func TestNewGroupTeacherCreator(t *testing.T) {
	const route = "/courses/:cid/groups"

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryId = 1
	if err := db.CreateCourse(admin.Id, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserId: teacher.Id, CourseId: course.Id}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.Id, course.Id); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserId: user.Id, CourseId: course.Id}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.Id, course.Id); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(teacher.Id))

	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(cont,
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{Id: user.Id})
	group_req := &pb.Group{Name: "Hein's Group", CourseId: course.Id, Users: users}

	respGroup, err := test_ag.CreateGroup(cont, group_req)
	if err != nil {
		t.Fatal(err)
	}

	group, err := db.GetGroup(respGroup.Id)
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	group.Enrollments = nil
	if !reflect.DeepEqual(&respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}
}

func TestNewGroupStudentCreateGroupWithTeacher(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryId = 1
	if err := db.CreateCourse(admin.Id, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserId: teacher.Id, CourseId: course.Id}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.Id, course.Id); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserId: user.Id, CourseId: course.Id}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.Id, course.Id); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(teacher.Id))

	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(cont,
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	testscms["token"] = fakeProvider

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{Id: user.Id})
	users = append(users, &pb.User{Id: teacher.Id})
	group_req := &pb.Group{Name: "Hein's Group", CourseId: course.Id, Users: users}

	_, err = test_ag.CreateGroup(cont, group_req)
	if err == nil {
		t.Error("Student trying to enroll teacher should not be possible!")
	}
}
