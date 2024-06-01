package scm

import (
	"context"
	"net/http"
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

	postReposIssuesByOwnerByRepoHandler := WithRequestMatchHandler(
		postReposIssuesByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")

			issue := MustUnmarshal[github.Issue](r.Body)
			for _, wantIssue := range wantIssues[owner][repo] {
				if wantIssue.Title == *issue.Title && wantIssue.Body == *issue.Body {
					issue.ID = github.Int64(int64(wantIssue.ID))
					issue.Number = github.Int(wantIssue.Number)
					issue.Repository = &github.Repository{
						Owner: &github.User{Login: github.String(owner)},
						Name:  github.String(repo),
					}
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write(MustMarshal(issue))
					return
				}
			}
		}),
	)
	httpClient := NewMockedHTTPClient(
		postReposIssuesByOwnerByRepoHandler,
	)
	s := &GithubSCM{
		logger:      qtest.Logger(t),
		client:      github.NewClient(httpClient),
		providerURL: "github.com",
	}
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "Repository", "Title", "Body", "Number"}, tt.opt.Organization, tt.opt.Repository, tt.opt.Title, tt.opt.Body, tt.opt.Number)
		t.Run(name, func(t *testing.T) {
			issue, err := s.CreateIssue(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateIssue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(issue, tt.wantIssue); diff != "" {
				t.Errorf("CreateIssue() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
