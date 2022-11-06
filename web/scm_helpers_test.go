package web_test

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestInitSCMs(t *testing.T) {
	scmConfig, err := scm.NewSCMConfig()
	if err != nil {
		t.Skip("Requires a valid SCM app")
	}
	mgr := scm.NewSCMManager(scmConfig)
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	ctx := context.Background()
	logger := qtest.Logger(t).Desugar()
	q := web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
	admin := qtest.CreateFakeUser(t, db, 1)
	course := &qf.Course{
		Name:             "Test course",
		OrganizationName: scm.GetTestOrganization(t),
	}
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Error(err)
	}
	if err := q.InitSCMs(ctx); err != nil {
		t.Error(err)
	}
	if _, ok := mgr.GetSCM(course.OrganizationName); !ok {
		t.Errorf("Missing scm client for organization %s", course.OrganizationName)
	}
}
