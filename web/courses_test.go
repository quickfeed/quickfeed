package web_test

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/web/grpc_service"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/metadata"

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

	adminUser := createFakeUser(t, db, 10)
	testscms := make(map[string]scm.SCM)

	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})

	user := createFakeUser(t, db, 11)
	testCourse := *allCourses[0]
	/*
		dbUser, err := db.GetUserByRemoteIdentity(&pb.RemoteIdentity{RemoteId: 10})
		if err != nil {
			t.Fatal(err)
		}*/
	// create metadata for user imitating contect coming from the browser
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(adminUser.Id))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

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

	enrollRequest := pb.ActionRequest{UserId: user.Id, CourseId: respCourse.Id}
	if _, err = test_ag.CreateEnrollment(cont, &enrollRequest); err != nil {
		t.Fatal(err)
	}

	if err = db.EnrollTeacher(user.Id, respCourse.Id); err != nil {
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

	if !reflect.DeepEqual(respCourse, course) {
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
	if !cmp.Equal(enrollment, wantEnrollment) {
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

	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(admin.Id))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

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

	enroll_request.Status = pb.Enrollment_STUDENT
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
	if !cmp.Equal(acceptedEnrollment, wantEnrollment) {
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
	courses, err := test_ag.GetCoursesWithEnrollment(context.Background(), courses_request)

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
	courses, err := test_ag.GetCoursesWithEnrollment(context.Background(), course_req)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses, err := db.GetCoursesByUser(user.Id, pb.Enrollment_REJECTED, pb.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(courses.Courses, wantCourses) {
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

	foundCourse, err := test_ag.GetCourse(context.Background(), &pb.RecordRequest{Id: course.Id})

	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(foundCourse, &course) {
		t.Errorf("have course %+v want %+v", foundCourse, course)
	}
}
