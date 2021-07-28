package scm_test

import (
	"context"
	"os"
	"testing"

	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

const (
	serverURL     = "https://53c51fa9.ngrok.io"
	gitHubTestOrg = "autograder-test"
	secret        = "the-secret-autograder-test"
)

// To run this test, please see instructions in the developer guide (dev.md).

// These tests only test listing existing hooks and creating a new one.
// They do not cover processing push events on a server.
// See web/hooks package for tests involving processing push events.

func TestGetOrganization(t *testing.T) {
	qfTestOrg := os.Getenv("QF_TEST_ORG")
	if len(qfTestOrg) < 1 {
		qfTestOrg = "qf101"
		t.Logf("This test requires access to the '%s' GitHub organization; to use another organization set the 'QF_TEST_ORG' environment variable", qfTestOrg)
	}
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skipf("This test requires that 'GITHUB_ACCESS_TOKEN' is set and that you have access to the '%v' GitHub organization", qfTestOrg)
	}

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	org, err := s.GetOrganization(context.Background(), &scm.GetOrgOptions{Name: qfTestOrg})
	if err != nil {
		t.Fatal(err)
	}
	if qfTestOrg == "qf101" {
		if org.ID != 77283363 {
			t.Errorf("scm.GetOrganization('qf101') = %d, expected 77283363", org.ID)
		}
	} else {
		// Otherwise we just print the organization result
		t.Logf("org: %v", org)
	}
}

func TestListHooks(t *testing.T) {
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' and access to the 'autograder-test' GitHub organization")
	}

	var s scm.SCM
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	hooks, err := s.ListHooks(ctx, nil, gitHubTestOrg)
	if err != nil {
		t.Fatal(err)
	}
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}

	hooks, err = s.ListHooks(ctx, &scm.Repository{Owner: gitHubTestOrg, Path: "tests"}, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}

	hooks, err = s.ListHooks(ctx, &scm.Repository{Path: "tests"}, "")
	if err == nil {
		t.Fatal("expected error 'ListHooks: called with missing or incompatible arguments: ...'")
	}
	t.Logf("%v %v", hooks, err)
}

func TestCreateHook(t *testing.T) {
	t.Skip("Disabled for now; need to add new method DeleteHook() before enabling again")
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' and access to the 'autograder-test' GitHub organization")
	}

	var s scm.SCM
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	opt := &scm.CreateHookOptions{
		URL:        serverURL,
		Secret:     secret,
		Repository: &scm.Repository{Owner: gitHubTestOrg, Path: "tests"},
	}
	err = s.CreateHook(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	hooks, err := s.ListHooks(ctx, opt.Repository, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}
}
