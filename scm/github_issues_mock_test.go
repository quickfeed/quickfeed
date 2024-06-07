package scm

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

func TestMockCreateIssue(t *testing.T) {
	wantIssues := map[string]map[string][]*Issue{
		"foo": {
			"meling-labs": {
				{ID: 1, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 2, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
			"josie-labs": {
				{ID: 3, Number: 1, Title: "First", Body: "xyz", Repository: "josie-labs"},
				{ID: 4, Number: 2, Title: "Second", Body: "abc", Repository: "josie-labs"},
			},
		},
		"bar": {
			"meling-labs": {
				{ID: 5, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 6, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
		},
	}

	tests := []struct {
		name      string
		opt       *IssueOptions
		wantIssue *Issue
		wantErr   bool
	}{
		{name: "IncompleteRequest", opt: &IssueOptions{}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Title: "Hello", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Organization: "foo", Title: "Hello", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "Hello"}, wantIssue: nil, wantErr: true},

		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"}, wantIssue: wantIssues["foo"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"}, wantIssue: wantIssues["foo"]["meling-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"}, wantIssue: wantIssues["foo"]["josie-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"}, wantIssue: wantIssues["foo"]["josie-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "bar", Repository: "meling-labs", Title: "First", Body: "xyz"}, wantIssue: wantIssues["bar"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "bar", Repository: "meling-labs", Title: "Second", Body: "abc"}, wantIssue: wantIssues["bar"]["meling-labs"][1], wantErr: false},
	}

	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "Repository", "Title", "Body", "Number"}, tt.opt.Organization, tt.opt.Repository, tt.opt.Title, tt.opt.Body, tt.opt.Number)
		t.Run(name, func(t *testing.T) {
			issue, err := s.CreateIssue(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateIssue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantIssue, issue); diff != "" {
				t.Errorf("CreateIssue() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockUpdateIssue(t *testing.T) {
	wantIssues := map[string]map[string][]*Issue{
		"foo": {
			"meling-labs": {
				{ID: 1, Number: 1, Title: "First 1", Body: "First Body", Repository: "meling-labs"},
				{ID: 2, Number: 2, Title: "Second 2", Body: "Second Body", Repository: "meling-labs"},
			},
			"josie-labs": {
				{ID: 3, Number: 1, Title: "First 1", Body: "First Body", Repository: "josie-labs"},
				{ID: 4, Number: 2, Title: "Second 2", Body: "Second Body", Repository: "josie-labs"},
			},
		},
		"bar": {
			"meling-labs": {
				{ID: 5, Number: 1, Title: "First 1", Body: "First Body", Repository: "meling-labs"},
				{ID: 6, Number: 2, Title: "Second 2", Body: "Second Body", Repository: "meling-labs"},
			},
		},
	}

	tests := []struct {
		name      string
		opt       *IssueOptions
		wantIssue *Issue
		wantErr   bool
	}{
		{name: "IncompleteRequest", opt: &IssueOptions{}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Title: "Hello", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Organization: "foo", Title: "Hello", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "Hello"}, wantIssue: nil, wantErr: true},

		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "First 1", Body: "First Body", Number: 1}, wantIssue: wantIssues["foo"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "Second 2", Body: "Second Body", Number: 2}, wantIssue: wantIssues["foo"]["meling-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "First 1", Body: "First Body", Number: 1}, wantIssue: wantIssues["foo"]["josie-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "Second 2", Body: "Second Body", Number: 2}, wantIssue: wantIssues["foo"]["josie-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "bar", Repository: "meling-labs", Title: "First 1", Body: "First Body", Number: 1}, wantIssue: wantIssues["bar"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "bar", Repository: "meling-labs", Title: "Second 2", Body: "Second Body", Number: 2}, wantIssue: wantIssues["bar"]["meling-labs"][1], wantErr: false},
	}

	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "Repository", "Title", "Body", "Number"}, tt.opt.Organization, tt.opt.Repository, tt.opt.Title, tt.opt.Body, tt.opt.Number)
		t.Run(name, func(t *testing.T) {
			issue, err := s.UpdateIssue(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateIssue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(issue, tt.wantIssue); diff != "" {
				t.Errorf("UpdateIssue() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockGetIssue(t *testing.T) {
	wantIssues := map[string]map[string][]*Issue{
		"foo": {
			"meling-labs": {
				{ID: 1, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 2, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
			"josie-labs": {
				{ID: 3, Number: 1, Title: "First", Body: "xyz", Repository: "josie-labs"},
				{ID: 4, Number: 2, Title: "Second", Body: "abc", Repository: "josie-labs"},
			},
		},
		"bar": {
			"meling-labs": {
				{ID: 5, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 6, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
		},
	}

	tests := []struct {
		name      string
		opt       *RepositoryOptions
		number    int
		wantIssue *Issue
		wantErr   bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Path: "meling-labs"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo", Path: "meling-labs"}, wantIssue: nil, wantErr: true},

		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Path: "meling-labs"}, number: 1, wantIssue: wantIssues["foo"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Path: "meling-labs"}, number: 2, wantIssue: wantIssues["foo"]["meling-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Path: "josie-labs"}, number: 1, wantIssue: wantIssues["foo"]["josie-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Path: "josie-labs"}, number: 2, wantIssue: wantIssues["foo"]["josie-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "bar", Path: "meling-labs"}, number: 1, wantIssue: wantIssues["bar"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "bar", Path: "meling-labs"}, number: 2, wantIssue: wantIssues["bar"]["meling-labs"][1], wantErr: false},
	}

	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Path)
		t.Run(name, func(t *testing.T) {
			issue, err := s.GetIssue(context.Background(), tt.opt, tt.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(issue, tt.wantIssue); diff != "" {
				t.Errorf("GetIssue() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockGetIssues(t *testing.T) {
	wantIssues := map[string]map[string][]*Issue{
		"foo": {
			"meling-labs": {
				{ID: 1, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 2, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
			"josie-labs": {
				{ID: 3, Number: 1, Title: "First", Body: "xyz", Repository: "josie-labs"},
				{ID: 4, Number: 2, Title: "Second", Body: "abc", Repository: "josie-labs"},
			},
		},
		"bar": {
			"meling-labs": {
				{ID: 5, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 6, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
		},
	}

	tests := []struct {
		name       string
		opt        *RepositoryOptions
		wantIssues []*Issue
		wantErr    bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantIssues: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantIssues: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Path: "meling-labs"}, wantIssues: nil, wantErr: true},

		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Path: "meling-labs"}, wantIssues: wantIssues["foo"]["meling-labs"], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Path: "josie-labs"}, wantIssues: wantIssues["foo"]["josie-labs"], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "bar", Path: "meling-labs"}, wantIssues: wantIssues["bar"]["meling-labs"], wantErr: false},
	}

	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Path)
		t.Run(name, func(t *testing.T) {
			issues, err := s.GetIssues(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(issues, tt.wantIssues); diff != "" {
				t.Errorf("GetIssues() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockCreateIssueComment(t *testing.T) {
	tests := []struct {
		name          string
		opt           *IssueCommentOptions
		wantCommentID int64
		wantErr       bool
	}{
		{name: "IncompleteRequest", opt: &IssueCommentOptions{}, wantCommentID: 0, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Body: "Hello"}, wantCommentID: 0, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Body: "Hello"}, wantCommentID: 0, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs"}, wantCommentID: 0, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Body: "Hello"}, wantCommentID: 0, wantErr: true},

		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Number: 1, Body: "Hello 1.1"}, wantCommentID: 1, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Number: 1, Body: "Hello 1.2"}, wantCommentID: 2, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Number: 1, Body: "Hello 1.3"}, wantCommentID: 3, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Number: 2, Body: "Hello 2.1"}, wantCommentID: 4, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Number: 2, Body: "Hello 2.2"}, wantCommentID: 5, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Number: 2, Body: "Hello 2.3"}, wantCommentID: 6, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", Number: 1, Body: "Hello 1.1"}, wantCommentID: 7, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", Number: 1, Body: "Hello 1.2"}, wantCommentID: 8, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", Number: 1, Body: "Hello 1.3"}, wantCommentID: 9, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", Number: 2, Body: "Hello 2.1"}, wantCommentID: 10, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", Number: 2, Body: "Hello 2.2"}, wantCommentID: 11, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", Number: 2, Body: "Hello 2.3"}, wantCommentID: 12, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "bar", Repository: "meling-labs", Number: 1, Body: "Hello 1.1"}, wantCommentID: 13, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "bar", Repository: "meling-labs", Number: 1, Body: "Hello 1.2"}, wantCommentID: 14, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "bar", Repository: "meling-labs", Number: 1, Body: "Hello 1.3"}, wantCommentID: 15, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "bar", Repository: "meling-labs", Number: 2, Body: "Hello 2.1"}, wantCommentID: 16, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "bar", Repository: "meling-labs", Number: 2, Body: "Hello 2.2"}, wantCommentID: 17, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "bar", Repository: "meling-labs", Number: 2, Body: "Hello 2.3"}, wantCommentID: 18, wantErr: false},
	}

	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(
			tt.name,
			[]string{"Organization", "Repository", "Number", "Body"},
			tt.opt.Organization, tt.opt.Repository, tt.opt.Number, tt.opt.Body,
		)
		t.Run(name, func(t *testing.T) {
			commentID, err := s.CreateIssueComment(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateIssueComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if commentID != tt.wantCommentID {
				t.Errorf("CreateIssueComment() = %v, want %v", commentID, tt.wantCommentID)
			}
		})
	}
}
