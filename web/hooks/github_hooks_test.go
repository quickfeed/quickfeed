package hooks

import (
	"log"
	"net/http"
	"testing"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// serverURL       = "https://cyclone.itest.run/hook/github/events"
	// serverURL       = "https://376fe6aa.ngrok.io/webhook"
	// gitHubTestOrg   = "autograder-test"
	// gitHubTestOrgID = 30462712
	secret = "the-secret-autograder-test"
)

// To enable this test, please see instructions in the developer guide (dev.md).
// You will also need access to the autograder-test organization; you may request
// access by sending your GitHub username to hein.meling at uis.no.

// On macOS, get ngrok using `brew cask install ngrok`.
// See steps to follow [here](https://groob.io/tutorial/go-github-webhook/).

func TestGitHubWebHook(t *testing.T) {
	t.Skip("Disabled; see comment in code")

	//TODO(meling) these lines are not necessary unless we use the scm.ListHook() and CreateHook() commands
	// accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	// if len(accessToken) < 1 {
	// 	t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' and access to the 'autograder-test' GitHub organization")
	// }
	//TODO(meling) Use scm.ListHook() to check if hooks installed, and scm.CreateHook() to create, if not.

	// zap.NewNop().Sugar()

	cfg := zap.NewDevelopmentConfig()
	// database logging is only enabled if the LOGDB environment variable is set
	// cfg = database.GormLoggerConfig(cfg)
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		log.Fatalf("can't initialize logger: %v\n", err)
	}
	defer logger.Sync()

	var db database.Database
	var runner ci.Runner
	webhook := NewGitHubWebHook(logger.Sugar(), db, runner, secret)

	log.Println("starting webhook server")
	http.HandleFunc("/webhook", webhook.Handle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
