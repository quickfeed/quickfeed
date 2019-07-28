package web_test

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/google/go-cmp/cmp"
	"github.com/markbates/goth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	_ "github.com/mattn/go-sqlite3"
)

var allCourses = []*pb.Course{
	{
		Name:            "Distributed Systems",
		CourseCreatorID: 1,
		Code:            "DAT520",
		Year:            2018,
		Tag:             "Spring",
		Provider:        "fake",
		OrganizationID:  1,
	},
	{
		Name:            "Operating Systems",
		CourseCreatorID: 1,
		Code:            "DAT320",
		Year:            2017,
		Tag:             "Fall",
		Provider:        "fake",
		OrganizationID:  2,
	}, {
		Name:            "New Systems",
		CourseCreatorID: 1,
		Code:            "DATx20",
		Year:            2019,
		Tag:             "Fall",
		Provider:        "fake",
		OrganizationID:  3,
	}, {
		Name:            "Hyped Systems",
		CourseCreatorID: 1,
		Code:            "DATx20",
		Year:            2019,
		Tag:             "Fall",
		Provider:        "fake",
		OrganizationID:  4,
	},
}

func TestGetCourses(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 10)
	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{})

	var testCourses []*pb.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(admin.ID, &testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	foundCourses, err := ags.GetCourses(context.Background(), &pb.Void{})
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
	userID := strconv.Itoa(int(user.GetID()))
	meta := metadata.New(map[string]string{"user": userID})
	return metadata.NewIncomingContext(ctx, meta)
}

// fakeProviderMap is a test helper function to create an SCM map.
func fakeProviderMap(t *testing.T) (scm.SCM, *auth.Scms) {
	t.Helper()
	scms := auth.NewScms()
	scm, err := scms.GetOrCreateSCMEntry(zap.NewNop(), "fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	return scm, scms
}

func fakeGothProvider() {
	baseURL := "fake"
	goth.UseProviders(&auth.FakeProvider{
		Callback: auth.GetCallbackURL(baseURL, "fake"),
	})
	goth.UseProviders(&auth.FakeProvider{
		Callback: auth.GetCallbackURL(baseURL, "fake-teacher"),
	})
}

func TestNewCourse(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// set up fake goth provider (only needs to be done once)
	fakeGothProvider()
	admin := createFakeUser(t, db, 10)
	ctx := withUserContext(context.Background(), admin)
	fakeScmProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{})

	for _, testCourse := range allCourses {
		// each course needs a separate directory
		fakeScmProvider.CreateOrganization(ctx, &scm.CreateOrgOptions{Path: "path", Name: "name"})

		respCourse, err := ags.CreateCourse(ctx, testCourse)
		if err != nil {
			t.Fatal(err)
		}

		course, err := db.GetCourse(respCourse.ID)
		if err != nil {
			t.Fatal(err)
		}

		testCourse.ID = respCourse.ID
		if !reflect.DeepEqual(course, testCourse) {
			t.Errorf("have database course\n %+v want\n %+v", course, testCourse)
		}
		if !reflect.DeepEqual(respCourse, course) {
			t.Errorf("have response course\n %+v want\n %+v", respCourse, course)
		}
	}
}

func TestNewCourseExistingRepos(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 10)
	ctx := withUserContext(context.Background(), admin)
	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{})

	directory, _ := fakeProvider.CreateOrganization(ctx, &scm.CreateOrgOptions{Path: "path", Name: "name"})
	for path, private := range web.RepoPaths {
		repoOptions := &scm.CreateRepositoryOptions{Path: path, Organization: directory, Private: private}
		fakeProvider.CreateRepository(ctx, repoOptions)
	}

	course, err := ags.CreateCourse(ctx, allCourses[0])
	if course != nil {
		t.Fatal("expected CreateCourse to fail with AlreadyExists")
	}
	if err != nil && status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected CreateCourse to fail with AlreadyExists, but got: %v", err)
	}
}

