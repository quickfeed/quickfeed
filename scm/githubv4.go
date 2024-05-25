package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/shurcooL/githubv4"
)

func (s *GithubSCM) DeleteIssue(ctx context.Context, opt *RepositoryOptions, issueNumber int) error {
	var q struct {
		Repository struct {
			Issue struct {
				ID githubv4.ID
			} `graphql:"issue(number:$issueNumber)"`
		} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
	}
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(opt.Owner),
		"repositoryName":  githubv4.String(opt.Path),
		"issueNumber":     githubv4.Int(issueNumber),
	}
	err := s.clientV4.Query(ctx, &q, variables)
	if err != nil {
		return err
	}

	var m struct {
		DeleteIssue struct {
			Repository struct {
				Name string
			}
		} `graphql:"deleteIssue(input:$input)"`
	}

	input := githubv4.DeleteIssueInput{
		IssueID: q.Repository.Issue.ID,
	}
	return s.clientV4.Mutate(ctx, &m, input, nil)
}

func (s *GithubSCM) DeleteIssues(ctx context.Context, opt *RepositoryOptions) error {
	// List all open and closed issues (and pull requests)
	issueList, _, err := s.client.Issues.ListByRepo(ctx, opt.Owner, opt.Path, &github.IssueListByRepoOptions{State: "all"})
	if err != nil {
		return fmt.Errorf("failed to fetch issues for %s: %w", opt.Path, err)
	}
	for _, issue := range issueList {
		if issue.IsPullRequest() {
			continue // ignore pull requests when deleting issues
		}
		if err = s.DeleteIssue(ctx, opt, *issue.Number); err != nil {
			return fmt.Errorf("failed to delete issue %d in %s: %w", *issue.Number, opt.Path, err)
		}
	}
	return nil
}
