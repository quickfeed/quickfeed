package database_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"gorm.io/gorm"
)

func TestGormDBGetEmptyRepo(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	repos, err := db.GetRepositories(&pb.Repository{RepositoryID: 10})
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
	repo := pb.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		UserID:         user.ID,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetRepositories(&pb.Repository{RepositoryID: repo.RepositoryID}); err != nil {
		t.Fatal(err)
	}
}

func TestGormDBCreateSingleRepoWithMissingUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	repo := pb.Repository{
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

	repo := pb.Repository{
		OrganizationID: 120,
		RepositoryID:   100,
		RepoType:       pb.Repository_COURSEINFO,
	}
	if err := db.CreateRepository(&repo); err != nil {
		t.Fatal(err)
	}

	gotRepo, err := db.GetRepositories(&pb.Repository{RepositoryID: repo.RepositoryID})
	if err != nil {
		t.Fatal(err)
	}
	if !gotRepo[0].RepoType.IsCourseRepo() {
		t.Fatalf("Expected course info repo (%v), but got: %v", pb.Repository_COURSEINFO, gotRepo[0].RepoType)
	}
}

