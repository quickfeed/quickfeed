package scm_test

import (
	"context"
	"testing"

	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/scm"
)

const (
	qf101Org   = "qf101"
	qf101OrdID = 77283363
	secret     = "the-secret-quickfeed-test"
)

// To run this test, please see instructions in the developer guide (dev.md).

// These tests only test listing existing hooks and creating a new one.
// They do not cover processing push events on a server.
// See web/hooks package for tests involving processing push events.

func TestGetOrganization(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	ctx := context.Background()
	s := qtest.TestSCMClient(ctx, t, qfTestOrg, "github", accessToken)
	org, err := s.GetOrganization(context.Background(), &scm.GetOrgOptions{Name: qfTestOrg})
	if err != nil {
		t.Fatal(err)
	}
	if qfTestOrg == qf101Org {
		if org.ID != qf101OrdID {
			t.Errorf("scm.GetOrganization('%s') = %d, expected %d", qfTestOrg, org.ID, qf101OrdID)
		}
	} else {
		// Otherwise, we just print the organization result
		t.Logf("org: %v", org)
	}
}

func TestListHooks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	ctx := context.Background()
	s := qtest.TestSCMClient(ctx, t, qfTestOrg, "github", accessToken)

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
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	serverURL := scm.GetWebHookServer(t)
	// Only enable this test to add a new webhook to your test course organization
	if serverURL == "" {
		t.Skip("Disabled pending support for deleting webhooks")
	}
	ctx := context.Background()
	s := qtest.TestSCMClient(ctx, t, qfTestOrg, "github", accessToken)

	opt := &scm.CreateHookOptions{
		URL:        serverURL,
		Secret:     secret,
		Repository: &scm.Repository{Owner: qfTestOrg, Path: "tests"},
	}
	if err := s.CreateHook(ctx, opt); err != nil {
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

// Test case for Creating new Issue on a git Repository
func TestCreateIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// TODO(vera): make a helper method to populate these from env variables or to fetch from the test given repo
	// Replace with Repository name
	repo := "test-labs"
	// Add Issue Title here
	title := "Test issue"
	// Add Issue body here
	body := "Test issue of testing"
	// Creating new Client
	ctx := context.Background()
	s := qtest.TestSCMClient(ctx, t, qfTestOrg, "github", accessToken)
	opt := &scm.CreateIssueOptions{
		Organization: qfTestOrg,
		Repository:   repo,
		Title:        title,
		Body:         body,
	}
	if _, err := s.CreateIssue(ctx, opt); err != nil {
		t.Fatal(err)
	}
}

func TestGetIssues(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// TODO(vera): make helper method to fetch repo/issue names
	// Replace with Repository name
	repo := "test-labs"

	ctx := context.Background()
	s := qtest.TestSCMClient(ctx, t, qfTestOrg, "github", accessToken)

	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  repo,
	}
	if _, err := s.GetRepoIssues(ctx, opt); err != nil {
		t.Fatal(err)
	}
}

func TestGetIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// Replace with Repository name
	// TODO(vera): make helper method to fetch repo/issue names, numbers
	repo := "test-labs"
	// Replace 0 with Issue Number in Repository
	issueNumber := 1

	ctx := context.Background()
	s := qtest.TestSCMClient(ctx, t, qfTestOrg, "github", accessToken)

	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  repo,
	}
	if _, err := s.GetRepoIssue(ctx, issueNumber, opt); err != nil {
		t.Fatal(err)
	}
}

// Test case for Updating existing Issue in a git Repository
func TestEditRepoIssue(t *testing.T) {
	// TODO(vera): make helper method to fetch repo/issue names, numbers
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// Replace with Repository name
	repo := "test-labs"
	// Add Issue Title here
	title := "Test issue"
	// Add Issue body here
	body := "Updated test issue"
	// Add Issue Number here
	issueNumber := 1

	ctx := context.Background()
	s := qtest.TestSCMClient(ctx, t, qfTestOrg, "github", accessToken)
	opt := &scm.CreateIssueOptions{
		Organization: qfTestOrg,
		Repository:   repo,
		Title:        title,
		Body:         body,
	}
	if _, err := s.EditRepoIssue(ctx, issueNumber, opt); err != nil {
		t.Fatal(err)
	}
}
