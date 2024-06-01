package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
)

// CreateIssue implements the SCM interface
func (s *GithubSCM) CreateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	newIssue := &github.IssueRequest{
		Title:     &opt.Title,
		Body:      &opt.Body,
		Assignee:  opt.Assignee,
		Assignees: opt.Assignees,
	}
	issue, _, err := s.client.Issues.Create(ctx, opt.Organization, opt.Repository, newIssue)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "CreateIssue",
			Message:  fmt.Sprintf("failed to create issue %q on %s/%s", opt.Title, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	s.logger.Debugf("Created issue %q on %s/%s", opt.Title, opt.Organization, opt.Repository)
	return toIssue(issue), nil
}

// UpdateIssue implements the SCM interface
func (s *GithubSCM) UpdateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
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
		return nil, ErrFailedSCM{
			Method:   "UpdateIssue",
			Message:  fmt.Sprintf("failed to update issue #%d on %s/%s", opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	s.logger.Debugf("Updated issue number %d on %s/%s", opt.Number, opt.Organization, opt.Repository)
	return toIssue(issue), nil
}

// GetIssue implements the SCM interface
func (s *GithubSCM) GetIssue(ctx context.Context, opt *RepositoryOptions, number int) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	issue, _, err := s.client.Issues.Get(ctx, opt.Owner, opt.Path, number)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "GetIssue",
			Message:  fmt.Sprintf("failed to get issue #%d on %s/%s", number, opt.Owner, opt.Path),
			GitError: err,
		}
	}
	return toIssue(issue), nil
}

// GetIssues implements the SCM interface
func (s *GithubSCM) GetIssues(ctx context.Context, opt *RepositoryOptions) ([]*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	issueList, _, err := s.client.Issues.ListByRepo(ctx, opt.Owner, opt.Path, &github.IssueListByRepoOptions{})
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "GetIssues",
			Message:  fmt.Sprintf("failed to get issues for %s/%s", opt.Owner, opt.Path),
			GitError: err,
		}
	}
	var issues []*Issue
	for _, issue := range issueList {
		issues = append(issues, toIssue(issue))
	}
	return issues, nil
}

// CreateIssueComment implements the SCM interface
func (s *GithubSCM) CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error) {
	if !opt.valid() {
		return 0, fmt.Errorf("missing fields: %+v", opt)
	}
	createdComment, _, err := s.client.Issues.CreateComment(ctx, opt.Organization, opt.Repository, opt.Number, &github.IssueComment{Body: &opt.Body})
	if err != nil {
		return 0, ErrFailedSCM{
			Method:   "CreateIssueComment",
			Message:  fmt.Sprintf("failed to create comment for issue #%d on %s/%s", opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	return createdComment.GetID(), nil
}

// UpdateIssueComment implements the SCM interface
func (s *GithubSCM) UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	if _, _, err := s.client.Issues.EditComment(ctx, opt.Organization, opt.Repository, opt.CommentID, &github.IssueComment{Body: &opt.Body}); err != nil {
		return ErrFailedSCM{
			Method:   "UpdateIssueComment",
			Message:  fmt.Sprintf("failed to edit comment %d on issue #%d on %s/%s", opt.CommentID, opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	return nil
}

// RequestReviewers implements the SCM interface
func (s *GithubSCM) RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	reviewersRequest := github.ReviewersRequest{
		Reviewers: opt.Reviewers,
	}
	if _, _, err := s.client.PullRequests.RequestReviewers(ctx, opt.Organization, opt.Repository, opt.Number, reviewersRequest); err != nil {
		return ErrFailedSCM{
			Method:   "RequestReviewers",
			Message:  fmt.Sprintf("failed to request reviewers for pull request #%d on %s/%s", opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	return nil
}

func toIssue(issue *github.Issue) *Issue {
	return &Issue{
		ID:         uint64(issue.GetID()),
		Title:      issue.GetTitle(),
		Body:       issue.GetBody(),
		Repository: issue.Repository.GetName(),
		Assignee:   issue.Assignee.GetName(),
		Number:     issue.GetNumber(),
		Status:     issue.GetState(),
	}
}
