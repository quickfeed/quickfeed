package scm

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/qlog"
)

func GetTestOrganization(t *testing.T) string {
	t.Helper()
	qfTestOrg := os.Getenv("QF_TEST_ORG")
	if len(qfTestOrg) < 1 {
		t.Skip("This test requires that the 'QF_TEST_ORG' is set and that you have access to said GitHub organization")
	}
	return qfTestOrg
}

func GetTestUser(t *testing.T) string {
	t.Helper()
	qfTestUser := os.Getenv("QF_TEST_USER")
	if len(qfTestUser) < 1 {
		t.Skip("This test requires that the 'QF_TEST_USER' is set and that the corresponding user repository exists in the GitHub organization")
	}
	return qfTestUser
}

func GetAccessToken(t *testing.T) string {
	t.Helper()
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		qfTestOrg := GetTestOrganization(t)
		t.Skipf("This test requires that 'GITHUB_ACCESS_TOKEN' is set and that you have access to the '%v' GitHub organization", qfTestOrg)
	}
	return accessToken
}

func GetWebHookServer(t *testing.T) string {
	t.Helper()
	serverURL := os.Getenv("QF_WEBHOOK_SERVER")
	if len(serverURL) < 1 {
		qfTestOrg := GetTestOrganization(t)
		t.Skipf("This test requires that 'QF_WEBHOOK_SERVER' is set and that you have access to the '%v' GitHub organization", qfTestOrg)
	}
	return serverURL
}

func GetTestSCM(t *testing.T) SCM {
	t.Helper()
	accessToken := GetAccessToken(t)
	s, err := NewSCMClient(qlog.Logger(t), accessToken)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
