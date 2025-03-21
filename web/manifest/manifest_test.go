package manifest_test

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/manifest"
)

func TestCreateQuickFeedApp(t *testing.T) {
	if os.Getenv("GITHUB_APP") == "" {
		t.Skipf("Skipping test. To run: GITHUB_APP=1 go test -v -run %s", t.Name())
	}
	// Load environment variables from $QUICKFEED/.env-testing.
	// Will not override variables already defined in the environment.
	const envFile = ".env-testing"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		t.Fatal(err)
	}
	if env.HasAppID() {
		return // App already created and configured.
	}
	if env.Domain() == "localhost" {
		t.Fatal(`Domain "localhost" is unsupported; use "127.0.0.1" instead.`)
	}
	if err := manifest.ReadyForAppCreation(envFile); err != nil {
		t.Fatal(err)
	}
	if err := manifest.CreateNewQuickFeedApp(web.NewDevelopmentServer, ":443", envFile, false); err != nil {
		t.Fatal(err)
	}
}
