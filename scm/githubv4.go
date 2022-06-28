package scm

import (
	"context"

	"github.com/google/go-github/v35/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type GithubV4SCM struct {
	GithubSCM
	clientV4 *githubv4.Client
}

func NewGithubV4SCMClient(logger *zap.SugaredLogger, token string) *GithubV4SCM {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return &GithubV4SCM{
		GithubSCM: GithubSCM{
			logger: logger,
			client: github.NewClient(httpClient),
			token:  token,
		},
		clientV4: githubv4.NewClient(httpClient),
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
