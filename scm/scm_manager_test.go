package scm_test

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/scm"
)

func TestSCMManager(t *testing.T) {
	scmConfig, err := scm.NewSCMConfig()
	if err != nil {
		t.Skip("This test needs a valid GitHub App")
	}

	tm, err := scm.NewSCMManager(scmConfig)
	if err != nil {
		t.Fatal(err)
	}
	qfTestOrg := scm.GetTestOrganization(t)
	ctx := context.Background()
	_, err = tm.GetOrCreateSCM(ctx, qtest.Logger(t), qfTestOrg)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := tm.Scms.GetSCM(qfTestOrg)
	if !ok {
		t.Errorf("Scm client for organization %s not found", qfTestOrg)
	}
}
