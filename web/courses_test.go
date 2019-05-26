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

// withUserContext is a test helper function to create metadata for the
// given user mimicking the context coming from the browser.
func withUserContext(ctx context.Context, user *pb.User) context.Context {
	userID := strconv.Itoa(int(user.GetId()))
	meta := metadata.New(map[string]string{"user": userID})
	return metadata.NewIncomingContext(ctx, meta)
}

// fakeProviderMap is a test helper function to create an SCM map.
func fakeProviderMap(ctx context.Context) map[string]scm.SCM {
	fakeProvider := scm.NewFakeSCMClient()
	scmMap := make(map[string]scm.SCM)
	// add the fake scm to the scm map with the fake token as key
	scmMap["token"] = fakeProvider
	return scmMap
}

func TestNewCourse(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	adminUser := createFakeUser(t, db, 10)
	ctx := withUserContext(context.Background(), adminUser)
	scmMap := fakeProviderMap(ctx)
	fakeProvider := scmMap["token"]

	ags := grpc_service.NewAutograderService(db, scmMap, web.BaseHookOptions{})
	for _, testCourse := range allCourses {
		// each course needs a separate directory
		fakeProvider.CreateDirectory(ctx, &scm.CreateDirectoryOptions{Path: "path", Name: "name"})

		respCourse, err := ags.CreateCourse(ctx, testCourse)
		if err != nil {
			t.Fatal(err)
		}

		course, err := db.GetCourse(respCourse.Id)
		if err != nil {
			t.Fatal(err)
		}

		testCourse.Id = respCourse.Id
		if !reflect.DeepEqual(course, testCourse) {
			t.Errorf("have database course\n %+v want\n %+v", course, testCourse)
		}
		if !reflect.DeepEqual(respCourse, course) {
			t.Errorf("have response course\n %+v want\n %+v", respCourse, course)
		}
	}
}

func TestEnrollmentProcess(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	ctx := withUserContext(context.Background(), admin)
	scmMap := fakeProviderMap(ctx)
	fakeProvider := scmMap["token"]
	fakeProvider.CreateDirectory(ctx, &scm.CreateDirectoryOptions{Path: "path", Name: "name"})

	ags := grpc_service.NewAutograderService(db, scmMap, web.BaseHookOptions{})
	course, err := ags.CreateCourse(ctx, allCourses[0])
	if err != nil {
		t.Fatal(err)
	}

	stud1 := createFakeUser(t, db, 2)
	enrollStud1 := &pb.ActionRequest{CourseId: course.Id, UserId: stud1.Id}
	if _, err = ags.CreateEnrollment(ctx, enrollStud1); err != nil {
		t.Fatal(err)
	}

	// verify that an appropriate enrollment was indeed created.
	pendingEnrollment, err := db.GetEnrollmentByCourseAndUser(course.Id, stud1.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment := &pb.Enrollment{
		Id:       pendingEnrollment.Id,
		CourseId: course.Id,
		UserId:   stud1.Id,
	}
	if !cmp.Equal(pendingEnrollment, wantEnrollment) {
		t.Errorf("have enrollment\n %+v\n want\n %+v", pendingEnrollment, wantEnrollment)
	}

	enrollStud1.Status = pb.Enrollment_STUDENT
	if _, err = ags.UpdateEnrollment(ctx, enrollStud1); err != nil {
		t.Fatal(err)
	}

	// verify that the enrollment was updated to student status.
	acceptedEnrollment, err := db.GetEnrollmentByCourseAndUser(course.Id, stud1.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Status = pb.Enrollment_STUDENT
	if !cmp.Equal(acceptedEnrollment, wantEnrollment) {
		t.Errorf("have enrollment %+v want %+v", acceptedEnrollment, wantEnrollment)
	}

	// create another user and enroll as student

	stud2 := createFakeUser(t, db, 3)
	enrollStud2 := &pb.ActionRequest{CourseId: course.Id, UserId: stud2.Id}
	if _, err = ags.CreateEnrollment(ctx, enrollStud2); err != nil {
		t.Fatal(err)
	}
	enrollStud2.Status = pb.Enrollment_STUDENT
	if _, err = ags.UpdateEnrollment(ctx, enrollStud2); err != nil {
		t.Fatal(err)
	}
	// verify that the stud2 was enrolled with student status.
	acceptedEnrollment, err = db.GetEnrollmentByCourseAndUser(course.Id, stud2.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Id = acceptedEnrollment.Id
	wantEnrollment.Status = pb.Enrollment_STUDENT
	wantEnrollment.UserId = stud2.Id
	if !cmp.Equal(acceptedEnrollment, wantEnrollment) {
		t.Errorf("have enrollment %+v want %+v", acceptedEnrollment, wantEnrollment)
	}

	// promote stud2 to teaching assistant

	enrollStud2.Status = pb.Enrollment_TEACHER
	if _, err = ags.UpdateEnrollment(ctx, enrollStud2); err != nil {
		t.Fatal(err)
	}
	// verify that the stud2 was promoted to teacher status.
	acceptedEnrollment, err = db.GetEnrollmentByCourseAndUser(course.Id, stud2.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Id = acceptedEnrollment.Id
	wantEnrollment.Status = pb.Enrollment_TEACHER
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
