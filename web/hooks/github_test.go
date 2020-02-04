package hooks

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	gitHubTestOrg   = "autograder-test"
	gitHubTestOrgID = 30462712
	secret          = "the-secret-autograder-test"
)

// To enable this test, please see instructions in the developer guide (dev.md).
// You will also need access to the autograder-test organization; you may request
// access by sending your GitHub username to hein.meling at uis.no.

// On macOS, get ngrok using `brew cask install ngrok`.
// See steps to follow [here](https://groob.io/tutorial/go-github-webhook/).

// To run this test, use the following (replace the forwarding URL with your own):
//
// NGROK_FWD=https://53c51fa9.ngrok.io go test -v -run TestGitHubWebHook
//
// This will block waiting for a push event from GitHub; meaning that you
// will manually have to create a push event to the 'tests' repository.

func TestGitHubWebHook(t *testing.T) {
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' and access to the 'autograder-test' GitHub organization")
	}
	serverURL := os.Getenv("NGROK_FWD")
	if len(serverURL) < 1 {
		t.Skip("This test requires a 'NGROK_FWD' and access to the 'autograder-test' GitHub organization")
	}

	logger := getLogger(true)
	defer logger.Sync()

	var s scm.SCM
	s, err := scm.NewSCMClient(logger, "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	opt := &scm.CreateHookOptions{
		URL:        serverURL,
		Secret:     secret,
		Repository: &scm.Repository{Owner: gitHubTestOrg, Path: "tests"},
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

	var db database.Database
	var runner ci.Runner
	webhook := NewGitHubWebHook(logger, db, runner, secret)

	log.Println("starting webhook server")
	http.HandleFunc("/webhook", webhook.Handle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getLogger(verbose bool) *zap.SugaredLogger {
	if verbose {
		cfg := zap.NewDevelopmentConfig()
		// database logging is only enabled if the LOGDB environment variable is set
		cfg = database.GormLoggerConfig(cfg)
		// add colorization
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// we only want stack trace enabled for panic level and above
		logger, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
		if err != nil {
			log.Fatalf("can't initialize logger: %v\n", err)
		}
		return logger.Sugar()
	}
	return zap.NewNop().Sugar()
}
