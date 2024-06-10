package scm_test

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func repoName(studName string) *string {
	return github.String(qf.StudentRepoName(studName))
}

var (
	mockIssues = []*scm.Issue{
		{
			ID:         1,
			Number:     1,
			Title:      "Test issue",
			Body:       "This is a test issue.",
			Repository: qf.StudentRepoName("test"),
			Assignee:   user,
		},
		{
			ID:         2,
			Number:     2,
			Title:      "Task 1",
			Body:       "Finish Task 1",
			Repository: qf.StudentRepoName("test"),
			Assignee:   "",
		},
		{
			ID:         3,
			Number:     3,
			Title:      "Task 1",
			Body:       "Finish Task 1",
			Repository: qf.StudentRepoName(user),
			Assignee:   "",
		},
	}
	ghOrg = github.Organization{ID: github.Int64(987), Login: github.String(qtest.MockOrg)}
	repos = []github.Repository{
		{Organization: &ghOrg, Name: repoName("test")},
		{Organization: &ghOrg, Name: repoName(user)},
	}
	mockRepos = []*scm.Repository{
		{
			ID:    1,
			OrgID: 1,
			Owner: qtest.MockOrg,
			Path:  qf.StudentRepoName("test"),
		},
		{
			ID:    2,
			OrgID: 1,
			Owner: qtest.MockOrg,
			Path:  qf.StudentRepoName(user),
		},
	}
)

