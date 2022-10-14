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
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// testQuickFeedService is a clone of the same function in quickfeed_helper_test.go.
// It is replicated here to avoid import cycle.
func testQuickFeedService(t *testing.T) (database.Database, func(), scm.SCM, *QuickFeedService) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	sc, mgr := scm.MockSCMManager(t)
	logger := qtest.Logger(t).Desugar()
	return db, cleanup, sc, NewQuickFeedService(logger, db, mgr, BaseHookOptions{}, &ci.Local{})
}

// MockQuickFeedServer is a test helper that starts a QuickFeed server in a goroutine, with the given options, typically interceptors.
// The returned function must be called to shut down the server and the end of a test.
func MockQuickFeedServer(t *testing.T, logger *zap.SugaredLogger, db database.Database, opts connect.Option) func(context.Context) {
	t.Helper()
	_, mgr := scm.MockSCMManager(t)
	qfService := NewQuickFeedService(logger.Desugar(), db, mgr, BaseHookOptions{}, &ci.Local{})

	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(qfService, opts))
	muxServer := &http.Server{
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		Addr:              "127.0.0.1:8081",
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}

	go func() {
		if err := muxServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				t.Errorf("Server exited with unexpected error: %v", err)
			}
			return
		}
	}()
	return func(ctx context.Context) {
		if err := muxServer.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
	}
}
