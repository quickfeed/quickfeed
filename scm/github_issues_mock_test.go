package scm

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
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
		"dat320": {
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

		{name: "CompleteRequest/OrgNotFound", opt: &IssueOptions{Organization: "buz", Repository: "meling-labs", Title: "First", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &IssueOptions{Organization: "foo", Repository: "lamport-labs", Title: "First", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/MissingTitle", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/MissingBody", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "First"}, wantIssue: nil, wantErr: true},

		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"}, wantIssue: wantIssues["foo"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"}, wantIssue: wantIssues["foo"]["meling-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"}, wantIssue: wantIssues["foo"]["josie-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"}, wantIssue: wantIssues["foo"]["josie-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"}, wantIssue: wantIssues["dat320"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"}, wantIssue: wantIssues["dat320"]["meling-labs"][1], wantErr: false},
	}

	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
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

func TestMockDeleteIssue(t *testing.T) {
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"},
		{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"},
		{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
	for _, opt := range createIssues {
		_, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
	}
	issues, err := s.GetIssues(context.Background(), &RepositoryOptions{Owner: "foo", Repo: "meling-labs"})
	if err != nil {
		t.Fatalf("failed to get issues: %v", err)
	}
	if len(issues) != 2 { // only two issues are created for foo/meling-labs
		t.Fatalf("expected %d issues, got %d", 2, len(issues))
	}

	tests := []struct {
		name    string
		opt     *RepositoryOptions
		number  int
		wantErr bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Repo: "meling-labs"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &RepositoryOptions{Owner: "buz", Repo: "meling-labs"}, number: 1, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &RepositoryOptions{Owner: "foo", Repo: "lamport-labs"}, number: 1, wantErr: true},
		{name: "CompleteRequest/MissingNumber", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, wantErr: true},
		{name: "CompleteRequest/WrongNumber", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, number: 543, wantErr: true},

		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, number: 1, wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, number: 2, wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "josie-labs"}, number: 1, wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "josie-labs"}, number: 2, wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "dat320", Repo: "meling-labs"}, number: 1, wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "dat320", Repo: "meling-labs"}, number: 2, wantErr: false},
	}
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Repo)
		t.Run(name, func(t *testing.T) {
			if err := s.DeleteIssue(context.Background(), tt.opt, tt.number); (err != nil) != tt.wantErr {
				t.Errorf("DeleteIssue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	issues, err = s.GetIssues(context.Background(), &RepositoryOptions{Owner: "foo", Repo: "meling-labs"})
	if err != nil {
		t.Fatalf("failed to get issues: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}
}

func TestMockDeleteIssues(t *testing.T) {
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"},
		{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"},
		{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
	for _, opt := range createIssues {
		_, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
	}
	allIssuesMap := make(map[int64]*Issue)
	queryOpts := []*RepositoryOptions{
		{Owner: "foo", Repo: "meling-labs"},
		{Owner: "foo", Repo: "josie-labs"},
		{Owner: "dat320", Repo: "meling-labs"},
	}
	for _, opt := range queryOpts {
		issues, err := s.GetIssues(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to get issues: %v", err)
		}
		for _, issue := range issues {
			allIssuesMap[issue.ID] = issue
		}
	}
	if len(allIssuesMap) != len(createIssues) {
		t.Fatalf("expected %d issues, got %d", len(createIssues), len(allIssuesMap))
	}

	tests := []struct {
		name                string
		opt                 *RepositoryOptions
		wantRemainingIssues int
		wantErr             bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Repo: "meling-labs"}, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &RepositoryOptions{Owner: "buz", Repo: "meling-labs"}, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &RepositoryOptions{Owner: "foo", Repo: "lamport-labs"}, wantErr: true},

		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, wantRemainingIssues: 4, wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "josie-labs"}, wantRemainingIssues: 2, wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "dat320", Repo: "meling-labs"}, wantRemainingIssues: 0, wantErr: false},
	}
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Repo)
		t.Run(name, func(t *testing.T) {
			if err := s.DeleteIssues(context.Background(), tt.opt); (err != nil) != tt.wantErr {
				t.Errorf("DeleteIssue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			issues, err := s.GetIssues(context.Background(), tt.opt)
			if err != nil {
				t.Fatalf("failed to get issues: %v", err)
			}
			if len(issues) != 0 {
				t.Errorf("expected no issues, got %d", len(issues))
			}
			gotIssues := 0
			for _, opt := range queryOpts {
				if opt.Owner == tt.opt.Owner && opt.Repo == tt.opt.Repo {
					continue
				}
				issues, err := s.GetIssues(context.Background(), opt)
				if err != nil {
					t.Fatalf("failed to get issues: %v", err)
				}
				gotIssues += len(issues)
			}
			if gotIssues != tt.wantRemainingIssues {
				t.Errorf("expected %d remaining issues, got %d", tt.wantRemainingIssues, gotIssues)
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
		"dat320": {
			"meling-labs": {
				{ID: 5, Number: 1, Title: "First 1", Body: "First Body", Repository: "meling-labs"},
				{ID: 6, Number: 2, Title: "Second 2", Body: "Second Body", Repository: "meling-labs"},
			},
		},
	}
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"},
		{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"},
		{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
	for _, opt := range createIssues {
		_, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
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

		{name: "CompleteRequest/OrgNotFound", opt: &IssueOptions{Organization: "buz", Repository: "meling-labs", Title: "First", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &IssueOptions{Organization: "foo", Repository: "lamport-labs", Title: "First", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/MissingTitle", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Body: "xyz"}, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/MissingBody", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "First"}, wantIssue: nil, wantErr: true},

		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "First 1", Body: "First Body", Number: 1}, wantIssue: wantIssues["foo"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "meling-labs", Title: "Second 2", Body: "Second Body", Number: 2}, wantIssue: wantIssues["foo"]["meling-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "First 1", Body: "First Body", Number: 1}, wantIssue: wantIssues["foo"]["josie-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "foo", Repository: "josie-labs", Title: "Second 2", Body: "Second Body", Number: 2}, wantIssue: wantIssues["foo"]["josie-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "dat320", Repository: "meling-labs", Title: "First 1", Body: "First Body", Number: 1}, wantIssue: wantIssues["dat320"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &IssueOptions{Organization: "dat320", Repository: "meling-labs", Title: "Second 2", Body: "Second Body", Number: 2}, wantIssue: wantIssues["dat320"]["meling-labs"][1], wantErr: false},
	}
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
		"dat320": {
			"meling-labs": {
				{ID: 5, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 6, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
		},
	}
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"},
		{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"},
		{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
	for _, opt := range createIssues {
		_, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
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
		{name: "IncompleteRequest", opt: &RepositoryOptions{Repo: "meling-labs"}, wantIssue: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, wantIssue: nil, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &RepositoryOptions{Owner: "buz", Repo: "meling-labs"}, number: 1, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &RepositoryOptions{Owner: "foo", Repo: "lamport-labs"}, number: 1, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/MissingNumber", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, wantIssue: nil, wantErr: true},
		{name: "CompleteRequest/WrongNumber", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, number: 543, wantIssue: nil, wantErr: true},

		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, number: 1, wantIssue: wantIssues["foo"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, number: 2, wantIssue: wantIssues["foo"]["meling-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "josie-labs"}, number: 1, wantIssue: wantIssues["foo"]["josie-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "josie-labs"}, number: 2, wantIssue: wantIssues["foo"]["josie-labs"][1], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "dat320", Repo: "meling-labs"}, number: 1, wantIssue: wantIssues["dat320"]["meling-labs"][0], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "dat320", Repo: "meling-labs"}, number: 2, wantIssue: wantIssues["dat320"]["meling-labs"][1], wantErr: false},
	}
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Repo)
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
		"dat320": {
			"meling-labs": {
				{ID: 5, Number: 1, Title: "First", Body: "xyz", Repository: "meling-labs"},
				{ID: 6, Number: 2, Title: "Second", Body: "abc", Repository: "meling-labs"},
			},
		},
	}
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"},
		{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"},
		{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
	for _, opt := range createIssues {
		_, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
	}

	tests := []struct {
		name       string
		opt        *RepositoryOptions
		wantIssues []*Issue
		wantErr    bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantIssues: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantIssues: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Repo: "meling-labs"}, wantIssues: nil, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &RepositoryOptions{Owner: "buz", Repo: "meling-labs"}, wantIssues: nil, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &RepositoryOptions{Owner: "foo", Repo: "lamport-labs"}, wantIssues: nil, wantErr: true},

		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "meling-labs"}, wantIssues: wantIssues["foo"]["meling-labs"], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "foo", Repo: "josie-labs"}, wantIssues: wantIssues["foo"]["josie-labs"], wantErr: false},
		{name: "CompleteRequest", opt: &RepositoryOptions{Owner: "dat320", Repo: "meling-labs"}, wantIssues: wantIssues["dat320"]["meling-labs"], wantErr: false},
	}
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Repo)
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

func TestMockGetIssues_CheckIssueNumbers(t *testing.T) {
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithRepos(repos...))
	var wantIssueNumbers []int
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "1st", Body: "Body 1"},
		{Organization: "foo", Repository: "meling-labs", Title: "2nd", Body: "Body 2"},
		{Organization: "foo", Repository: "meling-labs", Title: "3rd", Body: "Body 3"},
		{Organization: "foo", Repository: "meling-labs", Title: "4th", Body: "Body 4"},
		{Organization: "foo", Repository: "meling-labs", Title: "5th", Body: "Body 5"},
		{Organization: "foo", Repository: "meling-labs", Title: "6th", Body: "Body 6"},
	}
	for _, opt := range createIssues {
		issue, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
		wantIssueNumbers = append(wantIssueNumbers, issue.Number)
	}

	gotIssueNumbers := make([]int, len(createIssues))
	issues, err := s.GetIssues(context.Background(), &RepositoryOptions{Owner: "foo", Repo: "meling-labs"})
	if err != nil {
		t.Fatalf("failed to get issues: %v", err)
	}
	for i, issue := range issues {
		gotIssueNumbers[i] = issue.Number
	}

	if diff := cmp.Diff(wantIssueNumbers, gotIssueNumbers); diff != "" {
		t.Errorf("GetIssues() mismatch (-want +got):\n%s", diff)
	}
}

func TestMockCreateIssueComment(t *testing.T) {
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"},
		{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"},
		{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
	for _, opt := range createIssues {
		_, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
	}

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

		{name: "CompleteRequest/OrgNotFound", opt: &IssueCommentOptions{Organization: "buz", Repository: "meling-labs", Number: 1, Body: "Hello 1.1"}, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &IssueCommentOptions{Organization: "foo", Repository: "lamport-labs", Number: 1, Body: "Hello 1.1"}, wantErr: true},
		{name: "CompleteRequest/MissingIssue", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Body: "Hello 1.1"}, wantErr: true},
		{name: "CompleteRequest/UnknownIssue", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Number: 654, Body: "Hello 1.1"}, wantErr: true},

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
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", Number: 1, Body: "Hello 1.1"}, wantCommentID: 13, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", Number: 1, Body: "Hello 1.2"}, wantCommentID: 14, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", Number: 1, Body: "Hello 1.3"}, wantCommentID: 15, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", Number: 2, Body: "Hello 2.1"}, wantCommentID: 16, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", Number: 2, Body: "Hello 2.2"}, wantCommentID: 17, wantErr: false},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", Number: 2, Body: "Hello 2.3"}, wantCommentID: 18, wantErr: false},
	}
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

