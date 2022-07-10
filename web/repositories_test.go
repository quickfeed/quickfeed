package web

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGetRepo(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{
		OrganizationID: 1,
		Code:           "DAT101",
	}
	qtest.CreateCourse(t, db, user, course)
	group := &qf.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.ID,
		Users:    []*qf.User{user},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	wantUserRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         user.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "http://assignment.com/",
	}
	if err := db.CreateRepository(wantUserRepo); err != nil {
		t.Fatal(err)
	}

	wantGroupRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   2,
		GroupID:        group.ID,
		RepoType:       qf.Repository_GROUP,
		HTMLURL:        "http://assignment.com/",
	}
	if err := db.CreateRepository(wantGroupRepo); err != nil {
		t.Fatal(err)
	}

	_, scms := qtest.FakeProviderMap(t)
	ags := NewQuickFeedService(zap.NewNop(), db, scms, BaseHookOptions{}, &ci.Local{})
	gotUserRepo, err := ags.getRepo(course, user.ID, qf.Repository_USER)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantUserRepo, gotUserRepo, protocmp.Transform()); diff != "" {
		t.Errorf("getRepo() mismatch (-wantUserRepo, +gotUserRepo):\n%s", diff)
	}

	gotGroupRepo, err := ags.getRepo(course, group.ID, qf.Repository_GROUP)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantGroupRepo, gotGroupRepo, protocmp.Transform()); diff != "" {
		t.Errorf("getRepo() mismatch (-wantGroupRepo, +gotGroupRepo):\n%s", diff)
	}

	_, err = ags.getRepo(course, group.ID+1, qf.Repository_GROUP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
	_, err = ags.getRepo(course, user.ID+1, qf.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
}

func TestGetRepositories(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{
		OrganizationID: 1,
		Code:           "DAT101",
	}
	qtest.CreateCourse(t, db, user, course)

	_, scms := qtest.FakeProviderMap(t)
	ags := NewQuickFeedService(zap.NewNop(), db, scms, BaseHookOptions{}, &ci.Local{})
	ctx := qtest.WithUserContext(context.Background(), user)

	// check that no repositories are returned when no repo types are specified
	repos, err := ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos.URLs) != 0 {
		t.Errorf("GetRepositories() got %v, want none", repos.URLs)
	}

	// check that empty user repository is returned before user repository has been created
	gotUserRepoURLs, err := ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantUserRepoURLs := &qf.Repositories{
		URLs: map[string]string{"USER": ""}, // no user repository exists yet
	}
	if diff := cmp.Diff(wantUserRepoURLs, gotUserRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserRepoURLs, +gotUserRepoURLs):\n%s", diff)
	}

	wantUserRepo := &qf.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         user.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "http://user.assignment.com/",
	}
	if err := db.CreateRepository(wantUserRepo); err != nil {
		t.Fatal(err)
	}

	// check that no repositories are returned when no repo types are specified
	repos, err = ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos.URLs) != 0 {
		t.Errorf("GetRepositories() got %v, want none", repos.URLs)
	}

	// check that user repository is returned when user repo type is specified
	gotUserRepoURLs, err = ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantUserRepoURLs = &qf.Repositories{
		URLs: map[string]string{"USER": wantUserRepo.HTMLURL},
	}
	if diff := cmp.Diff(wantUserRepoURLs, gotUserRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserRepoURLs, +gotUserRepoURLs):\n%s", diff)
	}

	// try to get group repository before group exists (user not enrolled in group)
	gotGroupRepoURLs, err := ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_GROUP,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantGroupRepoURLs := &qf.Repositories{
		URLs: map[string]string{"GROUP": ""}, // no group repository exists yet
	}
	if diff := cmp.Diff(wantGroupRepoURLs, gotGroupRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantGroupRepoURLs, +gotGroupRepoURLs):\n%s", diff)
	}

	group := &qf.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.ID,
		Users:    []*qf.User{user},
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
	gotGroupRepoURLs, err = ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_GROUP,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantGroupRepoURLs = &qf.Repositories{
		URLs: map[string]string{"GROUP": wantGroupRepo.HTMLURL},
	}
	if diff := cmp.Diff(wantGroupRepoURLs, gotGroupRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantGroupRepoURLs, +gotGroupRepoURLs):\n%s", diff)
	}

	// check that both user and group repositories are returned when both repo types are specified
	gotUserGroupRepoURLs, err := ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
			qf.Repository_GROUP,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantUserGroupRepoURLs := &qf.Repositories{
		URLs: map[string]string{
			"USER":  wantUserRepo.HTMLURL,
			"GROUP": wantGroupRepo.HTMLURL,
		},
	}
	if diff := cmp.Diff(wantUserGroupRepoURLs, gotUserGroupRepoURLs, protocmp.Transform()); diff != "" {
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
	gotAllRepoURLs, err := ags.GetRepositories(ctx, &qf.URLRequest{
		CourseID: course.ID,
		RepoTypes: []qf.Repository_Type{
			qf.Repository_USER,
			qf.Repository_GROUP,
			qf.Repository_INFO,
			qf.Repository_ASSIGNMENTS,
			qf.Repository_TESTS,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantAllRepoURLs := &qf.Repositories{
		URLs: map[string]string{
			"ASSIGNMENTS": wantAssignmentsRepo.HTMLURL,
			"COURSEINFO":  wantInfoRepo.HTMLURL,
			"TESTS":       wantTestsRepo.HTMLURL,
			"USER":        wantUserRepo.HTMLURL,
			"GROUP":       wantGroupRepo.HTMLURL,
		},
	}
	if diff := cmp.Diff(wantAllRepoURLs, gotAllRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantAllRepoURLs, +gotAllRepoURLs):\n%s", diff)
	}
}
