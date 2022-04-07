package scm_test

import (
	"context"
	"testing"

	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
)

func TestGitHubApp(t *testing.T) {
	testOrg := scm.GetTestOrganization(t)
	client := scm.GetTestClient(t, testOrg)
	ctx := context.Background()
	if _, _, err := client.Organizations.Get(ctx, testOrg); err != nil {
		t.Fatal(err)
	}
	sc := scm.NewGithubSCMClient(log.Zap(false).Sugar(), client, scm.GetAccessToken(t))
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{Name: testOrg})
	if err != nil {
		t.Fatal(err)
	}
	_, err = sc.GetRepositories(ctx, org)
	if err != nil {
		t.Fatal(err)
	}
}
