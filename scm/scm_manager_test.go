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
		t.Skip("Requires a valid SCM app")
	}
	mgr := scm.NewSCMManager(scmConfig)
	qfTestOrg := scm.GetTestOrganization(t)
	ctx := context.Background()
	_, err = mgr.GetOrCreateSCM(ctx, qtest.Logger(t), qfTestOrg)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := mgr.Scms.GetSCM(qfTestOrg)
	if !ok {
		t.Errorf("Scm client for organization %s not found", qfTestOrg)
	}
}
