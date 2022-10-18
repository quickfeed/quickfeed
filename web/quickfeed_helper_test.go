package web_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

// Deprecated: Will be replaced by MockQuickFeedClient
func testQuickFeedService(t *testing.T) (database.Database, func(), scm.SCM, *web.QuickFeedService) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	sc, mgr := scm.MockSCMManager(t)
	logger := qtest.Logger(t).Desugar()
	return db, cleanup, sc, web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
}

// MockQuickFeedClient returns a QuickFeed client for invoking RPCs.
func MockQuickFeedClient(t *testing.T, db database.Database, opts connect.Option) (func(context.Context), qfconnect.QuickFeedServiceClient) {
	t.Helper()
	logger := qtest.Logger(t)

	if opts == nil {
		opts = connect.WithInterceptors()
	}
	shutdown := web.MockQuickFeedServer(t, logger, db, opts)

	const serverURL = "http://127.0.0.1:8081"
	return func(ctx context.Context) {
		shutdown(ctx)
	}, qfconnect.NewQuickFeedServiceClient(http.DefaultClient, serverURL)
}
