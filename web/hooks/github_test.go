package hooks

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	logq "github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
)

const (
	secret = "the-secret-quickfeed-test"
)

// To run these tests, please see instructions in the developer guide (dev.md).
// On macOS, get ngrok using `brew install ngrok`.
// See steps to follow [here](https://groob.io/tutorial/go-github-webhook/).

// To run these tests, use the following (replace the forwarding URL with your own):
//
// QF_WEBHOOK_SERVER=https://53c51fa9.ngrok.io go test -v -run <test name> -timeout 999999s
// This will create a new webhook with URL `https://53c51fa9.ngrok.io/webhook`
// The -timeout flag is not necessary, but stops the test from timing out after 10 minutes.
//
// These tests will then block waiting for an event from GitHub; meaning that you
// will manually have to create these events.

// TODO(meling) add code to create a push event to the tests repository and trigger synchronizing tasks with issues on repositories.
// TestGitHubWebHook tests listening to events from the tests repository.
func TestGitHubWebHook(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	serverURL := scm.GetWebHookServer(t)
	s := scm.GetTestSCM(t)

	logger := logq.Logger(t)
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()
	opt := &scm.CreateHookOptions{
		URL:        serverURL + "/webhook",
		Secret:     secret,
		Repository: &scm.Repository{Owner: qfTestOrg, Path: "tests"},
	}
	err := s.CreateHook(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	hooks, err := s.ListHooks(ctx, opt.Repository, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}

	var db database.Database
	var runner ci.Runner
	webhook := NewGitHubWebHook(logger, db, runner, secret)

	log.Println("starting webhook server")
	http.HandleFunc("/webhook", webhook.Handle())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
