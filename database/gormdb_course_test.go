package database_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGormDBCreateCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:           "name",
		Code:           "code",
		Year:           2017,
		Tag:            "tag",
		OrganizationID: 1,
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)
	if course.ID == 0 {
		t.Error("expected id to be set")
	}

	// check that admin (teacher) was automatically enrolled when creating course
	enroll, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if enroll.CourseID != course.ID || enroll.UserID != admin.ID {
		t.Errorf("expected user %d to be enrolled in course %d, but got user %d and course %d", admin.ID, course.ID, enroll.UserID, enroll.CourseID)
	}
	if enroll.Status != qf.Enrollment_TEACHER || enroll.State != qf.Enrollment_VISIBLE {
		t.Errorf("expected enrolled user to be teacher and visible, but got status: %v and state: %v", enroll.Status, enroll.State)
	}

	// check that no users were enrolled as students
	enrolls, err := db.GetEnrollmentsByCourse(course.ID, qf.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	if len(enrolls) > 0 {
		t.Errorf("expected no enrollments, but got %d enrollments: %v", len(enrolls), enrolls)
	}

	// check that exactly one user was enrolled as teacher for the course
	enrolls, err = db.GetEnrollmentsByCourse(course.ID, qf.Enrollment_TEACHER)
	if err != nil {
		t.Fatal(err)
	}
	if len(enrolls) != 1 {
		t.Errorf("expected exactly one enrollment, but got %d enrollments: %v", len(enrolls), enrolls)
	}
	for _, enroll := range enrolls {
		if enroll.CourseID != course.ID || enroll.UserID != admin.ID {
			t.Errorf("expected user %d to be enrolled in course %d, but got user %d and course %d", admin.ID, course.ID, enroll.UserID, enroll.CourseID)
		}
		if enroll.Status != qf.Enrollment_TEACHER || enroll.State != qf.Enrollment_VISIBLE {
			t.Errorf("expected enrolled user to be teacher and visible, but got status: %v and state: %v", enroll.Status, enroll.State)
		}
	}
}

func TestGormDBGetCoursesByUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	c1 := &qf.Course{OrganizationID: 1, Code: "DAT101", Year: 1}
	c2 := &qf.Course{OrganizationID: 2, Code: "DAT101", Year: 2}
	c3 := &qf.Course{OrganizationID: 3, Code: "DAT101", Year: 3}
	c4 := &qf.Course{OrganizationID: 4, Code: "DAT101", Year: 4}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)
	qtest.CreateCourse(t, db, admin, c3)
	qtest.CreateCourse(t, db, admin, c4)

	user := qtest.CreateFakeUser(t, db, 10)
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: c1.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: c2.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: c3.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.ID, c2.ID); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: c3.ID,
		Status:   qf.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	gotCourses, err := db.GetCoursesByUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*qf.Course{
		{ID: c1.ID, OrganizationID: 1, Code: "DAT101", Year: 1, CourseCreatorID: admin.ID, Enrolled: qf.Enrollment_PENDING},
		{ID: c2.ID, OrganizationID: 2, Code: "DAT101", Year: 2, CourseCreatorID: admin.ID, Enrolled: qf.Enrollment_NONE},
		{ID: c3.ID, OrganizationID: 3, Code: "DAT101", Year: 3, CourseCreatorID: admin.ID, Enrolled: qf.Enrollment_STUDENT},
		{ID: c4.ID, OrganizationID: 4, Code: "DAT101", Year: 4, CourseCreatorID: admin.ID, Enrolled: qf.Enrollment_NONE},
	}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCoursesByUser() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}
}

func TestGormDBCreateCourseNonAdmin(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	qtest.CreateCourse(t, db, admin, &qf.Course{})

	nonAdmin := qtest.CreateFakeUser(t, db, 11)
	// the following should fail to create a course
	if err := db.CreateCourse(nonAdmin.ID, &qf.Course{}); err == nil {
		t.Fatal(err)
	}
}

func TestGormDBGetCourses(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	c1 := &qf.Course{OrganizationID: 1, Code: "DAT101", Year: 1}
	c2 := &qf.Course{OrganizationID: 2, Code: "DAT101", Year: 2}
	c3 := &qf.Course{OrganizationID: 3, Code: "DAT101", Year: 3}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)
	qtest.CreateCourse(t, db, admin, c3)

	gotCourses, err := db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}
	wantCourses := []*qf.Course{c1, c2, c3}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	// An empty list should return the same as no argument, it makes no
	// sense to ask the database to return no courses.
	gotCourses, err = db.GetCourses([]uint64{}...)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	gotCourses, err = db.GetCourses(c1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses = []*qf.Course{c1}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	gotCourses, err = db.GetCourses(c1.ID, c2.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantCourses = []*qf.Course{c1, c2}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}
}

