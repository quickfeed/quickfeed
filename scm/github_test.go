package scm_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
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
	qfTestUser := scm.GetTestUser(t)
	s := scm.GetTestSCM(t)
	org, err := s.GetOrganization(context.Background(), &scm.GetOrgOptions{
		Name:     qfTestOrg,
		Username: qfTestUser,
	})
	if err != nil {
		t.Fatal(err)
	}
	if qfTestOrg == qf101Org {
		if org.ID != qf101OrdID {
			t.Errorf("GetOrganization(%q) = %d, expected %d", qfTestOrg, org.ID, qf101OrdID)
		}
	} else {
		// Otherwise, we just print the organization result
		t.Logf("org: %v", org)
	}
}

func TestListHooks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s := scm.GetTestSCM(t)

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
	serverURL := scm.GetWebHookServer(t)
	// Only enable this test to add a new webhook to your test course organization
	if serverURL == "" {
		t.Skip("Disabled pending support for deleting webhooks")
	}

	s := scm.GetTestSCM(t)

	ctx := context.Background()
	opt := &scm.CreateHookOptions{
		URL:        serverURL,
		Secret:     secret,
		Repository: &scm.Repository{Owner: qfTestOrg, Path: "tests"},
	}
	err := s.CreateHook(ctx, opt)
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
	qfTestUser := scm.GetTestUser(t)
	s := scm.GetTestSCM(t)

	issue, cleanup := createIssue(t, s, qfTestOrg, qf.StudentRepoName(qfTestUser))
	defer cleanup()

	if !(issue.Title == "Test Issue" && issue.Body == "Test Body") {
		t.Errorf("scm.TestCreateIssue: issue: %v", issue)
	}
}

// NOTE: This test only works if the given repository has no previous issues
func TestGetIssues(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	qfTestUser := scm.GetTestUser(t)
	s := scm.GetTestSCM(t)

	ctx := context.Background()
	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  qf.StudentRepoName(qfTestUser),
	}

	wantIssueIDs := []int{}
	for i := 1; i <= 5; i++ {
		issue, cleanup := createIssue(t, s, qfTestOrg, opt.Path)
		defer cleanup()
		wantIssueIDs = append(wantIssueIDs, issue.Number)
	}

	gotIssueIDs := []int{}
	gotIssues, err := s.GetIssues(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range gotIssues {
		gotIssueIDs = append(gotIssueIDs, issue.Number)
	}

	less := func(a, b int) bool { return a < b }
	if equal := cmp.Equal(wantIssueIDs, gotIssueIDs, cmpopts.SortSlices(less)); !equal {
		t.Errorf("scm.GetIssues() mismatch wantIssueIDs: %v, gotIssueIDs: %v", wantIssueIDs, gotIssueIDs)
	}
}

func TestGetIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	qfTestUser := scm.GetTestUser(t)
	s := scm.GetTestSCM(t)

	ctx := context.Background()
	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  qf.StudentRepoName(qfTestUser),
	}

	wantIssue, cleanup := createIssue(t, s, opt.Owner, opt.Path)
	defer cleanup()

	gotIssue, err := s.GetIssue(ctx, opt, wantIssue.Number)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantIssue, gotIssue); diff != "" {
		t.Errorf("scm.TestGetIssue() mismatch (-wantIssue +gotIssue):\n%s", diff)
	}
}

// Test case for Updating existing Issue in a git Repository
func TestUpdateIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	qfTestUser := scm.GetTestUser(t)
	s := scm.GetTestSCM(t)

	ctx := context.Background()

	opt := &scm.IssueOptions{
		Organization: qfTestOrg,
		Repository:   qf.StudentRepoName(qfTestUser),
		Title:        "Updated Issue",
		Body:         "Updated Issue Body",
	}

	issue, cleanup := createIssue(t, s, opt.Organization, opt.Repository)
	defer cleanup()

	opt.Number = issue.Number
	gotIssue, err := s.UpdateIssue(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	if gotIssue.Title != opt.Title || gotIssue.Body != opt.Body {
		t.Fatalf("scm.TestUpdateIssue() want (title: %s, body: %s), got (title: %s, body: %s)", opt.Title, opt.Body, gotIssue.Title, gotIssue.Body)
	}
}

func TestRequestReviewers(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s := scm.GetTestSCM(t)

	// Set these when testing
	opt := &scm.RequestReviewersOptions{
		Organization: qfTestOrg,
		Repository:   "repo-name",
		Number:       1,
		Reviewers:    []string{"reviewer-login"},
	}
	err := s.RequestReviewers(context.Background(), opt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateIssueComment(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	qfTestUser := scm.GetTestUser(t)
	s := scm.GetTestSCM(t)

	body := "Test"
	opt := &scm.IssueCommentOptions{
		Organization: qfTestOrg,
		Repository:   qf.StudentRepoName(qfTestUser),
		Body:         body,
	}

	issue, cleanup := createIssue(t, s, opt.Organization, opt.Repository)
	defer cleanup()

	opt.Number = issue.Number
	_, err := s.CreateIssueComment(context.Background(), opt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateIssueComment(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	qfTestUser := scm.GetTestUser(t)
	s := scm.GetTestSCM(t)

	body := "Issue Comment"
	opt := &scm.IssueCommentOptions{
		Organization: qfTestOrg,
		Repository:   qf.StudentRepoName(qfTestUser),
		Body:         body,
	}

	issue, cleanup := createIssue(t, s, opt.Organization, opt.Repository)
	defer cleanup()

	opt.Number = issue.Number
	// The created comment will be deleted when the parent issue is deleted.
	commentID, err := s.CreateIssueComment(context.Background(), opt)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: We do not currently return the updated comment, so we cannot verify its content.
	opt.Body = "Updated Issue Comment"
	opt.CommentID = commentID
	if err := s.UpdateIssueComment(context.Background(), opt); err != nil {
		t.Fatal(err)
	}
}

// createIssue on the given repository; returns the issue and a cleanup function.
func createIssue(t *testing.T, s scm.SCM, org, repo string) (*scm.Issue, func()) {
	t.Helper()
	issue, err := s.CreateIssue(context.Background(), &scm.IssueOptions{
		Organization: org,
		Repository:   repo,
		Title:        "Test Issue",
		Body:         "Test Body",
	})
	if err != nil {
		t.Fatal(err)
	}

	return issue, func() {
		if err := s.DeleteIssue(context.Background(), &scm.RepositoryOptions{
			Owner: org, Path: repo,
		}, issue.Number); err != nil {
			t.Fatal(err)
		}
	}
}
