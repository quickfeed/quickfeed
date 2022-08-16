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
	sc, mgr := qtest.FakeProviderMap(t)
	logger := qlog.Logger(t).Desugar()
	return db, cleanup, sc, web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
}
