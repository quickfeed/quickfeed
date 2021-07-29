package scm

import (
	"os"
	"testing"
)

func GetTestOrganization(t *testing.T) string {
	t.Helper()
	qfTestOrg := os.Getenv("QF_TEST_ORG")
	if len(qfTestOrg) < 1 {
		t.Skip("This test requires that the 'QF_TEST_ORG' is set and that you have access to said GitHub organization")
	}
	return qfTestOrg
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
