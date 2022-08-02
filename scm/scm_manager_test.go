package scm_test

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/scm"
)

func TestSCMManager(t *testing.T) {
	appID, err := env.AppID()
	if err != nil {
		t.Fatal(err)
	}
	appKey, err := env.AppKey()
	if err != nil {
		t.Fatal(err)
	}

	tm, err := scm.NewSCMManager(appID, appKey)
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
