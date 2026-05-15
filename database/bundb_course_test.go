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
)

func TestBunDBCreateCourse(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:              "name",
		Code:              "code",
		Year:              2017,
		Tag:               "tag",
		ScmOrganizationID: 1,
	}

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)
	if course.GetID() == 0 {
		t.Error("expected id to be set")
	}

	enroll, err := db.GetEnrollmentByCourseAndUser(course.GetID(), admin.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if enroll.GetCourseID() != course.GetID() || enroll.GetUserID() != admin.GetID() {
		t.Errorf("expected user %d to be enrolled in course %d, but got user %d and course %d", admin.GetID(), course.GetID(), enroll.GetUserID(), enroll.GetCourseID())
	}
	if enroll.GetStatus() != qf.Enrollment_TEACHER || enroll.GetState() != qf.Enrollment_VISIBLE {
		t.Errorf("expected enrolled user to be teacher and visible, but got status: %v and state: %v", enroll.GetStatus(), enroll.GetState())
	}

	enrolls, err := db.GetEnrollmentsByCourse(course.GetID(), qf.Enrollment_STUDENT)
	if err != nil {
		t.Fatal(err)
	}
	if len(enrolls) > 0 {
		t.Errorf("expected no enrollments, but got %d enrollments: %v", len(enrolls), enrolls)
	}

	enrolls, err = db.GetEnrollmentsByCourse(course.GetID(), qf.Enrollment_TEACHER)
	if err != nil {
		t.Fatal(err)
	}
	if len(enrolls) != 1 {
		t.Errorf("expected exactly one enrollment, but got %d enrollments: %v", len(enrolls), enrolls)
	}
	for _, enroll := range enrolls {
		if enroll.GetCourseID() != course.GetID() || enroll.GetUserID() != admin.GetID() {
			t.Errorf("expected user %d to be enrolled in course %d, but got user %d and course %d", admin.GetID(), course.GetID(), enroll.GetUserID(), enroll.GetCourseID())
		}
		if enroll.GetStatus() != qf.Enrollment_TEACHER || enroll.GetState() != qf.Enrollment_VISIBLE {
			t.Errorf("expected enrolled user to be teacher and visible, but got status: %v and state: %v", enroll.GetStatus(), enroll.GetState())
		}
	}
}

func TestBunDBGetCoursesByUser(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	c1 := &qf.Course{ScmOrganizationID: 1, Code: "DAT101", Year: 1}
	c2 := &qf.Course{ScmOrganizationID: 2, Code: "DAT101", Year: 2}
	c3 := &qf.Course{ScmOrganizationID: 3, Code: "DAT101", Year: 3}
	c4 := &qf.Course{ScmOrganizationID: 4, Code: "DAT101", Year: 4}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)
	qtest.CreateCourse(t, db, admin, c3)
	qtest.CreateCourse(t, db, admin, c4)

	user := qtest.CreateFakeUser(t, db)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.GetID(), CourseID: c1.GetID()}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.GetID(), CourseID: c2.GetID()}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.GetID(), CourseID: c3.GetID()}); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(user.GetID(), c2.GetID()); err != nil {
		t.Fatal(err)
	}
	query, err := db.GetEnrollmentByCourseAndUser(c3.GetID(), user.GetID())
	if err != nil {
		t.Fatal(err)
	}
	query.Status = qf.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	gotCourses, err := db.GetCoursesByUser(user.GetID())
	if err != nil {
		t.Fatal(err)
	}

	wantCourses := []*qf.Course{
		{ID: c1.GetID(), ScmOrganizationID: 1, Code: "DAT101", Year: 1, CourseCreatorID: admin.GetID(), Enrolled: qf.Enrollment_PENDING},
		{ID: c2.GetID(), ScmOrganizationID: 2, Code: "DAT101", Year: 2, CourseCreatorID: admin.GetID(), Enrolled: qf.Enrollment_NONE},
		{ID: c3.GetID(), ScmOrganizationID: 3, Code: "DAT101", Year: 3, CourseCreatorID: admin.GetID(), Enrolled: qf.Enrollment_STUDENT},
		{ID: c4.GetID(), ScmOrganizationID: 4, Code: "DAT101", Year: 4, CourseCreatorID: admin.GetID(), Enrolled: qf.Enrollment_NONE},
	}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCoursesByUser() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}
}