func TestEnrollmentProcess(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	ctx := withUserContext(context.Background(), admin)
	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{})
	fakeProvider.CreateOrganization(ctx, &scm.CreateOrgOptions{Path: "path", Name: "name"})

	course, err := ags.CreateCourse(ctx, allCourses[0])
	if err != nil {
		t.Fatal(err)
	}

	stud1 := createFakeUser(t, db, 2)
	enrollStud1 := &pb.Enrollment{CourseID: course.ID, UserID: stud1.ID}
	if _, err = ags.CreateEnrollment(ctx, enrollStud1); err != nil {
		t.Fatal(err)
	}

	// verify that a pending enrollment was indeed created.
	pendingEnrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, stud1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment := &pb.Enrollment{
		ID:       pendingEnrollment.ID,
		CourseID: course.ID,
		UserID:   stud1.ID,
		Status:   pb.Enrollment_PENDING,
		Course:   course,
		User:     stud1,
	}
	// can't use: wantEnrollment.User.RemoveRemoteID()
	wantEnrollment.User.RemoteIdentities = nil
	if !cmp.Equal(pendingEnrollment, wantEnrollment) {
		t.Errorf("enrollment\nhave %+v\nwant %+v\n", pendingEnrollment, wantEnrollment)
	}

	enrollStud1.Status = pb.Enrollment_STUDENT
	if _, err = ags.UpdateEnrollment(ctx, enrollStud1); err != nil {
		t.Fatal(err)
	}

	// verify that the enrollment was updated to student status.
	acceptedEnrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, stud1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Status = pb.Enrollment_STUDENT
	if !cmp.Equal(acceptedEnrollment, wantEnrollment) {
		t.Errorf("enrollment\nhave %+v\nwant %+v\n", acceptedEnrollment, wantEnrollment)
	}

	// create another user and enroll as student

	stud2 := createFakeUser(t, db, 3)
	enrollStud2 := &pb.Enrollment{CourseID: course.ID, UserID: stud2.ID}
	if _, err = ags.CreateEnrollment(ctx, enrollStud2); err != nil {
		t.Fatal(err)
	}
	enrollStud2.Status = pb.Enrollment_STUDENT
	if _, err = ags.UpdateEnrollment(ctx, enrollStud2); err != nil {
		t.Fatal(err)
	}
	// verify that the stud2 was enrolled with student status.
	acceptedEnrollment, err = db.GetEnrollmentByCourseAndUser(course.ID, stud2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = acceptedEnrollment.ID
	wantEnrollment.Status = pb.Enrollment_STUDENT
	wantEnrollment.UserID = stud2.ID
	wantEnrollment.User = stud2
	wantEnrollment.User.RemoteIdentities = nil
	if !cmp.Equal(acceptedEnrollment, wantEnrollment) {
		t.Errorf("enrollment\nhave %+v\nwant %+v\n", acceptedEnrollment, wantEnrollment)
	}

	// promote stud2 to teaching assistant

	enrollStud2.Status = pb.Enrollment_TEACHER
	if _, err = ags.UpdateEnrollment(ctx, enrollStud2); err != nil {
		t.Fatal(err)
	}
	// verify that the stud2 was promoted to teacher status.
	acceptedEnrollment, err = db.GetEnrollmentByCourseAndUser(course.ID, stud2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.ID = acceptedEnrollment.ID
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
	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{})

	var testCourses []*pb.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(admin.ID, &testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourses[2].ID); err != nil {
		t.Fatal(err)
	}

	courses_request := &pb.RecordRequest{ID: user.ID}
	courses, err := ags.GetCoursesWithEnrollment(context.Background(), courses_request)
	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*pb.Course{
		{ID: testCourses[0].ID, Enrolled: pb.Enrollment_PENDING},
		{ID: testCourses[1].ID, Enrolled: pb.Enrollment_REJECTED},
		{ID: testCourses[2].ID, Enrolled: pb.Enrollment_STUDENT},
		{ID: testCourses[3].ID, Enrolled: pb.Enrollment_NONE},
	}
	for i, course := range courses.Courses {
		if course.ID != wantCourses[i].ID {
			t.Errorf("have course %+v want %+v", course.ID, wantCourses[i].ID)
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
		err := db.CreateCourse(admin.ID, &testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	user := createFakeUser(t, db, 2)
	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{})

	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}); err != nil {
		t.Fatal(err)
	}

	// user enrollment is rejected for course 1 and enrolled for course 2, still pending for course 0
	if err := db.RejectEnrollment(user.ID, testCourses[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourses[2].ID); err != nil {
		t.Fatal(err)
	}

	stats := make([]pb.Enrollment_UserStatus, 0)
	stats = append(stats, pb.Enrollment_REJECTED, pb.Enrollment_STUDENT)
	course_req := &pb.RecordRequest{ID: user.ID, Statuses: stats}
	courses, err := ags.GetCoursesWithEnrollment(context.Background(), course_req)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses, err := db.GetCoursesByUser(user.ID, pb.Enrollment_REJECTED, pb.Enrollment_STUDENT)
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
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}
	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{})

	foundCourse, err := ags.GetCourse(context.Background(), &pb.RecordRequest{ID: course.ID})
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(foundCourse, &course) {
		t.Errorf("have course %+v want %+v", foundCourse, course)
	}
}
