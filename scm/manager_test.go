package scm_test

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/scm"
)

func TestSCMManager(t *testing.T) {
	const appName = "QuickFeed Testing App"
	mgr := scm.GetSCMManager(t)
	qfTestOrg := scm.GetTestOrganization(t)
	ctx := context.Background()
	createdSCM, err := mgr.GetOrCreateSCM(ctx, qtest.Logger(t), qfTestOrg)
	if err != nil {
		t.Logf(scm.InstallInstructions, appName, qfTestOrg, appName)
		t.Fatal(err)
	}
	gotSCM, ok := mgr.GetSCM(qfTestOrg)
	if !ok {
		t.Errorf("Scm client for organization %s not found", qfTestOrg)
	}
	if gotSCM != createdSCM {
		t.Errorf("Expected scm client %v, got %v", createdSCM, gotSCM)
	}
}
