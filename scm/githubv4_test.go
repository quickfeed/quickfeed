package scm_test

import (
	"context"
	"testing"
	"time"

	"github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
)

func TestDeleteIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	qfTestUser := scm.GetTestUser(t)

	s := scm.NewGithubV4SCMClient(zap.NewNop().Sugar(), accessToken)

	ctx := context.Background()
	repo, err := s.GetRepository(ctx, &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  ag.StudentRepoName(qfTestUser),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("repo: %v", repo.Path)

	opt := &scm.CreateIssueOptions{
		Organization: qfTestOrg,
		Repository:   repo.Path,
		Title:        "Dummy Title",
		Body:         "Dummy body of the issue",
	}
	issue, err := s.CreateIssue(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	repoOpt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  repo.Path,
	}
	issues, err := s.GetIssues(ctx, repoOpt)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range issues {
		t.Logf("with new issue: %v", issue)
	}

	err = s.DeleteIssue(ctx, repoOpt, issue.IssueNumber)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("---------")

	time.Sleep(200 * time.Millisecond)

	issues, err = s.GetIssues(ctx, repoOpt)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range issues {
		t.Logf("after deleting issue: %v", issue)
	}
}

// This test will delete all open and closed issues for the test user and organization.
// The test is normally disabled, since its use is mainly for testing with GitHub test repos.
func disabledTestDeleteAllIssues(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	qfTestUser := scm.GetTestUser(t)

	s := scm.NewGithubV4SCMClient(zap.NewNop().Sugar(), accessToken)

	ctx := context.Background()
	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  ag.StudentRepoName(qfTestUser),
	}
	// List all open and closed issues
	issueList, _, err := s.BypassClient.Issues.ListByRepo(ctx, opt.Owner, opt.Path, &github.IssueListByRepoOptions{State: "all"})
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range issueList {
		t.Logf("Deleting issue #%d", *issue.Number)
		if err = s.DeleteIssue(ctx, opt, *issue.Number); err != nil {
			t.Fatal(err)
		}
	}
}
