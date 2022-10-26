package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

// Deprecated: Will be replaced by MockQuickFeedClient
func testQuickFeedService(t *testing.T) (database.Database, func(), scm.SCM, *web.QuickFeedService) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	sc, mgr := scm.MockSCMManager(t)
	logger := qtest.Logger(t).Desugar()
	return db, cleanup, sc, web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
}

// MockClient returns a QuickFeed client for invoking RPCs.
func MockClient(t *testing.T, db database.Database, opts connect.Option) qfconnect.QuickFeedServiceClient {
	t.Helper()
	_, mgr := scm.MockSCMManager(t)
	logger := qtest.Logger(t)
	qfService := web.NewQuickFeedService(logger.Desugar(), db, mgr, web.BaseHookOptions{}, &ci.Local{})

	if opts == nil {
		opts = connect.WithInterceptors()
	}
	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(qfService, opts))
	server := httptest.NewUnstartedServer(router)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	return qfconnect.NewQuickFeedServiceClient(server.Client(), server.URL)
}

func MockClientWithUser(t *testing.T, db database.Database, user *qf.User) (qfconnect.QuickFeedServiceClient, string) {
	t.Helper()
	_, mgr := scm.MockSCMManager(t)
	logger := qtest.Logger(t)
	qfService := web.NewQuickFeedService(logger.Desugar(), db, mgr, web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	cookie, err := tm.NewAuthCookie(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	opts := connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
	)
	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(qfService, opts))
	server := httptest.NewUnstartedServer(router)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	return qfconnect.NewQuickFeedServiceClient(server.Client(), server.URL), cookie.String()
}
