package web

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// testQuickFeedService is a clone of the same function in quickfeed_helper_test.go.
// It is replicated here to avoid import cycle.
func testQuickFeedService(t *testing.T) (database.Database, func(), scm.SCM, *QuickFeedService) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	sc, mgr := scm.MockSCMManager(t)
	logger := qlog.Logger(t).Desugar()
	return db, cleanup, sc, NewQuickFeedService(logger, db, mgr, BaseHookOptions{}, &ci.Local{})
}

// StartGrpcAuthServer will set up mux server with interceptors passed in opts. If opts argument is nil,
// the server will start with full interceptor chain.
func StartGrpcAuthServer(t *testing.T, qfService *QuickFeedService, tm *auth.TokenManager, opts connect.Option) (func(), func(context.Context)) {
	t.Helper()

	router := http.NewServeMux()
	if opts == nil {
		router.Handle(qfService.NewQuickFeedHandler(tm))
	} else {
		router.Handle(qfconnect.NewQuickFeedServiceHandler(qfService, opts))
	}

	muxServer := &http.Server{
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		Addr:              "127.0.0.1:8081",
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}

	return func() {
			if err := muxServer.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					t.Errorf("Server exited with unexpected error: %v", err)
				}
				return
			}
		}, func(ctx context.Context) {
			if err := muxServer.Shutdown(ctx); err != nil {
				t.Fatal(err)
			}
		}
}