func TestGormDBGetCourse(t *testing.T) {
	wantCourse := &qf.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 1234,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, wantCourse)

	// Get the created course.
	gotCourse, err := db.GetCourse(wantCourse.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse, +gotCourse):\n%s", diff)
	}
}

func TestGormDBGetCourseNoRecord(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if _, err := db.GetCourse(10, false); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBUpdateCourse(t *testing.T) {
	course := &qf.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		Dockerfile:     "Dockerfile1",
		OrganizationID: 1234,
	}
	wantCourse := &qf.Course{
		Name:           "Test Course Edit",
		Code:           "DAT100-1",
		Year:           2018,
		Tag:            "Autumn",
		Dockerfile:     "Another Dockerfile1",
		OrganizationID: 12345,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	qtest.CreateCourse(t, db, admin, course)

	wantCourse.ID = course.ID
	wantCourse.CourseCreatorID = admin.ID
	if err := db.UpdateCourse(wantCourse); err != nil {
		t.Fatal(err)
	}

	// Get the updated course.
	gotCourse, err := db.GetCourse(course.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("course mismatch (-want +got):\n%s", diff)
	}
}

func TestGormDBGetCourseByOrganization(t *testing.T) {
	wantCourse := &qf.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 1234,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, wantCourse)

	// Get the created course.
	gotCourse, err := db.GetCourseByOrganizationID(wantCourse.OrganizationID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse, +gotCourse):\n%s", diff)
	}
}

func TestGormDBCourseUniqueConstraint(t *testing.T) {
	// Test that a course with the same organization ID or code and year cannot be created.
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wantCourse := &qf.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 1235,
	}
	course := &qf.Course{
		Name:           "Test Course 2",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 1234,
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, wantCourse); err != nil {
		t.Fatal(err)
	}

	// CreateCourse should fail because the unique constraint (course.code, course.year) is violated
	if err := db.CreateCourse(admin.ID, course); err != nil && !errors.Is(err, database.ErrCourseExists) {
		t.Fatal(err)
	}

	// CreateCourse should fail because OrganizationID is not unique
	if err := db.CreateCourse(admin.ID, &qf.Course{OrganizationID: wantCourse.OrganizationID}); err != nil && !errors.Is(err, database.ErrCourseExists) {
		t.Fatal(err)
	}

	gotCourse, err := db.GetCourse(wantCourse.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("course mismatch (-want +got):\n%s", diff)
	}

	// Now create a course with same code but different year
	course.Year = 2018
	// CreateCourse should succeed because the unique constraint (course.code, course.year) is not violated
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}
}

func TestGetCourseTeachers(t *testing.T) {
	tests := map[string]struct {
		wantTeachers, students []*qf.User
	}{
		"Basic": {
			wantTeachers: []*qf.User{{Login: "teacher1"}, {Login: "teacher2"}},
			students:     []*qf.User{{Login: "student1"}},
		},
		"No teachers": {
			wantTeachers: []*qf.User{},
			students:     []*qf.User{{Login: "student1"}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)
			defer cleanup()
			admin := qtest.CreateFakeUser(t, db, 1)
			course := &qf.Course{}
			qtest.CreateCourse(t, db, admin, course)
			nextRemoteID := uint64(2)
			for _, teacher := range tt.wantTeachers {
				if err := db.CreateUser(teacher); err != nil {
					t.Error(err)
				}
				qtest.EnrollTeacher(t, db, teacher, course)
				nextRemoteID++
			}
			for _, student := range tt.students {
				if err := db.CreateUser(student); err != nil {
					t.Error(err)
				}
				qtest.EnrollStudent(t, db, student, course)
				nextRemoteID++
			}
			// We add the admin to the list of wantTeachers,
			// since the admin is always registered as a teacher when the course is created.
			tt.wantTeachers = append(tt.wantTeachers, admin)
			sort.Slice(tt.wantTeachers, func(i, j int) bool {
				return tt.wantTeachers[i].GetID() < tt.wantTeachers[j].GetID()
			})
			gotTeachers, err := db.GetCourseTeachers(course)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wantTeachers, gotTeachers, protocmp.Transform()); diff != "" {
				t.Errorf("GetCourseTeachers mismatch (-wantTeachers +gotTeachers):\n%s", diff)
			}
		})
	}
}
