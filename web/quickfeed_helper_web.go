package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

// MockClient returns a QuickFeed client for invoking RPCs.
func MockClient(t *testing.T, db database.Database, opts connect.Option) qfconnect.QuickFeedServiceClient {
	t.Helper()
	mgr := scm.MockManager(t, scm.WithMockOrgs())
	logger := qtest.Logger(t)
	qfService := NewQuickFeedService(logger.Desugar(), db, mgr, BaseHookOptions{}, &ci.Local{})

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

func MockClientWithOption(t *testing.T, db database.Database, mockOpt scm.MockOption, clientOpts ...connect.ClientOption) (qfconnect.QuickFeedServiceClient, *auth.TokenManager) {
	t.Helper()
	mgr := scm.MockManager(t, mockOpt)
	logger := qtest.Logger(t)
	qfService := NewQuickFeedService(logger.Desugar(), db, mgr, BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	opts := connect.WithInterceptors(
		interceptor.NewValidationInterceptor(logger),
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
