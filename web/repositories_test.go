package web

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGetRepo(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := qtest.CreateFakeUser(t, db, 1)
	course := &types.Course{
		OrganizationID: 1,
		Code:           "DAT101",
	}
	qtest.CreateCourse(t, db, user, course)
	group := &types.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.ID,
		Users:    []*types.User{user},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	wantUserRepo := &types.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         user.ID,
		RepoType:       types.Repository_USER,
		HTMLURL:        "http://assignment.com/",
	}
	if err := db.CreateRepository(wantUserRepo); err != nil {
		t.Fatal(err)
	}

	wantGroupRepo := &types.Repository{
		OrganizationID: 1,
		RepositoryID:   2,
		GroupID:        group.ID,
		RepoType:       types.Repository_GROUP,
		HTMLURL:        "http://assignment.com/",
	}
	if err := db.CreateRepository(wantGroupRepo); err != nil {
		t.Fatal(err)
	}

	_, scms := qtest.FakeProviderMap(t)
	ags := NewQuickFeedService(zap.NewNop(), db, scms, BaseHookOptions{}, &ci.Local{})
	gotUserRepo, err := ags.getRepo(course, user.ID, types.Repository_USER)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantUserRepo, gotUserRepo, protocmp.Transform()); diff != "" {
		t.Errorf("getRepo() mismatch (-wantUserRepo, +gotUserRepo):\n%s", diff)
	}

	gotGroupRepo, err := ags.getRepo(course, group.ID, types.Repository_GROUP)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantGroupRepo, gotGroupRepo, protocmp.Transform()); diff != "" {
		t.Errorf("getRepo() mismatch (-wantGroupRepo, +gotGroupRepo):\n%s", diff)
	}

	_, err = ags.getRepo(course, group.ID+1, types.Repository_GROUP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
	_, err = ags.getRepo(course, user.ID+1, types.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
}

func TestGetRepositories(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := qtest.CreateFakeUser(t, db, 1)
	course := &types.Course{
		OrganizationID: 1,
		Code:           "DAT101",
	}
	qtest.CreateCourse(t, db, user, course)

	_, scms := qtest.FakeProviderMap(t)
	ags := NewQuickFeedService(zap.NewNop(), db, scms, BaseHookOptions{}, &ci.Local{})
	ctx := qtest.WithUserContext(context.Background(), user)

	// check that no repositories are returned when no repo types are specified
	repos, err := ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos.URLs) != 0 {
		t.Errorf("GetRepositories() got %v, want none", repos.URLs)
	}

	// check that empty user repository is returned before user repository has been created
	gotUserRepoURLs, err := ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
		RepoTypes: []types.Repository_Type{
			types.Repository_USER,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantUserRepoURLs := &types.Repositories{
		URLs: map[string]string{"USER": ""}, // no user repository exists yet
	}
	if diff := cmp.Diff(wantUserRepoURLs, gotUserRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserRepoURLs, +gotUserRepoURLs):\n%s", diff)
	}

	wantUserRepo := &types.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         user.ID,
		RepoType:       types.Repository_USER,
		HTMLURL:        "http://user.assignment.com/",
	}
	if err := db.CreateRepository(wantUserRepo); err != nil {
		t.Fatal(err)
	}

	// check that no repositories are returned when no repo types are specified
	repos, err = ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos.URLs) != 0 {
		t.Errorf("GetRepositories() got %v, want none", repos.URLs)
	}

	// check that user repository is returned when user repo type is specified
	gotUserRepoURLs, err = ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
		RepoTypes: []types.Repository_Type{
			types.Repository_USER,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantUserRepoURLs = &types.Repositories{
		URLs: map[string]string{"USER": wantUserRepo.HTMLURL},
	}
	if diff := cmp.Diff(wantUserRepoURLs, gotUserRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserRepoURLs, +gotUserRepoURLs):\n%s", diff)
	}

	// try to get group repository before group exists (user not enrolled in group)
	gotGroupRepoURLs, err := ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
		RepoTypes: []types.Repository_Type{
			types.Repository_GROUP,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantGroupRepoURLs := &types.Repositories{
		URLs: map[string]string{"GROUP": ""}, // no group repository exists yet
	}
	if diff := cmp.Diff(wantGroupRepoURLs, gotGroupRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantGroupRepoURLs, +gotGroupRepoURLs):\n%s", diff)
	}

	group := &types.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.ID,
		Users:    []*types.User{user},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	wantGroupRepo := &types.Repository{
		OrganizationID: 1,
		RepositoryID:   2,
		GroupID:        group.ID,
		RepoType:       types.Repository_GROUP,
		HTMLURL:        "http://group.assignment.com/",
	}
	if err := db.CreateRepository(wantGroupRepo); err != nil {
		t.Fatal(err)
	}

	// check that group repository is returned when group repo type is specified
	gotGroupRepoURLs, err = ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
		RepoTypes: []types.Repository_Type{
			types.Repository_GROUP,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantGroupRepoURLs = &types.Repositories{
		URLs: map[string]string{"GROUP": wantGroupRepo.HTMLURL},
	}
	if diff := cmp.Diff(wantGroupRepoURLs, gotGroupRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantGroupRepoURLs, +gotGroupRepoURLs):\n%s", diff)
	}

	// check that both user and group repositories are returned when both repo types are specified
	gotUserGroupRepoURLs, err := ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
		RepoTypes: []types.Repository_Type{
			types.Repository_USER,
			types.Repository_GROUP,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantUserGroupRepoURLs := &types.Repositories{
		URLs: map[string]string{
			"USER":  wantUserRepo.HTMLURL,
			"GROUP": wantGroupRepo.HTMLURL,
		},
	}
	if diff := cmp.Diff(wantUserGroupRepoURLs, gotUserGroupRepoURLs, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantUserGroupRepoURLs, +gotUserGroupRepoURLs):\n%s", diff)
	}

	wantAssignmentsRepo := &types.Repository{
		OrganizationID: 1,
		RepositoryID:   3,
		RepoType:       types.Repository_ASSIGNMENTS,
		HTMLURL:        "http://assignments.assignment.com/",
	}
	if err := db.CreateRepository(wantAssignmentsRepo); err != nil {
		t.Fatal(err)
	}
	wantInfoRepo := &types.Repository{
		OrganizationID: 1,
		RepositoryID:   4,
		RepoType:       types.Repository_COURSEINFO,
		HTMLURL:        "http://info.assignment.com/",
	}
	if err := db.CreateRepository(wantInfoRepo); err != nil {
		t.Fatal(err)
	}
	wantTestsRepo := &types.Repository{
		OrganizationID: 1,
		RepositoryID:   5,
		RepoType:       types.Repository_TESTS,
		HTMLURL:        "http://tests.assignment.com/",
	}
	if err := db.CreateRepository(wantTestsRepo); err != nil {
		t.Fatal(err)
	}

	// check that all repositories are returned when all repo types are specified
	gotAllRepoURLs, err := ags.GetRepositories(ctx, &types.URLRequest{
		CourseID: course.ID,
		RepoTypes: []types.Repository_Type{
			types.Repository_USER,
			types.Repository_GROUP,
			types.Repository_COURSEINFO,
			types.Repository_ASSIGNMENTS,
			types.Repository_TESTS,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantAllRepoURLs := &types.Repositories{
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
