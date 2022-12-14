package web_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetRepositories(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm, _ := MockClientWithUser(t, db)

	teacher := qtest.CreateFakeUser(t, db, 1)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)
	// student, not in a group
	student := qtest.CreateFakeUser(t, db, 2)
	qtest.EnrollStudent(t, db, student, course)
	// student, in a group
	groupStudent := qtest.CreateFakeUser(t, db, 3)
	qtest.EnrollStudent(t, db, groupStudent, course)
	group := &qf.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.ID,
		Users:    []*qf.User{groupStudent},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}
	// user, not enrolled in the course
	notEnrolledUser := qtest.CreateFakeUser(t, db, 5)

	// create repositories for users and group
	teacherRepo := &qf.Repository{
		ScmOrganizationID: course.ScmOrganizationID,
		ScmRepositoryID:   1,
		UserID:            teacher.ID,
		HTMLURL:           "teacher.repo",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(teacherRepo); err != nil {
		t.Fatal(err)
	}
	studentRepo := &qf.Repository{
		ScmOrganizationID: course.ScmOrganizationID,
		ScmRepositoryID:   2,
		UserID:            student.ID,
		HTMLURL:           "student.repo",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(studentRepo); err != nil {
		t.Fatal(err)
	}
	groupStudentRepo := &qf.Repository{
		ScmOrganizationID: course.ScmOrganizationID,
		ScmRepositoryID:   3,
		UserID:            groupStudent.ID,
		HTMLURL:           "group.student.repo",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(groupStudentRepo); err != nil {
		t.Fatal(err)
	}
	groupRepo := &qf.Repository{
		ScmOrganizationID: course.ScmOrganizationID,
		ScmRepositoryID:   4,
		GroupID:           1,
		HTMLURL:           "group.repo",
		RepoType:          qf.Repository_GROUP,
	}
	if err := db.CreateRepository(groupRepo); err != nil {
		t.Fatal(err)
	}

	// create course repositories
	info := &qf.Repository{
		ScmRepositoryID:   5,
		ScmOrganizationID: course.ScmOrganizationID,
		HTMLURL:           "course.info",
		RepoType:          qf.Repository_INFO,
	}
	if err := db.CreateRepository(info); err != nil {
		t.Fatal(err)
	}
	assignments := &qf.Repository{
		ScmRepositoryID:   6,
		ScmOrganizationID: course.ScmOrganizationID,
		HTMLURL:           "course.assignments",
		RepoType:          qf.Repository_ASSIGNMENTS,
	}
	if err := db.CreateRepository(assignments); err != nil {
		t.Fatal(err)
	}
	testRepo := &qf.Repository{
		ScmRepositoryID:   7,
		ScmOrganizationID: course.ScmOrganizationID,
		HTMLURL:           "course.tests",
		RepoType:          qf.Repository_TESTS,
	}
	if err := db.CreateRepository(testRepo); err != nil {
		t.Fatal(err)
	}

	teacherCookie := Cookie(t, tm, teacher)
	studentCookie := Cookie(t, tm, student)
	groupStudentCookie := Cookie(t, tm, groupStudent)
	missingEnrollmentCookie := Cookie(t, tm, notEnrolledUser)

	ctx := context.Background()

	tests := []struct {
		name      string
		courseID  uint64
		cookie    string
		wantRepos *qf.Repositories
		wantErr   bool
	}{
		{
			name:      "incorrect course ID",
			courseID:  123,
			cookie:    teacherCookie,
			wantRepos: nil,
			wantErr:   true,
		},
		{
			name:      "user without course enrollment",
			courseID:  course.ID,
			cookie:    missingEnrollmentCookie,
			wantRepos: nil,
			wantErr:   true,
		},
		{
			name:     "course teacher",
			courseID: course.ID,
			cookie:   teacherCookie,
			wantRepos: &qf.Repositories{
				URLs: map[uint32]string{
					uint32(qf.Repository_ASSIGNMENTS): assignments.HTMLURL,
					uint32(qf.Repository_INFO):        info.HTMLURL,
					uint32(qf.Repository_TESTS):       testRepo.HTMLURL,
					uint32(qf.Repository_USER):        teacherRepo.HTMLURL,
				},
			},
			wantErr: false,
		},
		{
			name:     "course student, not in a group",
			courseID: course.ID,
			cookie:   studentCookie,
			wantRepos: &qf.Repositories{
				URLs: map[uint32]string{
					uint32(qf.Repository_ASSIGNMENTS): assignments.HTMLURL,
					uint32(qf.Repository_INFO):        info.HTMLURL,
					uint32(qf.Repository_USER):        studentRepo.HTMLURL,
				},
			},
			wantErr: false,
		},
		{
			name:     "course student, in a group",
			courseID: course.ID,
			cookie:   groupStudentCookie,
			wantRepos: &qf.Repositories{
				URLs: map[uint32]string{
					uint32(qf.Repository_ASSIGNMENTS): assignments.HTMLURL,
					uint32(qf.Repository_INFO):        info.HTMLURL,
					uint32(qf.Repository_USER):        groupStudentRepo.HTMLURL,
					uint32(qf.Repository_GROUP):       groupRepo.HTMLURL,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		resp, err := client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.CourseRequest{
			CourseID: tt.courseID,
		}, tt.cookie))
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
		if !tt.wantErr {
			if diff := cmp.Diff(tt.wantRepos, resp.Msg, protocmp.Transform()); diff != "" {
				t.Errorf("%s mismatch repositories (-want +got):\n%s", tt.name, diff)
			}
		}
	}
}