func TestBunDBCreateCourseNonAdmin(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, &qf.Course{})

	nonAdmin := qtest.CreateFakeUser(t, db)
	if err := db.CreateCourse(nonAdmin.GetID(), &qf.Course{}); err == nil {
		t.Fatal("non-admin user should not be able to create a course")
	}
}

func TestBunDBGetCourses(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	c1 := &qf.Course{ScmOrganizationID: 1, Code: "DAT101", Year: 1}
	c2 := &qf.Course{ScmOrganizationID: 2, Code: "DAT101", Year: 2}
	c3 := &qf.Course{ScmOrganizationID: 3, Code: "DAT101", Year: 3}
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

	gotCourses, err = db.GetCourses([]uint64{}...)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	gotCourses, err = db.GetCourses(c1.GetID())
	if err != nil {
		t.Fatal(err)
	}
	wantCourses = []*qf.Course{c1}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}

	gotCourses, err = db.GetCourses(c1.GetID(), c2.GetID())
	if err != nil {
		t.Fatal(err)
	}
	wantCourses = []*qf.Course{c1, c2}
	if diff := cmp.Diff(wantCourses, gotCourses, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourses() mismatch (-wantCourses, +gotCourses):\n%s", diff)
	}
}

func TestBunDBGetCourse(t *testing.T) {
	wantCourse := &qf.Course{
		Name:              "Test Course",
		Code:              "DAT100",
		Year:              2017,
		Tag:               "Spring",
		ScmOrganizationID: 1234,
	}

	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, wantCourse)

	gotCourse, err := db.GetCourse(wantCourse.GetID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse, +gotCourse):\n%s", diff)
	}
}

func TestBunDBGetCourseNoRecord(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if _, err := db.GetCourse(10); !isNotFound(err) {
		t.Errorf("have error '%v' wanted not-found", err)
	}
}

func TestBunDBUpdateCourse(t *testing.T) {
	course := &qf.Course{
		Name:              "Test Course",
		Code:              "DAT100",
		Year:              2017,
		Tag:               "Spring",
		DockerfileDigest:  "0x123abc",
		ScmOrganizationID: 1234,
	}
	wantCourse := &qf.Course{
		Name:              "Test Course Edit",
		Code:              "DAT100-1",
		Year:              2018,
		Tag:               "Autumn",
		DockerfileDigest:  "0x123def",
		ScmOrganizationID: 12345,
	}

	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	wantCourse.ID = course.GetID()
	wantCourse.CourseCreatorID = admin.GetID()
	if err := db.UpdateCourse(wantCourse); err != nil {
		t.Fatal(err)
	}

	gotCourse, err := db.GetCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("course mismatch (-want +got):\n%s", diff)
	}
}

func TestBunDBGetCourseByOrganization(t *testing.T) {
	wantCourse := &qf.Course{
		Name:              "Test Course",
		Code:              "DAT100",
		Year:              2017,
		Tag:               "Spring",
		ScmOrganizationID: 1234,
	}

	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, wantCourse)

	gotCourse, err := db.GetCourseByOrganizationID(wantCourse.GetScmOrganizationID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourse() mismatch (-wantCourse, +gotCourse):\n%s", diff)
	}
}

func TestBunDBCourseUniqueConstraint(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	wantCourse := &qf.Course{
		Name:              "Test Course",
		Code:              "DAT100",
		Year:              2017,
		Tag:               "Spring",
		ScmOrganizationID: 1235,
	}
	course := &qf.Course{
		Name:              "Test Course 2",
		Code:              "DAT100",
		Year:              2017,
		Tag:               "Spring",
		ScmOrganizationID: 1234,
	}

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, wantCourse)

	if err := db.CreateCourse(admin.GetID(), course); err != nil && !errors.Is(err, database.ErrCourseExists) {
		t.Fatal(err)
	}

	if err := db.CreateCourse(admin.GetID(), &qf.Course{ScmOrganizationID: wantCourse.GetScmOrganizationID()}); err != nil && !errors.Is(err, database.ErrCourseExists) {
		t.Fatal(err)
	}

	gotCourse, err := db.GetCourse(wantCourse.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantCourse, gotCourse, protocmp.Transform()); diff != "" {
		t.Errorf("course mismatch (-want +got):\n%s", diff)
	}

	course.Year = 2018
	qtest.CreateCourse(t, db, admin, course)
}

