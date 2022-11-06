package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGormDBGetEmptyRepo(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	repos, err := db.GetRepositories(&qf.Repository{RepositoryID: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 0 {
		t.Fatalf("Expected no repositories, but got: %v", repos)
	}
}

func TestGormDBGetSingleRepoWithUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user := qtest.CreateFakeUser(t, db, 10)
	repo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		UserID:         user.ID,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetRepositories(&qf.Repository{RepositoryID: repo.RepositoryID}); err != nil {
		t.Fatal(err)
	}
}

func TestGormDBCreateSingleRepoWithMissingUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	repo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		UserID:         20,
	}
	if err := db.CreateRepository(&repo); err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}
}

func TestGormDBGetCourseRepoType(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	repo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		RepoType:       qf.Repository_INFO,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	gotRepos, err := db.GetRepositories(&qf.Repository{RepositoryID: repo.RepositoryID})
	if err != nil {
		t.Fatal(err)
	}
	if !gotRepos[0].RepoType.IsCourseRepo() {
		t.Fatalf("Expected course info repo (%v), but got: %v", qf.Repository_INFO, gotRepos[0].RepoType)
	}
}

func TestGormDeleteRepo(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	repo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		RepoType:       qf.Repository_INFO,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}
	if err := db.DeleteRepository(repo.RepositoryID); err != nil {
		t.Fatal(err)
	}
	gotRepos, err := db.GetRepositories(&qf.Repository{RepositoryID: repo.RepositoryID})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotRepos) != 0 {
		t.Fatalf("Expected no repositories, but got: %v", gotRepos)
	}
}

func TestGetRepositoriesByOrganization(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 1234,
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		UserID:         user.ID,
		RepoType:       qf.Repository_INFO,
		HTMLURL:        "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   102,
		UserID:         user.ID,
		RepoType:       qf.Repository_ASSIGNMENTS,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	wantRepo := []*qf.Repository{&repoCourseInfo, &repoAssignment}

	gotRepo, err := db.GetRepositories(&qf.Repository{OrganizationID: 120})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantRepo, gotRepo, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantRepo, +gotRepo):\n%s", diff)
	}
}

func TestGetRepoByCourseIdUserIdAndType(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		ID:             1234,
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 120,
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 10)
	userTwo := qtest.CreateFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		UserID:         user.ID,
		RepoType:       qf.Repository_INFO,
		HTMLURL:        "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   102,
		UserID:         user.ID,
		RepoType:       qf.Repository_ASSIGNMENTS,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for user
	repoUser := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   103,
		UserID:         user.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUser); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for userTwo
	repoUserTwo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   104,
		UserID:         userTwo.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUserTwo); err != nil {
		t.Fatal(err)
	}

	wantRepo := []*qf.Repository{&repoUserTwo}

	repoQuery := &qf.Repository{
		OrganizationID: course.OrganizationID,
		UserID:         userTwo.ID,
		RepoType:       qf.Repository_USER,
	}
	gotRepo, err := db.GetRepositories(repoQuery)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantRepo, gotRepo, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantRepo, +gotRepo):\n%s", diff)
	}
}

func TestGetRepositoryByCourseUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 120,
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 10)
	userTwo := qtest.CreateFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		UserID:         user.ID,
		RepoType:       qf.Repository_INFO,
		HTMLURL:        "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   102,
		UserID:         user.ID,
		RepoType:       qf.Repository_ASSIGNMENTS,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for user
	repoUser := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   103,
		UserID:         user.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUser); err != nil {
		t.Fatal(err)
	}

	// Creating UserRepo for userTwo
	repoUserTwo := qf.Repository{
		OrganizationID: 120,
		RepositoryID:   104,
		UserID:         userTwo.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoUserTwo); err != nil {
		t.Fatal(err)
	}

	wantRepo := []*qf.Repository{&repoUserTwo}

	repoQuery := &qf.Repository{
		OrganizationID: course.OrganizationID,
		UserID:         userTwo.ID,
		RepoType:       qf.Repository_USER,
	}
	gotRepo, err := db.GetRepositories(repoQuery)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantRepo, gotRepo, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantRepo, +gotRepo):\n%s", diff)
	}
}

func TestGetRepositoriesByCourseIdAndType(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:           "Test Course",
		Code:           "DAT100",
		Year:           2017,
		Tag:            "Spring",
		OrganizationID: 1234,
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	user := qtest.CreateFakeUser(t, db, 11)

	// Creating Course info repo
	repoCourseInfo := qf.Repository{
		OrganizationID: 1234,
		RepositoryID:   100,
		UserID:         user.ID,
		RepoType:       qf.Repository_INFO,
		HTMLURL:        "http://repoCourseInfo.com/",
	}
	if err := db.CreateRepository(&repoCourseInfo); err != nil {
		t.Fatal(err)
	}

	// Creating AssignmentRepo
	repoAssignment := qf.Repository{
		OrganizationID: 1234,
		RepositoryID:   102,
		UserID:         user.ID,
		RepoType:       qf.Repository_ASSIGNMENTS,
		HTMLURL:        "http://repoAssignment.com/",
	}
	if err := db.CreateRepository(&repoAssignment); err != nil {
		t.Fatal(err)
	}

	wantRepo := []*qf.Repository{&repoCourseInfo}

	repoQuery := &qf.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       qf.Repository_INFO,
	}
	gotRepo, err := db.GetRepositories(repoQuery)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantRepo, gotRepo, protocmp.Transform()); diff != "" {
		t.Errorf("GetRepositories() mismatch (-wantRepo, +gotRepo):\n%s", diff)
	}
}
