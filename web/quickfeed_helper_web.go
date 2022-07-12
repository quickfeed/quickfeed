package web

import (
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
)

// testQuickFeedService is a clone of the same function in quickfeed_helper_test.go.
// It is replicated here to avoid import cycle.
func testQuickFeedService(t *testing.T) (database.Database, func(), scm.SCM, *QuickFeedService) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	scm, scms := qtest.FakeProviderMap(t)
	logger := qlog.Logger(t).Desugar()
	return db, cleanup, scm, NewQuickFeedService(logger, db, scms, BaseHookOptions{}, &ci.Local{})
}
