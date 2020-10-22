package scm_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/scm"
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

func TestGetOrganization(t *testing.T) {
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
	orgOpts := &scm.GetOrgOptions{
		ID:       0,
		Name:     "dat320-2020",
		Username: "meling",
	}
	org, err := s.GetOrganization(ctx, orgOpts)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("org: %v", org)
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

func TestDeleteSubscription(t *testing.T) {
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' and access to the 'autograder-test' GitHub organization")
	}

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	repos, err := s.ListWatched(ctx)
	if err != nil {
		t.Fatal(err)
	}

	ignore := func(path string) bool {
		ignoreRepos := []string{"assignments", "tests", "course-info", "meling-labs", "meling-stud-labs"}
		for _, ignore := range ignoreRepos {
			if strings.Contains(path, ignore) {
				return true
			}
		}
		return false
	}

	// orgs:=[]string{"uis-dat320-fall2014", "uis-dat630-fall2015", "uis-dat320", "uis-dat520-s16", "uis-dat320-fall16", "dat520-2017", "uis-dat320-fall17"}
	// orgs := []string{"dat320-2020"}
	orgs := []string{"uis-dat520", "uis-dat520-s18", "uis-dat100", "uis-dat520-s2019", "uis-dat240-test", "uis-dat240-fall19", "dat320-2019", "dat520-2020", "dat310-spring20"}
	for _, org := range orgs {
		for _, repo := range repos {
			if strings.Contains(repo.Owner, org) && !ignore(repo.Path) {
				t.Logf("deleting subscription for repo: %v/%v", repo.Owner, repo.Path)
				err = s.DeleteRepositorySubscription(ctx, &scm.RepositoryOptions{ID: repo.ID})
				if err != nil {
					t.Fatal(err)
				}
			} else {
				t.Logf("remaining repo subscription: %v/%v", repo.Owner, repo.Path)
			}
		}
	}
}
