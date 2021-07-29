package scm_test

import (
	"context"
	"os"
	"testing"

	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

const (
	qf101Org   = "qf101"
	qf101OrdID = 77283363
	serverURL  = "https://53c51fa9.ngrok.io" // TODO(meling) Should make this a persistent URL for receiving hook events
	secret     = "the-secret-quickfeed-test"
)

// To run this test, please see instructions in the developer guide (dev.md).

// These tests only test listing existing hooks and creating a new one.
// They do not cover processing push events on a server.
// See web/hooks package for tests involving processing push events.

func TestGetOrganization(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	org, err := s.GetOrganization(context.Background(), &scm.GetOrgOptions{Name: qfTestOrg})
	if err != nil {
		t.Fatal(err)
	}
	if qfTestOrg == qf101Org {
		if org.ID != qf101OrdID {
			t.Errorf("scm.GetOrganization('%s') = %d, expected %d", qfTestOrg, org.ID, qf101OrdID)
		}
	} else {
		// Otherwise we just print the organization result
		t.Logf("org: %v", org)
	}
}

func TestListHooks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	hooks, err := s.ListHooks(ctx, nil, qfTestOrg)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how which hooks might be registered
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}

	hooks, err = s.ListHooks(ctx, &scm.Repository{Owner: qfTestOrg, Path: "tests"}, "")
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how which hooks might be registered
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}

	hooks, err = s.ListHooks(ctx, &scm.Repository{Path: "tests"}, "")
	if err == nil {
		t.Fatal("expected error 'ListHooks: called with missing or incompatible arguments: ...'")
	}
	// We don't actually test anything here since we don't know how which hooks might be registered
	t.Logf("%v %v", hooks, err)
}

func TestCreateHook(t *testing.T) {
	// Only enable this test to add a new webhook to your test course organization
	if os.Getenv("QF_WEBHOOK") == "" {
		t.Skip("Disabled pending support for deleting webhooks")
	}
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	opt := &scm.CreateHookOptions{
		URL:        serverURL,
		Secret:     secret,
		Repository: &scm.Repository{Owner: qfTestOrg, Path: "tests"},
	}
	err = s.CreateHook(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	hooks, err := s.ListHooks(ctx, opt.Repository, "")
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how which hooks might be registered
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}
}
