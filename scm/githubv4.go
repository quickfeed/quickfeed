package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v35/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type GithubV4SCM struct {
	GithubSCM
	clientV4     *githubv4.Client
	bypassClient *github.Client
}

func NewGithubV4SCMClient(logger *zap.SugaredLogger, token string) *GithubV4SCM {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := github.NewClient(httpClient)
	return &GithubV4SCM{
		GithubSCM: GithubSCM{
			logger: logger,
			client: client,
			token:  token,
		},
		clientV4:     githubv4.NewClient(httpClient),
		bypassClient: client,
	}
}

func (s *GithubV4SCM) DeleteIssue(ctx context.Context, opt *RepositoryOptions, issueNumber int) error {
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

func (s *GithubV4SCM) DeleteIssues(ctx context.Context, opt *RepositoryOptions) error {
	// List all open and closed issues
	issueList, _, err := s.bypassClient.Issues.ListByRepo(ctx, opt.Owner, opt.Path, &github.IssueListByRepoOptions{State: "all"})
	if err != nil {
		return fmt.Errorf("failed to fetch issues for %s: %w", opt.Path, err)
	}
	for _, issue := range issueList {
		if err = s.DeleteIssue(ctx, opt, *issue.Number); err != nil {
			return fmt.Errorf("failed to delete issue %d in %s: %w", *issue.Number, opt.Path, err)
		}
	}
	return nil
}