func TestMockUpdateIssueComment(t *testing.T) {
	createIssues := []*IssueOptions{
		{Organization: "foo", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "meling-labs", Title: "Second", Body: "abc"},
		{Organization: "foo", Repository: "josie-labs", Title: "First", Body: "xyz"},
		{Organization: "foo", Repository: "josie-labs", Title: "Second", Body: "abc"},
		{Organization: "dat320", Repository: "meling-labs", Title: "First", Body: "xyz"},
		{Organization: "dat320", Repository: "meling-labs", Title: "Second", Body: "abc"},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgDat320), WithRepos(repos...), WithMockCourses())
	for _, opt := range createIssues {
		_, err := s.CreateIssue(context.Background(), opt)
		if err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
	}
	initialComments := []*IssueCommentOptions{
		{Organization: "foo", Repository: "meling-labs", Number: 1, Body: "Hello 1.1"},
		{Organization: "foo", Repository: "meling-labs", Number: 1, Body: "Hello 1.2"},
		{Organization: "foo", Repository: "meling-labs", Number: 1, Body: "Hello 1.3"},
		{Organization: "foo", Repository: "meling-labs", Number: 2, Body: "Hello 2.1"},
		{Organization: "foo", Repository: "meling-labs", Number: 2, Body: "Hello 2.2"},
		{Organization: "foo", Repository: "meling-labs", Number: 2, Body: "Hello 2.3"},
		{Organization: "foo", Repository: "josie-labs", Number: 1, Body: "Hello 1.1"},
		{Organization: "foo", Repository: "josie-labs", Number: 1, Body: "Hello 1.2"},
		{Organization: "foo", Repository: "josie-labs", Number: 1, Body: "Hello 1.3"},
		{Organization: "foo", Repository: "josie-labs", Number: 2, Body: "Hello 2.1"},
		{Organization: "foo", Repository: "josie-labs", Number: 2, Body: "Hello 2.2"},
		{Organization: "foo", Repository: "josie-labs", Number: 2, Body: "Hello 2.3"},
		{Organization: "dat320", Repository: "meling-labs", Number: 1, Body: "Hello 1.1"},
		{Organization: "dat320", Repository: "meling-labs", Number: 1, Body: "Hello 1.2"},
		{Organization: "dat320", Repository: "meling-labs", Number: 1, Body: "Hello 1.3"},
		{Organization: "dat320", Repository: "meling-labs", Number: 2, Body: "Hello 2.1"},
		{Organization: "dat320", Repository: "meling-labs", Number: 2, Body: "Hello 2.2"},
		{Organization: "dat320", Repository: "meling-labs", Number: 2, Body: "Hello 2.3"},
	}
	for _, comment := range initialComments {
		if _, err := s.CreateIssueComment(context.Background(), comment); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		opt         *IssueCommentOptions
		wantComment github.IssueComment
		wantErr     bool
	}{
		{name: "IncompleteRequest", opt: &IssueCommentOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Body: "Hello"}, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Body: "Hello"}, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs"}, wantErr: true},
		{name: "IncompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Body: "Hello"}, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &IssueCommentOptions{Organization: "buz", Repository: "meling-labs", CommentID: 1, Body: "Hello 1.1"}, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &IssueCommentOptions{Organization: "foo", Repository: "lamport-labs", CommentID: 1, Body: "Hello 1.1"}, wantErr: true},
		{name: "CompleteRequest/MissingCommentID", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", Body: "Hello 1.1"}, wantErr: true},
		{name: "CompleteRequest/UnknownCommentID", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", CommentID: 654, Body: "Hello 1.1"}, wantErr: true},

		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", CommentID: 1, Body: "World 1.1"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(1), Body: github.String("World 1.1")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", CommentID: 2, Body: "World 1.2"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(2), Body: github.String("World 1.2")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", CommentID: 3, Body: "World 1.3"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(3), Body: github.String("World 1.3")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", CommentID: 4, Body: "World 2.1"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(4), Body: github.String("World 2.1")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", CommentID: 5, Body: "World 2.2"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(5), Body: github.String("World 2.2")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "meling-labs", CommentID: 6, Body: "World 2.3"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(6), Body: github.String("World 2.3")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", CommentID: 7, Body: "World 1.1"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(7), Body: github.String("World 1.1")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", CommentID: 8, Body: "World 1.2"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(8), Body: github.String("World 1.2")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", CommentID: 9, Body: "World 1.3"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(9), Body: github.String("World 1.3")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", CommentID: 10, Body: "World 2.1"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(10), Body: github.String("World 2.1")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", CommentID: 11, Body: "World 2.2"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(11), Body: github.String("World 2.2")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "foo", Repository: "josie-labs", CommentID: 12, Body: "World 2.3"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(12), Body: github.String("World 2.3")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", CommentID: 13, Body: "World 1.1"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(13), Body: github.String("World 1.1")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", CommentID: 14, Body: "World 1.2"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(14), Body: github.String("World 1.2")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", CommentID: 15, Body: "World 1.3"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(15), Body: github.String("World 1.3")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", CommentID: 16, Body: "World 2.1"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(16), Body: github.String("World 2.1")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", CommentID: 17, Body: "World 2.2"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(17), Body: github.String("World 2.2")}},
		{name: "CompleteRequest", opt: &IssueCommentOptions{Organization: "dat320", Repository: "meling-labs", CommentID: 18, Body: "World 2.3"}, wantErr: false, wantComment: github.IssueComment{ID: github.Int64(18), Body: github.String("World 2.3")}},
	}
	for _, tt := range tests {
		name := qtest.Name(
			tt.name,
			[]string{"Organization", "Repository", "CommentID", "Body"},
			tt.opt.Organization, tt.opt.Repository, tt.opt.CommentID, tt.opt.Body,
		)
		t.Run(name, func(t *testing.T) {
			err := s.UpdateIssueComment(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateIssueComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			// verify the state of the issue comment
			for _, comment := range s.comments[tt.opt.Organization][tt.opt.Repository][int64(tt.opt.Number)] {
				if *comment.ID == tt.opt.CommentID {
					if diff := cmp.Diff(tt.wantComment, comment); diff != "" {
						t.Errorf("UpdateIssueComment() mismatch (-want +got):\n%s", diff)
					}
					return
				}
			}
		})
	}
}

func TestMockRequestReviewers(t *testing.T) {
	tests := []struct {
		name          string
		opt           *RequestReviewersOptions
		wantErr       bool
		wantReviewers []string
	}{
		{name: "IncompleteRequest", opt: &RequestReviewersOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &RequestReviewersOptions{Organization: "foo"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RequestReviewersOptions{Repository: "meling-labs"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RequestReviewersOptions{Organization: "foo", Repository: "meling-labs"}, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &RequestReviewersOptions{Organization: "buz", Repository: "meling-labs", Number: 1, Reviewers: []string{"meling", "leslie"}}, wantErr: true},
		{name: "CompleteRequest/RepoNotFound", opt: &RequestReviewersOptions{Organization: "foo", Repository: "lamport-labs", Number: 1, Reviewers: []string{"meling", "leslie"}}, wantErr: true},
		{name: "CompleteRequest/MissingNumber", opt: &RequestReviewersOptions{Organization: "foo", Repository: "meling-labs", Reviewers: []string{"meling", "leslie"}}, wantErr: true},
		{name: "CompleteRequest/WrongNumber", opt: &RequestReviewersOptions{Organization: "foo", Repository: "meling-labs", Number: 543, Reviewers: []string{"meling", "leslie"}}, wantErr: true},
		{name: "CompleteRequest/MissingReviewers", opt: &RequestReviewersOptions{Organization: "foo", Repository: "meling-labs", Number: 1}, wantErr: true},

		{name: "CompleteRequest", opt: &RequestReviewersOptions{Organization: "foo", Repository: "meling-labs", Number: 1, Reviewers: []string{"meling", "leslie"}}, wantErr: false, wantReviewers: []string{"meling", "leslie"}},
		{name: "CompleteRequest", opt: &RequestReviewersOptions{Organization: "foo", Repository: "meling-labs", Number: 2, Reviewers: []string{"lamport", "jostein"}}, wantErr: false, wantReviewers: []string{"lamport", "jostein"}},
		{name: "CompleteRequest", opt: &RequestReviewersOptions{Organization: "foo", Repository: "josie-labs", Number: 1, Reviewers: []string{"meling", "leslie"}}, wantErr: false, wantReviewers: []string{"meling", "leslie"}},
		{name: "CompleteRequest", opt: &RequestReviewersOptions{Organization: "foo", Repository: "josie-labs", Number: 2, Reviewers: []string{"meling", "jostein"}}, wantErr: false, wantReviewers: []string{"meling", "jostein"}},
	}

	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMockCourses(), WithReviewers(reviewers))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "Repository", "Number"}, tt.opt.Organization, tt.opt.Repository, tt.opt.Number)
		t.Run(name, func(t *testing.T) {
			if err := s.RequestReviewers(context.Background(), tt.opt); (err != nil) != tt.wantErr {
				t.Errorf("RequestReviewers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.wantReviewers, s.reviewers[tt.opt.Organization][tt.opt.Repository][tt.opt.Number].Reviewers); diff != "" {
				t.Errorf("RequestReviewers() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
