package scm_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"golang.org/x/oauth2"
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
	s, qfTestUser := scm.GetTestSCM(t)
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

func TestGetIssues(t *testing.T) {
	s := scm.NewMockSCMClient()
	s.Repositories = map[uint64]*scm.Repository{
		1: {
			ID:    1,
			OrgID: 1,
			Owner: qtest.MockOrg,
			Path:  qf.StudentRepoName("test"),
		},
	}

	ctx := context.Background()
	opt := &scm.RepositoryOptions{
		Owner: qtest.MockOrg,
		Path:  qf.StudentRepoName("test"),
	}

	wantIssueIDs := []int{}
	for i := 1; i <= 5; i++ {
		issue, cleanup := createIssue(t, s, opt.Owner, opt.Path)
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
	s, qfTestUser := scm.GetTestSCM(t)

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
	s, qfTestUser := scm.GetTestSCM(t)

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

	client := githubTestClient(t)
	ctx := context.Background()
	pullReq, _, err := client.PullRequests.Create(ctx, qfTestOrg, repo, &github.NewPullRequest{
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

	_, _, err = client.PullRequests.Edit(ctx, qfTestOrg, repo, *pullReq.Number, &github.PullRequest{
		State: github.String("closed"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("PullRequest %d closed", *pullReq.Number)
}

func githubTestClient(t *testing.T) *github.Client {
	t.Helper()
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: scm.GetAccessToken(t)})
	return github.NewClient(oauth2.NewClient(context.Background(), src))
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

func TestUpdateIssueComment(t *testing.T) {
	s := scm.NewMockSCMClient()
	repo := &scm.Repository{
		ID:    1,
		OrgID: 1,
		Owner: qtest.MockOrg,
		Path:  qf.StudentRepoName("user"),
	}
	s.Repositories = map[uint64]*scm.Repository{
		1: repo,
	}

	body := "Issue Comment"
	opt := &scm.IssueCommentOptions{
		Organization: repo.Owner,
		Repository:   repo.Path,
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
	ctx := context.Background()

	tests := []struct {
		name      string
		opt       *scm.RepositoryOptions
		wantEmpty bool
	}{
		{
			"tests repo, assume not empty",
			&scm.RepositoryOptions{
				Path:  "tests",
				Owner: qfTestOrg,
			},
			false,
		},
		{
			"info repo, assume empty",
			&scm.RepositoryOptions{
				Path:  "info",
				Owner: qfTestOrg,
			},
			true,
		},
		{
			"non-existent repo, handle as empty",
			&scm.RepositoryOptions{
				Path:  "some-other-repo",
				Owner: qfTestOrg,
			},
			true,
		},
	}
	for _, tt := range tests {
		if empty := s.RepositoryIsEmpty(ctx, tt.opt); empty != tt.wantEmpty {
			t.Errorf("%s: expected empty repository: %v, got = %v, ", tt.name, tt.wantEmpty, empty)
		}
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
