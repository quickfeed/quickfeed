package scm_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestDeleteIssue(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

	ctx := context.Background()
	repo, err := s.GetRepository(ctx, &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  qf.StudentRepoName(qfTestUser),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("repo: %v", repo.Path)

	opt := &scm.IssueOptions{
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

	err = s.DeleteIssue(ctx, repoOpt, issue.Number)
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
// The test is skipped unless run with: SCM_TESTS=1 go test -v -run TestDeleteAllIssues
func TestDeleteAllIssues(t *testing.T) {
	if os.Getenv("SCM_TESTS") == "" {
		t.SkipNow()
	}
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

	ctx := context.Background()
	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  qf.StudentRepoName(qfTestUser),
	}
	if err := s.DeleteIssues(ctx, opt); err != nil {
		t.Fatal(err)
	}
}
