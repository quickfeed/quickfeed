package web_test

import (
	"os"
	"testing"

	"github.com/autograde/quickfeed/web/auth"
	"github.com/markbates/goth"
)

func TestMain(m *testing.M) {
	// set up fake goth provider (only needs to be done once)
	fakeGothProvider()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func fakeGothProvider() {
	baseURL := "fake"
	goth.UseProviders(&auth.FakeProvider{
		Callback: auth.GetCallbackURL(baseURL, "fake"),
	})
	goth.UseProviders(&auth.FakeProvider{
		Callback: auth.GetCallbackURL(baseURL, "fake-teacher"),
	})
}
