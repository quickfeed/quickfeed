package scm

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

func GetTestOrganization(t *testing.T) string {
	t.Helper()
	qfTestOrg := os.Getenv("QF_TEST_ORG")
	if len(qfTestOrg) < 1 {
		t.Skip("This test requires that the 'QF_TEST_ORG' is set and that you have access to said GitHub organization")
	}
	return qfTestOrg
}

func GetTestSCM(t *testing.T) (*GithubSCM, string) {
	t.Helper()
	accessToken := GetAccessToken(t)
	scmClient := NewGithubUserClient(qtest.Logger(t), accessToken)
	user, _, err := scmClient.client.Users.Get(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	return scmClient, *user.Login
}

func GetAccessToken(t *testing.T) string {
	t.Helper()
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires that 'GITHUB_ACCESS_TOKEN' is set.")
	}
	return accessToken
}

const (
	envFile               = ".env-testing"
	appName               = "QuickFeed Testing App"
	appCreateInstructions = `The %s is not configured in %s.
To create the %s and configure it, run:
%% cd web/manifest ; GITHUB_APP=1 go test -v -run TestCreateQuickFeedApp`
	InstallInstructions = `You may need to manually install the %s on the %s organization.

From the user that installed the %s:
Select Settings -> Developer settings -> GitHub Apps ->
Select Edit: QuickFeed Testing ->
Select Install App (left menu)
Find your test organization in the list and click Install
`
)

var (
	mgr  *Manager
	once sync.Once
)

func GetSCMManager(t *testing.T) *Manager {
	if mgr != nil {
		return mgr
	}
	once.Do(func() {
		t.Helper()
		if os.Getenv("GITHUB_APP") == "" {
			t.Skipf("Skipping test. To run: GITHUB_APP=1 go test -v -run %s", t.Name())
		}
		// Load environment variables from $QUICKFEED/.env-testing.
		// Will not override variables already defined in the environment.
		if err := env.Load(env.RootEnv(envFile)); err != nil {
			t.Fatal(err)
		}
		if !env.HasAppID() {
			t.Fatalf(appCreateInstructions, appName, envFile, appName)
		}
		var err error
		mgr, err = NewSCMManager()
		if err != nil {
			t.Fatal(err)
		}
	})
	return mgr
}

func GetAppSCM(t *testing.T) SCM {
	t.Helper()
	if os.Getenv("GITHUB_APP") == "" {
		t.Skipf("Skipping test. To run: GITHUB_APP=1 go test -v -run %s", t.Name())
	}
	if mgr == nil {
		GetSCMManager(t)
	}
	qfTestOrg := GetTestOrganization(t)
	appSCM, err := mgr.GetOrCreateSCM(context.Background(), qtest.Logger(t), qfTestOrg)
	if err != nil {
		t.Logf(InstallInstructions, appName, qfTestOrg, appName)
		t.Fatal(err)
	}
	return appSCM
}
