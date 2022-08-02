package web_test

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestMakeSCMs(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	ctx := context.Background()
	logger := qlog.Logger(t).Desugar()
	id, err := env.AppID()
	if err != nil {
		t.Skip("Requires application ID")
	}
	key, err := env.AppKey()
	if err != nil {
		t.Skip("Requires application key")
	}
	s, err := scm.NewSCMManager(id, key)
	if err != nil {
		t.Fatal(err)
	}
	q := web.NewQuickFeedService(logger, db, s, web.BaseHookOptions{}, &ci.Local{})
	admin := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{
		Name:             "Test course",
		OrganizationPath: scm.GetTestOrganization(t),
		Provider:         "fake",
	}
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}
	if err := q.MakeSCMs(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := q.GetSCM(ctx, course.OrganizationPath); err != nil {
		t.Fatal(err)
	}
}
