package web_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetRepositories(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	teacher := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, teacher, course)
	// student, not in a group
	student := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student, course)
	// student, in a group
	groupStudent := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, groupStudent, course)
	group := &qf.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.GetID(),
		Users:    []*qf.User{groupStudent},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}
	// user, not enrolled in the course
	notEnrolledUser := qtest.CreateFakeUser(t, db)

	// create repositories for users and group
	teacherRepo := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		ScmRepositoryID:   1,
		UserID:            teacher.GetID(),
		HTMLURL:           "teacher.repo",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(teacherRepo); err != nil {
		t.Fatal(err)
	}
	studentRepo := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		ScmRepositoryID:   2,
		UserID:            student.GetID(),
		HTMLURL:           "student.repo",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(studentRepo); err != nil {
		t.Fatal(err)
	}
	groupStudentRepo := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		ScmRepositoryID:   3,
		UserID:            groupStudent.GetID(),
		HTMLURL:           "group.student.repo",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(groupStudentRepo); err != nil {
		t.Fatal(err)
	}
	groupRepo := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
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
		ScmOrganizationID: course.GetScmOrganizationID(),
		HTMLURL:           "course.info",
		RepoType:          qf.Repository_INFO,
	}
	if err := db.CreateRepository(info); err != nil {
		t.Fatal(err)
	}
	assignments := &qf.Repository{
		ScmRepositoryID:   6,
		ScmOrganizationID: course.GetScmOrganizationID(),
		HTMLURL:           "course.assignments",
		RepoType:          qf.Repository_ASSIGNMENTS,
	}
	if err := db.CreateRepository(assignments); err != nil {
		t.Fatal(err)
	}
	testRepo := &qf.Repository{
		ScmRepositoryID:   7,
		ScmOrganizationID: course.GetScmOrganizationID(),
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
			courseID:  course.GetID(),
			cookie:    missingEnrollmentCookie,
			wantRepos: nil,
			wantErr:   true,
		},
		{
			name:     "course teacher",
			courseID: course.GetID(),
			cookie:   teacherCookie,
			wantRepos: &qf.Repositories{
				URLs: map[uint32]string{
					uint32(qf.Repository_ASSIGNMENTS): assignments.GetHTMLURL(),
					uint32(qf.Repository_INFO):        info.GetHTMLURL(),
					uint32(qf.Repository_TESTS):       testRepo.GetHTMLURL(),
					uint32(qf.Repository_USER):        teacherRepo.GetHTMLURL(),
				},
			},
			wantErr: false,
		},
		{
			name:     "course student, not in a group",
			courseID: course.GetID(),
			cookie:   studentCookie,
			wantRepos: &qf.Repositories{
				URLs: map[uint32]string{
					uint32(qf.Repository_ASSIGNMENTS): assignments.GetHTMLURL(),
					uint32(qf.Repository_INFO):        info.GetHTMLURL(),
					uint32(qf.Repository_USER):        studentRepo.GetHTMLURL(),
				},
			},
			wantErr: false,
		},
		{
			name:     "course student, in a group",
			courseID: course.GetID(),
			cookie:   groupStudentCookie,
			wantRepos: &qf.Repositories{
				URLs: map[uint32]string{
					uint32(qf.Repository_ASSIGNMENTS): assignments.GetHTMLURL(),
					uint32(qf.Repository_INFO):        info.GetHTMLURL(),
					uint32(qf.Repository_USER):        groupStudentRepo.GetHTMLURL(),
					uint32(qf.Repository_GROUP):       groupRepo.GetHTMLURL(),
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

func TestQuickFeedService_IsEmptyRepo(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)

	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user"})
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, user, course)

	student := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "student"})
	qtest.EnrollStudent(t, db, student, course)

	// student in a group
	groupStudent := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "groupStudent"})
	qtest.EnrollStudent(t, db, groupStudent, course)

	// create repositories for users and group
	userRepo := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		ScmRepositoryID:   1,
		UserID:            user.GetID(), // 1
		HTMLURL:           "user",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(userRepo); err != nil {
		t.Fatal(err)
	}
	group := &qf.Group{
		ID:       1,
		Name:     "1001-HackingCrew",
		CourseID: course.GetID(),
		Users:    []*qf.User{groupStudent},
	}
	g, err := client.CreateGroup(context.Background(), qtest.RequestWithCookie(group, "cookie"))
	if err != nil {
		t.Fatal(err)
	}
	group = g.Msg

	tests := []struct {
		name    string
		request *qf.RepositoryRequest
		create  bool
		wantErr bool
	}{
		{name: "CourseNotFound", request: &qf.RepositoryRequest{CourseID: 123, UserID: user.GetID()}, wantErr: true},    // unable to get SCM client for unknown course -> error
		{name: "UserNotFound", request: &qf.RepositoryRequest{CourseID: course.GetID(), UserID: 123}, wantErr: false},   // lookup invalid user should have no repositories (no error)
		{name: "GroupNotFound", request: &qf.RepositoryRequest{CourseID: course.GetID(), GroupID: 123}, wantErr: false}, // lookup invalid group should have no repositories (no error)

		{name: "UserHasNoRepositories", request: &qf.RepositoryRequest{CourseID: 1, UserID: student.GetID()}, wantErr: false},                    // lookup valid user with no repositories should return no repositories (no error)
		{name: "GroupHasNoRepositories", request: &qf.RepositoryRequest{CourseID: course.GetID(), GroupID: group.GetID()}, wantErr: false},            // lookup valid group with no repositories should return no repositories (no error)
		{name: "GroupHasRepositories", request: &qf.RepositoryRequest{CourseID: course.GetID(), GroupID: group.GetID()}, create: true, wantErr: true}, // lookup for group with repositories -> error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.create {
				// trigger group repository creation on SCM
				group.Status = qf.Group_APPROVED
				group.Users = append(group.GetUsers(), user)
				if _, err := client.UpdateGroup(context.Background(), qtest.RequestWithCookie(group, "cookie")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := client.IsEmptyRepo(context.Background(), qtest.RequestWithCookie(tt.request, "cookie")); (err != nil) != tt.wantErr {
				t.Errorf("IsEmptyRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
