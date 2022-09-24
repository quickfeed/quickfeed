package web_test

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

// Deprecated: Will be replaced by MockQuickFeedClient
func testQuickFeedService(t *testing.T) (database.Database, func(), scm.SCM, *web.QuickFeedService) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	sc, mgr := scm.MockSCMManager(t)
	logger := qlog.Logger(t).Desugar()
	return db, cleanup, sc, web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
}

// MockQuickFeedClient returns a QuickFeed client for invoking RPCs.
// Currently no interceptors are passed.
// TODO(meling): Consider passing interceptors as input to this function to accommodate different test scenarios.
func MockQuickFeedClient(t *testing.T) (database.Database, func(context.Context), qfconnect.QuickFeedServiceClient) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	logger := qtest.Logger(t)

	shutdown := web.MockQuickFeedServer(t, logger, db, connect.WithInterceptors(
	// interceptor.Validation(logger),
	))

	return db, func(ctx context.Context) {
		cleanup()
		shutdown(ctx)
	}, qtest.QuickFeedClient("")
}
