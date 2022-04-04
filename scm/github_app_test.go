package scm_test

import (
	"context"
	"testing"

	"github.com/autograde/quickfeed/scm"
)

func TestGitHubApp(t *testing.T) {
	app, err := scm.NewApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	qfTestOrg := scm.GetTestOrganization(t)

	client, err := app.NewInstallationClient(ctx, qfTestOrg)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = client.Organizations.Get(ctx, qfTestOrg); err != nil {
		t.Fatal(err)
	}
}
