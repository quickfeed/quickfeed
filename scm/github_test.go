package scm_test

import (
	"context"
	"os"
	"testing"

	"github.com/autograde/aguis/scm"
	"go.uber.org/zap"
)

const (
	serverURL       = "https://53c51fa9.ngrok.io"
	gitHubTestOrg   = "autograder-test"
	gitHubTestOrgID = 30462712
	secret          = "the-secret-autograder-test"
)

// To enable this test, please see instructions in the developer guide (dev.md).
// You will also need access to the autograder-test organization; you may request
// access by sending your GitHub username to hein.meling at uis.no.

// These tests only test listing existing hooks and creating a new one.
// They do not cover processing push events on a server.
// See web/hooks package for tests involving processing push events.

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

	hooks, err = s.ListHooks(ctx, &scm.Repository{Owner: gitHubTestOrg, Path: "tests"}, gitHubTestOrg)
	if err == nil {
		t.Fatal("expected error 'ListHooks: called with missing or incompatible arguments: ...'")
	}
	t.Logf("%v %v", hooks, err)
}

func TestCreateHook(t *testing.T) {
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
