package web_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func testQuickFeedService(t *testing.T) (database.Database, func(), scm.SCM, *web.QuickFeedService) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	scm, scms := qtest.FakeProviderMap(t)
	logger := qlog.Logger(t).Desugar()
	return db, cleanup, scm, web.NewQuickFeedService(logger, db, scms, web.BaseHookOptions{}, &ci.Local{})
}
