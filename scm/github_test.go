package scm_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

const (
	qf101Org   = "qf101"
	qf101OrdID = 77283363
)

// To run this test, please see instructions in the developer guide (dev.md).

func TestGetOrganization(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)
	org, err := s.GetOrganization(context.Background(), &scm.OrganizationOptions{
		Name:     qfTestOrg,
		Username: qfTestUser,
	})
	if err != nil {
		t.Fatal(err)
	}
	if qfTestOrg == qf101Org {
		if org.GetScmOrganizationID() != qf101OrdID {
			t.Errorf("GetOrganization(%q) = %d, expected %d", qfTestOrg, org.GetScmOrganizationID(), qf101OrdID)
		}
	} else {
		// Otherwise, we just print the organization result
		t.Logf("org: %v", org)
	}
}

// Test case for Creating new Issue on a git Repository
func TestCreateIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

	issue, cleanup := createIssue(t, s, qfTestOrg, qf.StudentRepoName(qfTestUser))
	defer cleanup()

	if !(issue.Title == "Test Issue" && issue.Body == "Test Body") {
		t.Errorf("scm.TestCreateIssue: issue: %v", issue)
	}
}

func TestGetIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Repo:  qf.StudentRepoName(qfTestUser),
	}

	wantIssue, cleanup := createIssue(t, s, opt.Owner, opt.Repo)
	defer cleanup()

	gotIssue, err := s.GetIssue(context.Background(), opt, wantIssue.Number)
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
	s, qfTestUser := scm.GetTestSCM(t)

	opt := &scm.IssueOptions{
		Organization: qfTestOrg,
		Repository:   qf.StudentRepoName(qfTestUser),
		Title:        "Updated Issue",
		Body:         "Updated Issue Body",
	}

	issue, cleanup := createIssue(t, s, opt.Organization, opt.Repository)
	defer cleanup()

	opt.Number = issue.Number
	gotIssue, err := s.UpdateIssue(context.Background(), opt)
	if err != nil {
		t.Fatal(err)
	}

	if gotIssue.Title != opt.Title || gotIssue.Body != opt.Body {
		t.Errorf("scm.TestUpdateIssue() want (title: %s, body: %s), got (title: %s, body: %s)", opt.Title, opt.Body, gotIssue.Title, gotIssue.Body)
	}
}

// This test will delete all open and closed issues for the test user and organization.
// The test is skipped unless run with: SCM_TESTS=1 go test -v -run TestDeleteAllIssues
func TestDeleteAllIssues(t *testing.T) {
	if os.Getenv("SCM_TESTS") == "" {
		t.SkipNow()
	}
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Repo:  qf.StudentRepoName(qfTestUser),
	}
	if err := s.DeleteIssues(context.Background(), opt); err != nil {
		t.Fatal(err)
	}
}