func TestMockSCMWithCourse(t *testing.T) {
	s := scm.NewMockSCMClientWithCourse()
	wantRepos := map[uint64]*scm.Repository{
		1: {
			ID:    1,
			Path:  "info",
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
		2: {
			ID:    2,
			Path:  "assignments",
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
		3: {
			ID:    3,
			Path:  "tests",
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
		4: {
			ID:    4,
			Path:  qf.StudentRepoName("user"),
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
	}
	if diff := cmp.Diff(wantRepos, s.Repositories); diff != "" {
		t.Errorf("mismatch repos (-want +got):\n%s", diff)
	}
}

func TestMockClone(t *testing.T) {
	dstDir := t.TempDir()
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	tests := []struct {
		name     string
		opt      *scm.CloneOptions
		wantPath string
		wantErr  bool
	}{
		{
			name: "student repository",
			opt: &scm.CloneOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("user"),
				DestDir:      dstDir,
			},
			wantPath: filepath.Join(dstDir, "user-labs"),
			wantErr:  false,
		},
		{
			name: "assignments repository",
			opt: &scm.CloneOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.AssignmentsRepo,
				DestDir:      dstDir,
			},
			wantPath: filepath.Join(dstDir, "assignments"),
			wantErr:  false,
		},
		{
			name: "tests repository",
			opt: &scm.CloneOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.TestsRepo,
				DestDir:      dstDir,
			},
			wantPath: filepath.Join(dstDir, "tests"),
			wantErr:  false,
		},
		{
			name: "missing organization info",
			opt: &scm.CloneOptions{
				Repository: qf.StudentRepoName(user),
				DestDir:    dstDir,
			},
			wantPath: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		path, err := s.Clone(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
		if path != tt.wantPath {
			t.Errorf("%s: expected path '%s', got '%s'", tt.name, tt.wantPath, path)
		}
	}
}

const user = "test_user"

func TestMockOrganizations(t *testing.T) {
	s := scm.NewMockedGithubSCMClient(qtest.Logger(t), scm.WithMockCourses())
	ctx := context.Background()
	for _, course := range qtest.MockCourses {
		if _, err := s.GetOrganization(ctx, &scm.OrganizationOptions{ID: course.ScmOrganizationID}); err != nil {
			t.Error(err)
		}
		if _, err := s.GetOrganization(ctx, &scm.OrganizationOptions{Name: course.ScmOrganizationName}); err != nil {
			t.Error(err)
		}
	}

	invalidOrgs := []struct {
		name       string
		id         uint64
		username   string
		permission string
		err        string
	}{
		{id: 0, name: "", username: "", permission: "", err: "invalid argument"},
		{id: 1234, name: "test_missing_org", username: user, permission: "read", err: "organization not found"},
	}

	for _, org := range invalidOrgs {
		if _, err := s.GetOrganization(ctx, &scm.OrganizationOptions{ID: org.id, Name: org.name}); err == nil {
			t.Errorf("expected error: %s", org.err)
		}
	}
}

func TestMockCreateIssue(t *testing.T) {
	s := scm.NewMockedGithubSCMClient(qtest.Logger(t), scm.WithRepos(repos...))
	ctx := context.Background()
	issue := mockIssues[0]

	tests := []struct {
		name      string
		opt       *scm.IssueOptions
		wantIssue *scm.Issue
		wantErr   bool
	}{
		{
			name: "correct options",
			opt: &scm.IssueOptions{
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			wantIssue: issue,
			wantErr:   false,
		},
		{
			name: "incorrect organization",
			opt: &scm.IssueOptions{
				Organization: "another-organization",
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			wantIssue: nil,
			wantErr:   true,
		},
		{
			name: "missing repository",
			opt: &scm.IssueOptions{
				Organization: qtest.MockOrg,
				Title:        issue.Title,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			wantIssue: nil,
			wantErr:   true,
		},
		{
			name: "missing title",
			opt: &scm.IssueOptions{
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			wantIssue: nil,
			wantErr:   true,
		},
		{
			name: "missing body",
			opt: &scm.IssueOptions{
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        issue.Title,
				Assignee:     &issue.Assignee,
			},
			wantIssue: nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.CreateIssue(ctx, tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
			}
			if diff := cmp.Diff(tt.wantIssue, got); diff != "" {
				t.Errorf("%s mismatch issue (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestMockUpdateIssue(t *testing.T) {
	s := scm.NewMockedGithubSCMClient(qtest.Logger(t), scm.WithRepos(repos...))
	ctx := context.Background()
	issue, err := s.CreateIssue(ctx, &scm.IssueOptions{
		Organization: qtest.MockOrg,
		Repository:   mockIssues[0].Repository,
		Title:        mockIssues[0].Title,
		Body:         mockIssues[0].Body,
		Assignee:     &mockIssues[0].Assignee,
		Number:       mockIssues[0].Number,
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		opt       *scm.IssueOptions
		wantIssue *scm.Issue
		wantErr   bool
	}{
		{
			name: "correct issue, no updates",
			opt: &scm.IssueOptions{
				Number:       issue.Number,
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			wantIssue: issue,
			wantErr:   false,
		},
		{
			name: "correct issue, update title and body",
			opt: &scm.IssueOptions{
				Number:       issue.Number,
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        "New Title",
				Body:         "New Body",
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			wantIssue: &scm.Issue{
				ID:         issue.ID,
				Number:     issue.Number,
				Title:      "New Title",
				Body:       "New Body",
				Repository: issue.Repository,
				Status:     issue.Status,
				Assignee:   issue.Assignee,
			},
			wantErr: false,
		},
		{
			name: "incorrect organization",
			opt: &scm.IssueOptions{
				Number:       issue.Number,
				Organization: "some-org",
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			wantIssue: nil,
			wantErr:   true,
		},
		{
			name: "invalid opts",
			opt: &scm.IssueOptions{
				Number:       issue.Number,
				Organization: qtest.MockOrg,
				Title:        issue.Title,
				Body:         issue.Body,
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			wantIssue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		gotIssue, err := s.UpdateIssue(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantIssue, gotIssue); diff != "" {
			t.Errorf("%s mismatch issue (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockGetIssue(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	issue := mockIssues[0]
	s.Repositories = map[uint64]*scm.Repository{
		1: mockRepos[0],
	}
	s.Issues = map[uint64]*scm.Issue{
		1: issue,
	}
	tests := []struct {
		name      string
		opt       *scm.RepositoryOptions
		number    int
		wantIssue *scm.Issue
		wantErr   bool
	}{
		{
			name: "correct issue",
			opt: &scm.RepositoryOptions{
				Path:  issue.Repository,
				Owner: qtest.MockOrg,
			},
			number:    issue.Number,
			wantIssue: issue,
			wantErr:   false,
		},
		{
			name: "incorrect issue number",
			opt: &scm.RepositoryOptions{
				Path:  issue.Repository,
				Owner: qtest.MockOrg,
			},
			number:    13,
			wantIssue: nil,
			wantErr:   true,
		},
		{
			name: "incorrect organization name",
			opt: &scm.RepositoryOptions{
				Path:  issue.Repository,
				Owner: "some-org",
			},
			number:    issue.Number,
			wantIssue: nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		gotIssue, err := s.GetIssue(ctx, tt.opt, tt.number)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantIssue, gotIssue); diff != "" {
			t.Errorf("%s mismatch issue (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockGetIssues(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	s.Repositories = map[uint64]*scm.Repository{
		1: mockRepos[0],
		2: mockRepos[1],
	}

	for _, issue := range mockIssues {
		if _, err := s.CreateIssue(ctx, &scm.IssueOptions{
			Organization: qtest.MockOrg,
			Repository:   issue.Repository,
			Title:        issue.Title,
			Body:         issue.Body,
			Assignee:     &issue.Assignee,
		}); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name       string
		opt        *scm.RepositoryOptions
		wantIssues []*scm.Issue
		wantErr    bool
	}{
		{
			name: "issues for 'test-labs' repo",
			opt: &scm.RepositoryOptions{
				Owner: qtest.MockOrg,
				Path:  mockIssues[0].Repository,
			},
			wantIssues: []*scm.Issue{mockIssues[0], mockIssues[1]},
			wantErr:    false,
		},
		{
			name: "issues for 'user-labs' repo",
			opt: &scm.RepositoryOptions{
				Owner: qtest.MockOrg,
				Path:  mockIssues[2].Repository,
			},
			wantIssues: []*scm.Issue{mockIssues[2]},
			wantErr:    false,
		},
		{
			name: "incorrect repository",
			opt: &scm.RepositoryOptions{
				Owner: qtest.MockOrg,
				Path:  "unknown-labs",
			},
			wantIssues: nil,
			wantErr:    true,
		},
		{
			name: "incorrect organization",
			opt: &scm.RepositoryOptions{
				Owner: "some-org",
				Path:  mockIssues[0].Repository,
			},
			wantIssues: nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		gotIssues, err := s.GetIssues(ctx, tt.opt)
		sort.Slice(gotIssues, func(i, j int) bool {
			return gotIssues[i].ID < gotIssues[j].ID
		})
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
			t.Errorf("%s mismatch issue (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockGetIssues2(t *testing.T) {
	s := scm.NewMockSCMClient()
	s.Repositories = map[uint64]*scm.Repository{
		1: {
			ID:    1,
			OrgID: 1,
			Owner: qtest.MockOrg,
			Path:  qf.StudentRepoName("test"),
		},
	}

	ctx := context.Background()
	opt := &scm.RepositoryOptions{
		Owner: qtest.MockOrg,
		Path:  qf.StudentRepoName("test"),
	}

	var wantIssueIDs []int
	for i := 1; i <= 5; i++ {
		issue, cleanup := createIssue(t, s, opt.Owner, opt.Path)
		defer cleanup()
		wantIssueIDs = append(wantIssueIDs, issue.Number)
	}

	var gotIssueIDs []int
	gotIssues, err := s.GetIssues(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range gotIssues {
		gotIssueIDs = append(gotIssueIDs, issue.Number)
	}

	less := func(a, b int) bool { return a < b }
	if equal := cmp.Equal(wantIssueIDs, gotIssueIDs, cmpopts.SortSlices(less)); !equal {
		t.Errorf("GetIssues() mismatch wantIssueIDs: %v, gotIssueIDs: %v", wantIssueIDs, gotIssueIDs)
	}
}

func TestMockDeleteIssue(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	s.Repositories = map[uint64]*scm.Repository{
		1: mockRepos[0],
		2: mockRepos[1],
	}
	for _, issue := range mockIssues {
		issueOptions := &scm.IssueOptions{
			Organization: qtest.MockOrg,
			Repository:   issue.Repository,
			Title:        issue.Title,
			Body:         issue.Body,
			Assignee:     &issue.Assignee,
		}
		if _, err := s.CreateIssue(ctx, issueOptions); err != nil {
			t.Fatal(err)
		}
		opt := &scm.RepositoryOptions{
			Owner: qtest.MockOrg,
			Path:  issue.Repository,
		}
		if err := s.DeleteIssue(ctx, opt, issue.Number); err != nil {
			t.Error(err)
		}
		if _, err := s.GetIssue(ctx, opt, issue.Number); err == nil {
			t.Error("expected error 'issue not found'")
		}
	}
}

func TestMockCreateGetDeleteIssueSequence(t *testing.T) {
	s := scm.NewMockedGithubSCMClient(qtest.Logger(t), scm.WithMockCourses())
	ctx := context.Background()

	opt := &scm.IssueOptions{
		Organization: qtest.MockOrg,
		Repository:   qf.StudentRepoName("meling"),
		Title:        "Dummy Title",
		Body:         "Dummy body of the issue",
	}
	issue, err := s.CreateIssue(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	repoOpt := &scm.RepositoryOptions{
		Owner: qtest.MockOrg,
		Path:  opt.Repository,
	}
	issues, err := s.GetIssues(ctx, repoOpt)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
		if len(issues) > 1 {
			for _, issue := range issues {
				t.Logf("unexpected issue: %v", issue)
			}
		}
	}

	if err = s.DeleteIssue(ctx, repoOpt, issue.Number); err != nil {
		t.Fatal(err)
	}

	issues, err = s.GetIssues(ctx, repoOpt)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("unexpected issue: %v", issue)
		}
	}
}

func TestMockDeleteIssues(t *testing.T) {
	ctx := context.Background()
	course := qtest.MockCourses[0]
	s := scm.NewMockSCMClient()
	s.Repositories = map[uint64]*scm.Repository{
		1: mockRepos[0],
		2: mockRepos[1],
	}

	tests := []struct {
		name       string
		opt        *scm.RepositoryOptions
		wantIssues map[uint64]*scm.Issue
		wantErr    bool
	}{
		{
			name: "delete all issues for 'user-labs' repo (issue 3)",
			opt: &scm.RepositoryOptions{
				Path:  qf.StudentRepoName(user),
				Owner: course.ScmOrganizationName,
			},
			wantIssues: map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1]},
			wantErr:    false,
		},
		{
			name: "delete all issues for 'test-labs' repo (issues 1 and 2)",
			opt: &scm.RepositoryOptions{
				Path:  "test-labs",
				Owner: course.ScmOrganizationName,
			},
			wantIssues: map[uint64]*scm.Issue{3: mockIssues[2]},
			wantErr:    false,
		},
		{
			name: "missing repository, nothing deleted",
			opt: &scm.RepositoryOptions{
				Owner: course.ScmOrganizationName,
				Path:  "some-labs",
			},
			wantIssues: map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1], 3: mockIssues[2]},
			wantErr:    true,
		},
		{
			name: "incorrect organization name",
			opt: &scm.RepositoryOptions{
				Owner: "organization",
				Path:  "test-labs",
			},
			wantIssues: map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1], 3: mockIssues[2]},
			wantErr:    true,
		},
		{
			name:       "invalid opt",
			opt:        &scm.RepositoryOptions{},
			wantIssues: map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1], 3: mockIssues[2]},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		issues := make(map[uint64]*scm.Issue)
		for _, issue := range mockIssues {
			issues[issue.ID] = issue
		}
		s.Issues = issues
		if err := s.DeleteIssues(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantIssues, s.Issues); diff != "" {
			t.Errorf("%s mismatch issues (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockCreateIssueComment(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	s.Repositories = map[uint64]*scm.Repository{
		1: mockRepos[0],
	}
	s.Issues = map[uint64]*scm.Issue{
		1: mockIssues[0],
		2: mockIssues[2],
	}

	tests := []struct {
		name       string
		opt        *scm.IssueCommentOptions
		wantNumber int64
		wantErr    bool
	}{
		{
			name: "comment 1 for issue 1",
			opt: &scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       1,
			},
			wantNumber: 1,
			wantErr:    false,
		},
		{
			name: "comment 2 for issue 1",
			opt: &scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       1,
			},
			wantNumber: 2,
			wantErr:    false,
		},
		{
			name: "comment 1 for issue 2",
			opt: &scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       2,
			},
			wantNumber: 3,
			wantErr:    false,
		},
		{
			name: "comment 2 for issue 2",
			opt: &scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       2,
			},
			wantNumber: 4,
			wantErr:    false,
		},
		{
			name: "invalid opts, missing organization",
			opt: &scm.IssueCommentOptions{
				Repository: qf.StudentRepoName("test"),
				Body:       "Comment",
				Number:     1,
			},
			wantNumber: 0,
			wantErr:    true,
		},
		{
			name: "invalid opts, missing repository",
			opt: &scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Body:         "Comment",
				Number:       1,
			},
			wantNumber: 0,
			wantErr:    true,
		},
		{
			name: "invalid opts, missing comment body",
			opt: &scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName(user),
				Number:       1,
			},
			wantNumber: 0,
			wantErr:    true,
		},
		{
			name: "incorrect organization name",
			opt: &scm.IssueCommentOptions{
				Organization: "organization",
				Repository:   qf.StudentRepoName(user),
				Body:         "Comment",
				Number:       1,
			},
			wantNumber: 0,
			wantErr:    true,
		},
		{
			name: "incorrect issue number",
			opt: &scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName(user),
				Body:         "Comment",
				Number:       5,
			},
			wantNumber: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		commentNumber, err := s.CreateIssueComment(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantNumber, commentNumber); diff != "" {
			t.Errorf("%s mismatch comment number (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockUpdateIssueComment(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	s.Repositories = map[uint64]*scm.Repository{
		1: mockRepos[0],
	}
	s.Issues = map[uint64]*scm.Issue{
		1: mockIssues[0],
	}
	s.IssueComments = map[uint64]string{
		1: "Not updated",
		2: "Not updated",
	}
	tests := []struct {
		name        string
		issueNumber int
		commentID   int64
		wantErr     bool
	}{
		{
			name:        "update issue 1",
			issueNumber: 1,
			commentID:   1,
			wantErr:     false,
		},
		{
			name:        "update issue 2",
			issueNumber: 1,
			commentID:   2,
			wantErr:     false,
		},
		{
			name:        "incorrect issue number",
			issueNumber: 5,
			commentID:   1,
			wantErr:     true,
		},
		{
			name:        "incorrect issue comment ID",
			issueNumber: 1,
			commentID:   4,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		if err := s.UpdateIssueComment(ctx, &scm.IssueCommentOptions{
			Organization: qtest.MockOrg,
			Repository:   mockRepos[0].Path,
			Number:       tt.issueNumber,
			CommentID:    tt.commentID,
			Body:         "Updated",
		}); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if !tt.wantErr {
			comment, ok := s.IssueComments[uint64(tt.commentID)]
			if !ok {
				t.Fatalf("%s: comment not found", tt.name)
			}
			if !ok || comment != "Updated" {
				t.Errorf("%s: expected comment body 'Updated', got '%s'", tt.name, comment)
			}
		}
	}
}

func TestMockUpdateIssueComment2(t *testing.T) {
	s := scm.NewMockSCMClient()
	repo := &scm.Repository{
		ID:    1,
		OrgID: 1,
		Owner: qtest.MockOrg,
		Path:  qf.StudentRepoName("user"),
	}
	s.Repositories = map[uint64]*scm.Repository{
		1: repo,
	}

	body := "Issue Comment"
	opt := &scm.IssueCommentOptions{
		Organization: repo.Owner,
		Repository:   repo.Path,
		Body:         body,
	}

	issue, cleanup := createIssue(t, s, opt.Organization, opt.Repository)
	defer cleanup()

	opt.Number = issue.Number
	// The created comment will be deleted when the parent issue is deleted.
	commentID, err := s.CreateIssueComment(context.Background(), opt)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: We do not currently return the updated comment, so we cannot verify its content.
	opt.Body = "Updated Issue Comment"
	opt.CommentID = commentID
	if err := s.UpdateIssueComment(context.Background(), opt); err != nil {
		t.Fatal(err)
	}
}

func TestMockCreateCourse(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	wantRepos := []string{qf.InfoRepo, qf.AssignmentsRepo, qf.TestsRepo, qf.StudentRepoName(user)}
	found := func(wantRepo string, repos []*scm.Repository) bool {
		for _, repo := range repos {
			if repo.Path == wantRepo {
				return true
			}
		}
		return false
	}

	tests := []struct {
		name      string
		opt       *scm.CourseOptions
		wantRepos int
		wantErr   bool
	}{
		{
			name: "invalid opt, missing org ID",
			opt: &scm.CourseOptions{
				CourseCreator: user,
			},
			wantRepos: 0,
			wantErr:   true,
		},
		{
			name: "invalid opt, missing user",
			opt: &scm.CourseOptions{
				OrganizationID: 1,
			},
			wantRepos: 0,
			wantErr:   true,
		},
		{
			name: "incorrect organization ID",
			opt: &scm.CourseOptions{
				OrganizationID: 123,
				CourseCreator:  user,
			},
			wantRepos: 0,
			wantErr:   true,
		},
		{
			name: "correct arguments",
			opt: &scm.CourseOptions{
				OrganizationID: 1,
				CourseCreator:  user,
			},
			wantRepos: len(wantRepos),
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		repos, err := s.CreateCourse(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Error(err)
		}
		if len(s.Repositories) != tt.wantRepos {
			t.Errorf("expected repositories: %d, got: %d", tt.wantRepos, len(repos))
		}
		if !tt.wantErr {
			for _, r := range wantRepos {
				if !found(r, repos) {
					t.Errorf("expected repository %s to be found", r)
				}
			}
		}
	}
}

func TestMockUpdateEnrollment(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	tests := []struct {
		name      string
		opt       *scm.UpdateEnrollmentOptions
		wantRepos map[uint64]*scm.Repository
		wantErr   bool
	}{
		{
			name: "invalid opt, missing course",
			opt: &scm.UpdateEnrollmentOptions{
				User:   user,
				Status: qf.Enrollment_STUDENT,
			},
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "invalid opt, missing user name",
			opt: &scm.UpdateEnrollmentOptions{
				Organization: qtest.MockOrg,
				Status:       qf.Enrollment_STUDENT,
			},
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "enroll teacher, no new repos",
			opt: &scm.UpdateEnrollmentOptions{
				Organization: qtest.MockOrg,
				User:         user,
				Status:       qf.Enrollment_TEACHER,
			},
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   false,
		},
		{
			name: "enroll student, new repo added",
			opt: &scm.UpdateEnrollmentOptions{
				Organization: qtest.MockOrg,
				User:         user,
				Status:       qf.Enrollment_STUDENT,
			},
			wantRepos: map[uint64]*scm.Repository{
				1: {
					ID:    1,
					Path:  qf.StudentRepoName(user),
					Owner: qtest.MockOrg,
					OrgID: 1,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		if _, err := s.UpdateEnrollment(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(s.Repositories, tt.wantRepos); diff != "" {
			t.Errorf("%s: mismatch repos (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockRejectEnrollment(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	repo := &scm.Repository{
		ID:    1,
		Owner: qtest.MockOrg,
		Path:  "test-group",
		OrgID: 1,
	}
	s.Repositories = map[uint64]*scm.Repository{
		1: repo,
	}
	tests := []struct {
		name      string
		opt       *scm.RejectEnrollmentOptions
		wantRepos map[uint64]*scm.Repository
		wantErr   bool
	}{
		{
			name: "invalid options, missing repo ID",
			opt: &scm.RejectEnrollmentOptions{
				OrganizationID: 1,
				User:           user,
			},
			wantRepos: map[uint64]*scm.Repository{1: repo},
			wantErr:   true,
		},
		{
			name: "invalid options, missing organization ID",
			opt: &scm.RejectEnrollmentOptions{
				RepositoryID: 1,
				User:         user,
			},
			wantRepos: map[uint64]*scm.Repository{1: repo},
			wantErr:   true,
		},
		{
			name: "invalid options, missing user login",
			opt: &scm.RejectEnrollmentOptions{
				RepositoryID:   1,
				OrganizationID: 1,
			},
			wantRepos: map[uint64]*scm.Repository{1: repo},
			wantErr:   true,
		},
		{
			name: "valid options, must remove repository",
			opt: &scm.RejectEnrollmentOptions{
				OrganizationID: 1,
				RepositoryID:   1,
				User:           user,
			},
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		if err := s.RejectEnrollment(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(s.Repositories, tt.wantRepos); diff != "" {
			t.Errorf("%s: mismatch repos (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockCreateGroup(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()

	paths := []string{"a_group", "another_group", "best_group"}

	groupRepos := []*scm.Repository{
		{
			ID:    1,
			Path:  paths[0],
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
		{
			ID:    2,
			Path:  paths[1],
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
	}
	tests := []struct {
		name      string
		opt       *scm.GroupOptions
		wantRepo  *scm.Repository
		wantRepos map[uint64]*scm.Repository
		wantErr   bool
	}{
		{
			name: "invalid opts, missing organization",
			opt: &scm.GroupOptions{
				GroupName: "test-group",
			},
			wantRepo:  nil,
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "invalid opts, missing group name",
			opt: &scm.GroupOptions{
				Organization: qtest.MockOrg,
			},
			wantRepo:  nil,
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "organization does not exist",
			opt: &scm.GroupOptions{
				Organization: "some-org",
				GroupName:    "group",
			},
			wantRepo:  nil,
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "add a new group",
			opt: &scm.GroupOptions{
				Organization: qtest.MockOrg,
				GroupName:    paths[0],
				Users:        []string{user},
			},
			wantRepo:  groupRepos[0],
			wantRepos: map[uint64]*scm.Repository{1: groupRepos[0]},
			wantErr:   false,
		},
		{
			name: "add another group",
			opt: &scm.GroupOptions{
				Organization: qtest.MockOrg,
				GroupName:    paths[1],
				Users:        []string{user},
			},
			wantRepo:  groupRepos[1],
			wantRepos: map[uint64]*scm.Repository{1: groupRepos[0], 2: groupRepos[1]},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		repo, err := s.CreateGroup(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantRepo, repo); diff != "" {
			t.Errorf("%s: mismatch repo (-want +got):\n%s", tt.name, diff)
		}
		if diff := cmp.Diff(tt.wantRepos, s.Repositories); diff != "" {
			t.Errorf("%s: mismatch repos (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestMockDeleteGroup(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	repositories := []*scm.Repository{
		{
			ID:    1,
			OrgID: 1,
			Owner: qtest.MockOrg,
			Path:  qf.StudentRepoName(user),
		},
		{
			ID:    2,
			OrgID: 1,
			Owner: qtest.MockOrg,
			Path:  "a_group",
		},
	}
	s.Repositories = map[uint64]*scm.Repository{
		1: repositories[0],
		2: repositories[1],
	}

	tests := []struct {
		name      string
		opt       *scm.RepositoryOptions
		wantRepos map[uint64]*scm.Repository
		wantErr   bool
	}{
		{
			name: "invalid opt, missing repo ID",
			opt:  &scm.RepositoryOptions{
				// empty
			},
			wantRepos: map[uint64]*scm.Repository{
				1: repositories[0],
				2: repositories[1],
			},
			wantErr: true,
		},
		{
			name: "correct opt, delete group repo with ID 2",
			opt: &scm.RepositoryOptions{
				ID: 2,
			},
			wantRepos: map[uint64]*scm.Repository{
				1: repositories[0],
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if err := s.DeleteGroup(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Error(err)
		}
		if diff := cmp.Diff(tt.wantRepos, s.Repositories); diff != "" {
			t.Errorf("%s: mismatch repos (-want +got):\n%s", tt.name, diff)
		}
	}
}
