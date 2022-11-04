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

	client, tm, _ := MockClientWithUser(t, db, false)

	teacher := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{
		OrganizationID: 1,
		Code:           "DAT101",
	}
	qtest.CreateCourse(t, db, teacher, course)
	cookie := Cookie(t, tm, teacher)

	ctx := context.Background()
	// check that no repositories are returned when no repo types are specified
	repos, err := client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	if len(repos.Msg.URLs) != 0 {
		t.Errorf("GetRepositories() got %v, want none", repos.Msg.URLs)
	}

	// check that empty user repository is returned before user repository has been created
	gotUserRepoURLs, err := client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	wantUserRepoURLs := &qf.Repositories{
		URLs: map[string]string{"USER": ""}, // no user repository exists yet
	}
	if diff := cmp.Diff(wantUserRepoURLs, gotUserRepoURLs.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserRepoURLs, +gotUserRepoURLs):\n%s", diff)
	}

	wantUserRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         teacher.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "http://user.assignment.com/",
	}
	if err := db.CreateRepository(wantUserRepo); err != nil {
		t.Fatal(err)
	}

	// check that no repositories are returned when no repo types are specified
	repos, err = client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	if len(repos.Msg.URLs) != 0 {
		t.Errorf("GetRepositories() got %v, want none", repos.Msg.URLs)
	}

	// check that user repository is returned when user repo type is specified
	gotUserRepoURLs, err = client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	wantUserRepoURLs = &qf.Repositories{
		URLs: map[string]string{"USER": wantUserRepo.HTMLURL},
	}
	if diff := cmp.Diff(wantUserRepoURLs, gotUserRepoURLs.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserRepoURLs, +gotUserRepoURLs):\n%s", diff)
	}

	// try to get group repository before group exists (user not enrolled in group)
	gotGroupRepoURLs, err := client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_GROUP,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	wantGroupRepoURLs := &qf.Repositories{
		URLs: map[string]string{"GROUP": ""}, // no group repository exists yet
	}
	if diff := cmp.Diff(wantGroupRepoURLs, gotGroupRepoURLs.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantGroupRepoURLs, +gotGroupRepoURLs):\n%s", diff)
	}

	group := &qf.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.ID,
		Users:    []*qf.User{teacher},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	wantGroupRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   2,
		GroupID:        group.ID,
		RepoType:       qf.Repository_GROUP,
		HTMLURL:        "http://group.assignment.com/",
	}
	if err := db.CreateRepository(wantGroupRepo); err != nil {
		t.Fatal(err)
	}

	// check that group repository is returned when group repo type is specified
	gotGroupRepoURLs, err = client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_GROUP,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	wantGroupRepoURLs = &qf.Repositories{
		URLs: map[string]string{"GROUP": wantGroupRepo.HTMLURL},
	}
	if diff := cmp.Diff(wantGroupRepoURLs, gotGroupRepoURLs.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantGroupRepoURLs, +gotGroupRepoURLs):\n%s", diff)
	}

	// check that both user and group repositories are returned when both repo types are specified
	gotUserGroupRepoURLs, err := client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
			qf.Repository_GROUP,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	wantUserGroupRepoURLs := &qf.Repositories{
		URLs: map[string]string{
			"USER":  wantUserRepo.HTMLURL,
			"GROUP": wantGroupRepo.HTMLURL,
		},
	}
	if diff := cmp.Diff(wantUserGroupRepoURLs, gotUserGroupRepoURLs.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserGroupRepoURLs, +gotUserGroupRepoURLs):\n%s", diff)
	}

	wantAssignmentsRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   3,
		RepoType:       qf.Repository_ASSIGNMENTS,
		HTMLURL:        "http://assignments.assignment.com/",
	}
	if err := db.CreateRepository(wantAssignmentsRepo); err != nil {
		t.Fatal(err)
	}
	wantInfoRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   4,
		RepoType:       qf.Repository_INFO,
		HTMLURL:        "http://info.assignment.com/",
	}
	if err := db.CreateRepository(wantInfoRepo); err != nil {
		t.Fatal(err)
	}
	wantTestsRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   5,
		RepoType:       qf.Repository_TESTS,
		HTMLURL:        "http://tests.assignment.com/",
	}
	if err := db.CreateRepository(wantTestsRepo); err != nil {
		t.Fatal(err)
	}

	// check that all repositories are returned when all repo types are specified
	gotAllRepoURLs, err := client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
			qf.Repository_GROUP,
			qf.Repository_INFO,
			qf.Repository_ASSIGNMENTS,
			qf.Repository_TESTS,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	wantAllRepoURLs := &qf.Repositories{
		URLs: map[string]string{
			"ASSIGNMENTS": wantAssignmentsRepo.HTMLURL,
			"INFO":        wantInfoRepo.HTMLURL,
			"TESTS":       wantTestsRepo.HTMLURL,
			"USER":        wantUserRepo.HTMLURL,
			"GROUP":       wantGroupRepo.HTMLURL,
		},
	}
	if diff := cmp.Diff(wantAllRepoURLs, gotAllRepoURLs.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantAllRepoURLs, +gotAllRepoURLs):\n%s", diff)
	}
}