// TestRequestReviewers tests the ability to request reviewers for a pull request.
// It will first create a pull request, then request reviewers for it and then closes the pull request.
//
// Note: This test requires manual steps before execution:
// 1. Create branch test-request-reviewers on the qfTestUser repo
// 2. Make edits on the test-request-reviewers branch
// 3. Push the changes to the qfTestUser repo
//
// The test is skipped unless run with: SCM_TESTS=1 go test -v -run TestRequestReviewers
func TestRequestReviewers(t *testing.T) {
	if os.Getenv("SCM_TESTS") == "" {
		t.SkipNow()
	}
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)
	repo := qf.StudentRepoName(qfTestUser)

	testReqReviewersBranch := "test-request-reviewers"

	ctx := context.Background()
	pullReq, _, err := s.Client().PullRequests.Create(ctx, qfTestOrg, repo, &github.NewPullRequest{
		Title: github.String("Test Request Reviewers"),
		Body:  github.String("Test Request Reviewers Body"),
		Head:  github.String(testReqReviewersBranch),
		Base:  github.String("master"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("PullRequest %d opened", *pullReq.Number)

	// Pick a reviewer that is not the current user (that created the PR above).
	var reviewer string
	for _, r := range []string{"meling", "JosteinLindhom"} {
		if r != qfTestUser {
			reviewer = r
			break
		}
	}

	opt := &scm.RequestReviewersOptions{
		Organization: qfTestOrg,
		Repository:   repo,
		Number:       *pullReq.Number,
		Reviewers:    []string{reviewer},
	}
	if err := s.RequestReviewers(ctx, opt); err != nil {
		t.Fatal(err)
	}
	t.Logf("PullRequest %d created with reviewer %v", *pullReq.Number, reviewer)

	_, _, err = s.Client().PullRequests.Edit(ctx, qfTestOrg, repo, *pullReq.Number, &github.PullRequest{
		State: github.String("closed"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("PullRequest %d closed", *pullReq.Number)
}

func TestCreateIssueComment(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

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

// TestFeedbackCommentFormat tests creating a feedback comment on a pull request, with the given result.
// Note: Manual step required to view the resulting comment: disable the cleanup() function.
// The test is skipped unless run with: SCM_TESTS=1 go test -v -run TestFeedbackCommentFormat
func TestFeedbackCommentFormat(t *testing.T) {
	if os.Getenv("SCM_TESTS") == "" {
		t.SkipNow()
	}
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

	opt := &scm.IssueCommentOptions{
		Organization: qfTestOrg,
		Repository:   qf.StudentRepoName(qfTestUser),
		Body:         "Some initial feedback",
	}
	// Using the IssueCommentOptions opt fields to create the issue; the opt will be used below.
	issue, cleanup := createIssue(t, s, opt.Organization, opt.Repository)
	defer cleanup()

	opt.Number = issue.Number
	// The created comment will be deleted when the parent issue is deleted.
	commentID, err := s.CreateIssueComment(context.Background(), opt)
	if err != nil {
		t.Fatal(err)
	}

	results := &score.Results{
		Scores: []*score.Score{
			{TestName: "Test1", TaskName: "1", Score: 5, MaxScore: 7, Weight: 2},
			{TestName: "Test2", TaskName: "1", Score: 3, MaxScore: 9, Weight: 3},
			{TestName: "Test3", TaskName: "1", Score: 8, MaxScore: 8, Weight: 5},
			{TestName: "Test4", TaskName: "1", Score: 2, MaxScore: 5, Weight: 1},
			{TestName: "Test5", TaskName: "1", Score: 5, MaxScore: 7, Weight: 1},
			{TestName: "Test6", TaskName: "2", Score: 5, MaxScore: 7, Weight: 1},
			{TestName: "Test7", TaskName: "3", Score: 5, MaxScore: 7, Weight: 1},
		},
	}
	body := results.MarkdownComment("1", 80)
	opt.CommentID = commentID
	opt.Body = body
	if err := s.UpdateIssueComment(context.Background(), opt); err != nil {
		t.Fatal(err)
	}
}

// This test assumes that the test organization has an empty "info" repository
// and non-empty "tests" repository.
func TestEmptyRepo(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, _ := scm.GetTestSCM(t)

	tests := []struct {
		name      string
		opt       *scm.RepositoryOptions
		wantEmpty bool
	}{
		{name: "NonEmptyRepo", opt: &scm.RepositoryOptions{Repo: "tests", Owner: qfTestOrg}, wantEmpty: false},
		{name: "NonEmptyRepo", opt: &scm.RepositoryOptions{ID: 328688692}, wantEmpty: false},
		{name: "EmptyRepo", opt: &scm.RepositoryOptions{Repo: "info", Owner: qfTestOrg}, wantEmpty: true},
		{name: "EmptyRepo", opt: &scm.RepositoryOptions{ID: 328688666}, wantEmpty: true},
		{name: "NonExistentRepo", opt: &scm.RepositoryOptions{Repo: "some-other-repo", Owner: qfTestOrg}, wantEmpty: true}, // treat non-existent repo as empty
	}
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"ID", "Owner", "Repo"}, tt.opt.ID, tt.opt.Owner, tt.opt.Repo)
		t.Run(name, func(t *testing.T) {
			if empty := s.RepositoryIsEmpty(context.Background(), tt.opt); empty != tt.wantEmpty {
				t.Errorf("RepositoryIsEmpty(%+v) = %t, want %t", *tt.opt, empty, tt.wantEmpty)
			}
		})
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
			Owner: org, Repo: repo,
		}, issue.Number); err != nil {
			t.Fatal(err)
		}
	}
}
