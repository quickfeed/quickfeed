package scm

import (
	"context"
	"os"
	"testing"

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

func GetTestSCM(t *testing.T) (SCM, string) {
	t.Helper()
	accessToken := GetAccessToken(t)
	scmClient := NewGithubSCMClient(qtest.Logger(t), accessToken)
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
