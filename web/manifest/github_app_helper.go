package manifest

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/web"
)

func CreateQuickFeedApp(t *testing.T) {
	t.Helper()
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
	if err := ReadyForAppCreation(envFile); err != nil {
		t.Fatal(err)
	}
	if err := CreateNewQuickFeedApp(web.NewDevelopmentServer, ":443", envFile); err != nil {
		t.Fatal(err)
	}
}
