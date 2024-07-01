package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
)

// CreateIssue implements the SCM interface
func (s *GithubSCM) CreateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	const op Op = "CreateIssue"
	m := M("failed to create issue %q on %s/%s", opt.Title, opt.Organization, opt.Repository)
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", opt))
	}
	newIssue := &github.IssueRequest{
		Title:     &opt.Title,
		Body:      &opt.Body,
		Assignee:  opt.Assignee,
		Assignees: opt.Assignees,
	}
	issue, _, err := s.client.Issues.Create(ctx, opt.Organization, opt.Repository, newIssue)
	if err != nil {
		return nil, E(op, m, fmt.Errorf("%s: %w", m, err))
	}
	s.logger.Debugf("Created issue %q on %s/%s", opt.Title, opt.Organization, opt.Repository)
	return toIssue(issue), nil
}

// UpdateIssue implements the SCM interface
func (s *GithubSCM) UpdateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	const op Op = "UpdateIssue"
	m := M("failed to update issue #%d on %s/%s", opt.Number, opt.Organization, opt.Repository)
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", opt))
	}
	issueReq := &github.IssueRequest{
		Title:     &opt.Title,
		Body:      &opt.Body,
		State:     &opt.State,
		Assignee:  opt.Assignee,
		Assignees: opt.Assignees,
	}
	issue, _, err := s.client.Issues.Edit(ctx, opt.Organization, opt.Repository, opt.Number, issueReq)
	if err != nil {
		return nil, E(op, m, fmt.Errorf("%s: %w", m, err))
	}
	s.logger.Debugf("Updated issue number %d on %s/%s", opt.Number, opt.Organization, opt.Repository)
	return toIssue(issue), nil
}

// GetIssue implements the SCM interface
func (s *GithubSCM) GetIssue(ctx context.Context, opt *RepositoryOptions, number int) (*Issue, error) {
	const op Op = "GetIssue"
	m := M("failed to get issue #%d on %s/%s", number, opt.Owner, opt.Path)
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", opt))
	}
	issue, _, err := s.client.Issues.Get(ctx, opt.Owner, opt.Path, number)
	if err != nil {
		return nil, E(op, m, fmt.Errorf("%s: %w", m, err))
	}
	return toIssue(issue), nil
}

// GetIssues implements the SCM interface
func (s *GithubSCM) GetIssues(ctx context.Context, opt *RepositoryOptions) ([]*Issue, error) {
	const op Op = "GetIssues"
	m := M("failed to get issues for %s/%s", opt.Owner, opt.Path)
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", opt))
	}
	issueList, _, err := s.client.Issues.ListByRepo(ctx, opt.Owner, opt.Path, &github.IssueListByRepoOptions{})
	if err != nil {
		return nil, E(op, m, fmt.Errorf("%s: %w", m, err))
	}
	var issues []*Issue
	for _, issue := range issueList {
		issues = append(issues, toIssue(issue))
	}
	return issues, nil
}

// CreateIssueComment implements the SCM interface
func (s *GithubSCM) CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error) {
	const op Op = "CreateIssueComment"
	m := M("failed to create comment for issue #%d on %s/%s", opt.Number, opt.Organization, opt.Repository)
	if !opt.valid() {
		return 0, E(op, m, fmt.Errorf("missing fields: %+v", opt))
	}
	createdComment, _, err := s.client.Issues.CreateComment(ctx, opt.Organization, opt.Repository, opt.Number, &github.IssueComment{Body: &opt.Body})
	if err != nil {
		return 0, E(op, m, fmt.Errorf("%s: %w", m, err))
	}
	return createdComment.GetID(), nil
}

// UpdateIssueComment implements the SCM interface
func (s *GithubSCM) UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error {
	const op Op = "UpdateIssueComment"
	m := M("failed to update comment %d on issue #%d on %s/%s", opt.CommentID, opt.Number, opt.Organization, opt.Repository)
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", opt))
	}
	if _, _, err := s.client.Issues.EditComment(ctx, opt.Organization, opt.Repository, opt.CommentID, &github.IssueComment{Body: &opt.Body}); err != nil {
		return E(op, m, fmt.Errorf("%s: %w", m, err))
	}
	return nil
}

// RequestReviewers implements the SCM interface
func (s *GithubSCM) RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error {
	const op Op = "RequestReviewers"
	m := M("failed to request reviewers for pull request #%d on %s/%s", opt.Number, opt.Organization, opt.Repository)
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", opt))
	}
	reviewersRequest := github.ReviewersRequest{
		Reviewers: opt.Reviewers,
	}
	if _, _, err := s.client.PullRequests.RequestReviewers(ctx, opt.Organization, opt.Repository, opt.Number, reviewersRequest); err != nil {
		return E(op, m, fmt.Errorf("%s: %w", m, err))
	}
	return nil
}

func toIssue(issue *github.Issue) *Issue {
	return &Issue{
		ID:         issue.GetID(),
		Title:      issue.GetTitle(),
		Body:       issue.GetBody(),
		Repository: issue.Repository.GetName(),
		Assignee:   issue.Assignee.GetName(),
		Number:     issue.GetNumber(),
		Status:     issue.GetState(),
	}
}
