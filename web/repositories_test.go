package web

import (
	"errors"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGetRepo(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user := qtest.CreateFakeUser(t, db, 1)
	course := &pb.Course{
		OrganizationID: 1,
		Code:           "DAT101",
	}
	qtest.CreateCourse(t, db, user, course)
	group := &pb.Group{
		Name:     "1001 Hacking Crew",
		CourseID: course.ID,
		Users:    []*pb.User{user},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	wantUserRepo := &pb.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         user.ID,
		RepoType:       pb.Repository_USER,
		HTMLURL:        "http://assignment.com/",
	}
	if err := db.CreateRepository(wantUserRepo); err != nil {
		t.Fatal(err)
	}

	wantGroupRepo := &pb.Repository{
		OrganizationID: 1,
		RepositoryID:   2,
		GroupID:        group.ID,
		RepoType:       pb.Repository_GROUP,
		HTMLURL:        "http://assignment.com/",
	}
	if err := db.CreateRepository(wantGroupRepo); err != nil {
		t.Fatal(err)
	}

	_, scms := qtest.FakeProviderMap(t)
	ags := NewAutograderService(zap.NewNop(), db, scms, BaseHookOptions{}, &ci.Local{})
	gotUserRepo, err := ags.getRepo(course, user.ID, pb.Repository_USER)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantUserRepo, gotUserRepo, protocmp.Transform()); diff != "" {
		t.Errorf("getRepo() mismatch (-wantUserRepo, +gotUserRepo):\n%s", diff)
	}

	gotGroupRepo, err := ags.getRepo(course, group.ID, pb.Repository_GROUP)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantGroupRepo, gotGroupRepo, protocmp.Transform()); diff != "" {
		t.Errorf("getRepo() mismatch (-wantGroupRepo, +gotGroupRepo):\n%s", diff)
	}

	_, err = ags.getRepo(course, group.ID+1, pb.Repository_GROUP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
	_, err = ags.getRepo(course, user.ID+1, pb.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
}