func TestBunGetCourseByStatus(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	course := &qf.Course{
		ID: 1,
	}
	adminUser := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, adminUser, course)
	teacherEnrollment := &qf.Enrollment{
		ID:       1,
		UserID:   adminUser.GetID(),
		CourseID: course.GetID(),
		Status:   qf.Enrollment_TEACHER,
	}
	studentEnrollment := qtest.EnrollUser(t, db, qtest.CreateFakeUser(t, db), course, qf.Enrollment_STUDENT)
	pendingEnrollment := qtest.EnrollUser(t, db, qtest.CreateFakeUser(t, db), course, qf.Enrollment_PENDING)
	noneEnrollment := qtest.EnrollUser(t, db, qtest.CreateFakeUser(t, db), course, qf.Enrollment_NONE)

	type args struct {
		courseID uint64
		status   qf.Enrollment_UserStatus
	}
	tests := []struct {
		name    string
		args    args
		want    *qf.Course
		wantErr bool
	}{
		{
			name: "none: no preloaded data",
			args: args{
				courseID: course.GetID(),
				status:   qf.Enrollment_NONE,
			},
			want: course,
		},
		{
			name: "pending: no preloaded data",
			args: args{
				courseID: course.GetID(),
				status:   qf.Enrollment_PENDING,
			},
			want: course,
		},
		{
			name: "student: preloaded assignments, active enrollments and groups",
			args: args{
				courseID: course.GetID(),
				status:   qf.Enrollment_STUDENT,
			},
			want: &qf.Course{
				Enrollments: []*qf.Enrollment{
					teacherEnrollment,
					studentEnrollment,
				},
			},
		},
		{
			name: "teacher: preloaded assignments, active enrollments and groups with detailed information",
			args: args{
				courseID: course.GetID(),
				status:   qf.Enrollment_TEACHER,
			},
			want: &qf.Course{
				Enrollments: []*qf.Enrollment{
					teacherEnrollment,
					studentEnrollment,
					pendingEnrollment,
					noneEnrollment,
				},
			},
		},
		{
			name: "invalid status",
			args: args{
				courseID: course.GetID(),
				status:   qf.Enrollment_UserStatus(10),
			},
			wantErr: true,
		},
		{
			name: "no course",
			args: args{
				courseID: 10,
				status:   qf.Enrollment_TEACHER,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetCourseByStatus(tt.args.courseID, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("BunDB.GetCourseByStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform(),
				protocmp.IgnoreFields(&qf.Enrollment{}, "user", "state"),
				protocmp.IgnoreFields(&qf.Course{}, "courseCreatorID", "ID"),
			); diff != "" {
				t.Errorf("BunDB.GetCourseByStatus() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBunGetCourseTeachers(t *testing.T) {
	tests := map[string]struct {
		wantTeachers, students []*qf.User
	}{
		"Basic": {
			wantTeachers: []*qf.User{
				{Login: "teacher1", Name: "Teacher One", Email: "teacher1@example.com", StudentID: "T001", ScmRemoteID: 1001},
				{Login: "teacher2", Name: "Teacher Two", Email: "teacher2@example.com", StudentID: "T002", ScmRemoteID: 1002},
			},
			students: []*qf.User{
				{Login: "student1", Name: "Student One", Email: "student1@example.com", StudentID: "S001", ScmRemoteID: 2001},
			},
		},
		"No teachers": {
			wantTeachers: []*qf.User{},
			students: []*qf.User{
				{Login: "student1", Name: "Student One", Email: "student1@example.com", StudentID: "S001", ScmRemoteID: 2002},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, cleanup := qtest.TestBunDB(t)
			defer cleanup()
			admin := qtest.CreateFakeUser(t, db)
			course := &qf.Course{}
			qtest.CreateCourse(t, db, admin, course)
			for _, teacher := range tt.wantTeachers {
				if err := db.CreateUser(teacher); err != nil {
					t.Error(err)
				}
				qtest.EnrollTeacher(t, db, teacher, course)
			}
			for _, student := range tt.students {
				if err := db.CreateUser(student); err != nil {
					t.Error(err)
				}
				qtest.EnrollStudent(t, db, student, course)
			}
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
