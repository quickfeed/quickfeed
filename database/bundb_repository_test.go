package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestBunDBGetEmptyRepo(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	repos, err := db.GetRepositories(&qf.Repository{ScmRepositoryID: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 0 {
		t.Fatalf("Expected no repositories, but got: %v", repos)
	}
}

func TestBunDBGetSingleRepoWithUser(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	user := qtest.CreateFakeUser(t, db)
	repo := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   100,
		UserID:            user.GetID(),
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetRepositories(&qf.Repository{ScmRepositoryID: repo.GetScmRepositoryID()}); err != nil {
		t.Fatal(err)
	}
}

func TestBunDBCreateSingleRepoWithMissingUser(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	repo := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   100,
		UserID:            20,
	}
	if err := db.CreateRepository(&repo); !isNotFound(err) {
		t.Fatalf("have error '%v' wanted sql.ErrNoRows", err)
	}
}

func TestBunDBGetCourseRepoType(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	repo := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   100,
		RepoType:          qf.Repository_INFO,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	gotRepos, err := db.GetRepositories(&qf.Repository{ScmRepositoryID: repo.GetScmRepositoryID()})
	if err != nil {
		t.Fatal(err)
	}
	if !gotRepos[0].GetRepoType().IsCourseRepo() {
		t.Fatalf("Expected course info repo (%v), but got: %v", qf.Repository_INFO, gotRepos[0].GetRepoType())
	}
}

func TestBunDeleteRepo(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	repo := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   100,
		RepoType:          qf.Repository_INFO,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}
	if err := db.DeleteRepository(repo.GetScmRepositoryID()); err != nil {
		t.Fatal(err)
	}
	gotRepos, err := db.GetRepositories(&qf.Repository{ScmRepositoryID: repo.GetScmRepositoryID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotRepos) != 0 {
		t.Fatalf("Expected no repositories, but got: %v", gotRepos)
	}
}

func TestBunGetRepositoriesByOrganization(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:              "Test Course",
		Code:              "DAT100",
		Year:              2017,
		Tag:               "Spring",
		ScmOrganizationID: 1234,
	}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db)

	repoCourseInfo := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   100,
		UserID:            user.GetID(),
		RepoType:          qf.Repository_INFO,
		HTMLURL:           "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	repoAssignment := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   102,
		UserID:            user.GetID(),
		RepoType:          qf.Repository_ASSIGNMENTS,
		HTMLURL:           "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	wantRepo := []*qf.Repository{&repoCourseInfo, &repoAssignment}

	gotRepo, err := db.GetRepositories(&qf.Repository{ScmOrganizationID: 120})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantRepo, gotRepo, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantRepo, +gotRepo):\n%s", diff)
	}
}

func TestBunGetRepoByCourseIdUserIdAndType(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	course := &qf.Course{
		ID:                1234,
		Name:              "Test Course",
		Code:              "DAT100",
		Year:              2017,
		Tag:               "Spring",
		ScmOrganizationID: 120,
	}

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db)
	userTwo := qtest.CreateFakeUser(t, db)

	repoCourseInfo := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   100,
		UserID:            user.GetID(),
		RepoType:          qf.Repository_INFO,
		HTMLURL:           "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	repoAssignment := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   102,
		UserID:            user.GetID(),
		RepoType:          qf.Repository_ASSIGNMENTS,
		HTMLURL:           "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	repoUser := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   103,
		UserID:            user.GetID(),
		RepoType:          qf.Repository_USER,
		HTMLURL:           "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUser); err != nil {
		t.Fatal(err)
	}

	repoUserTwo := qf.Repository{
		ScmOrganizationID: 120,
		ScmRepositoryID:   104,
		UserID:            userTwo.GetID(),
		RepoType:          qf.Repository_USER,
		HTMLURL:           "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUserTwo); err != nil {
		t.Fatal(err)
	}

	wantRepo := []*qf.Repository{&repoUserTwo}

	repoQuery := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		UserID:            userTwo.GetID(),
		RepoType:          qf.Repository_USER,
	}
	gotRepo, err := db.GetRepositories(repoQuery)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantRepo, gotRepo, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantRepo, +gotRepo):\n%s", diff)
	}
}
