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

func MockClientWithUser(t *testing.T, db database.Database, clientOpts ...connect.ClientOption) (qfconnect.QuickFeedServiceClient, *auth.TokenManager, scm.SCM) {
	t.Helper()
	scmClient, mgr := scm.MockSCMManager(t)
	logger := qtest.Logger(t)
	qfService := web.NewQuickFeedService(logger.Desugar(), db, mgr, web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

	opts := connect.WithInterceptors(
		interceptor.NewTokenAuthInterceptor(logger, tm, db),
		interceptor.NewUserInterceptor(logger, tm),
		interceptor.NewAccessControlInterceptor(tm),
	)
	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(qfService, opts))
	server := httptest.NewUnstartedServer(router)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	return qfconnect.NewQuickFeedServiceClient(server.Client(), server.URL, clientOpts...), tm, scmClient
}

func MockClientWithUserAndCourse(t *testing.T, db database.Database, clientOpts ...connect.ClientOption) (qfconnect.QuickFeedServiceClient, *auth.TokenManager) {
	t.Helper()
	_, mgr := scm.MockSCMManagerWithCourse(t)
	logger := qtest.Logger(t)
	qfService := web.NewQuickFeedService(logger.Desugar(), db, mgr, web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

	opts := connect.WithInterceptors(
		interceptor.NewTokenAuthInterceptor(logger, tm, db),
		interceptor.NewUserInterceptor(logger, tm),
		interceptor.NewAccessControlInterceptor(tm),
	)
	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(qfService, opts))
	server := httptest.NewUnstartedServer(router)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	return qfconnect.NewQuickFeedServiceClient(server.Client(), server.URL, clientOpts...), tm
}

func Cookie(t *testing.T, tm *auth.TokenManager, user *qf.User) string {
	t.Helper()
	cookie, err := tm.NewAuthCookie(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	return cookie.String()
}
