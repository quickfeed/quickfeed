package scm_test

import (
	"context"
	"testing"

	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
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
		// Otherwise, we just print the organization result
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
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	serverURL := scm.GetWebHookServer(t)
	// Only enable this test to add a new webhook to your test course organization
	if serverURL == "" {
		t.Skip("Disabled pending support for deleting webhooks")
	}

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

// Test case for Creating new Issue on a git Repository
func TestCreateIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// Replace with Repository name
	repo := "Replace with Repository Name"
	// Add Issue Title here
	title := "Replace with  Issue Title"
	// Add Issue body here
	body := "Replace with Issue Body"
	// Creating new Client
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	opt := &scm.CreateIssueOptions{
		Organization: qfTestOrg,
		Repository:   repo,
		Title:        title,
		Body:         body,
	}
	_, err = s.CreateIssue(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetIssues(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// Replace with Repository name
	repo := "Replace with Repository Name"

	// Creating new Client
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	opt := &scm.IssueOptions{
		Organization: qfTestOrg,
		Repository:   repo,
	}
	_, err = s.GetRepoIssues(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// Replace with Repository name
	repo := "Replace with Repository Name"
	// Replace 0 with Issue Number in Repository
	issueNumber := 0

	// Creating new Client
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	opt := &scm.IssueOptions{
		Organization: qfTestOrg,
		Repository:   repo,
		IssueNumber:  issueNumber,
	}
	_, err = s.GetRepoIssue(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}
}

// Test case for Updating existing Issue in a git Repository
func TestEditRepoIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	// Replace with Repository name
	repo := "Testing"
	// Add Issue Title here
	title := "Replace with new Issue Title"
	// Add Issue body here
	body := "Replace with new Issue Body"
	// Add Issue Number here
	issueNumber := 20
	// Creating new Client
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	opt1 := &scm.IssueOptions{
		Organization: qfTestOrg,
		Repository:   repo,
		IssueNumber:  issueNumber,
	}

	opt2 := &scm.CreateIssueOptions{
		Organization: qfTestOrg,
		Repository:   repo,
		Title:        title,
		Body:         body,
	}

	_, err = s.EditRepoIssue(ctx, opt1, opt2)
	if err != nil {
		t.Fatal(err)
	}
}
