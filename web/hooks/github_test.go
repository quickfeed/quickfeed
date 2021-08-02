package hooks

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	logq "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
)

const (
	secret = "the-secret-quickfeed-test"
)

// To run this test, please see instructions in the developer guide (dev.md).

// On macOS, get ngrok using `brew install ngrok`.
// See steps to follow [here](https://groob.io/tutorial/go-github-webhook/).

// To run this test, use the following (replace the forwarding URL with your own):
//
// QF_WEBHOOK_SERVER=https://53c51fa9.ngrok.io go test -v -run TestGitHubWebHook
//
// This will create a new webhook with URL `https://53c51fa9.ngrok.io/webhook`
// for the $QF_TEST_ORG/tests repository for handling push events.
//
// This test will then block waiting for a push event from GitHub; meaning that you
// will manually have to create a push event to the 'tests' repository.
//
// TODO(meling) add code to create a push event to the tests repository.

func TestGitHubWebHook(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	serverURL := scm.GetWebHookServer(t)

	logger := logq.Zap(true).Sugar()
	defer func() { _ = logger.Sync() }()

	s, err := scm.NewSCMClient(logger, "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	opt := &scm.CreateHookOptions{
		URL:        serverURL + "/webhook",
		Secret:     secret,
		Repository: &scm.Repository{Owner: qfTestOrg, Path: "tests"},
	}
	err = s.CreateHook(ctx, opt)
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

	// TODO(meling) db is nil; will cause handling of push event to panic; will need a database with content for this to work fully.
	var db database.Database
	var runner ci.Runner
	webhook := NewGitHubWebHook(logger, db, runner, secret)

	log.Println("starting webhook server")
	http.HandleFunc("/webhook", webhook.Handle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func TestExtractChanges(t *testing.T) {
	modifiedFiles := []string{
		"go.mod",
		"go.sum",
		"exercise.go",
		"README.md",
		"lab2/fib.go",
		"lab3/detector/fd.go",
		"paxos/proposer.go",
		"/hallo",
		"",
	}
	want := map[string]bool{
		"lab2":  true,
		"lab3":  true,
		"paxos": true,
	}
	got := make(map[string]bool)
	extractChanges(modifiedFiles, got)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("content mismatch (-want +got):\n%s", diff)
	}
}
