package scm_test

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

var (
	mockIssues = []*scm.Issue{
		{
			ID:         1,
			Title:      "Test issue",
			Body:       "This is a test issue.",
			Repository: qf.StudentRepoName("test"),
			Number:     1,
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
	wantTeams := map[uint64]*scm.Team{
		1: {
			ID:           1,
			Name:         scm.TeachersTeam,
			Organization: qtest.MockOrg,
		},
	}
	if diff := cmp.Diff(wantTeams, s.Teams); diff != "" {
		t.Errorf("mismatch teams (-want +got):\n%s", diff)
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
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	for _, course := range qtest.MockCourses {
		if _, err := s.GetOrganization(ctx, &scm.OrganizationOptions{ID: course.OrganizationID}); err != nil {
			t.Error(err)
		}
		if _, err := s.GetOrganization(ctx, &scm.OrganizationOptions{Name: course.OrganizationName}); err != nil {
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
		{id: 123, name: "test_missing_org", username: user, permission: "read", err: "organization not found"},
	}

	for _, org := range invalidOrgs {
		if _, err := s.GetOrganization(ctx, &scm.OrganizationOptions{ID: org.id, Name: org.name}); err == nil {
			t.Errorf("expected error: %s", org.err)
		}
	}
}

var mockTeams = []*scm.Team{
	{
		ID:           1,
		Organization: qtest.MockOrg,
		Name:         "a_team",
	},
	{
		ID:           2,
		Organization: qtest.MockOrg,
		Name:         "another_team",
	},
	{
		ID:           3,
		Organization: qtest.MockOrg,
		Name:         "best_team",
	},
}

func TestUpdateMockTeamMembers(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	course := qtest.MockCourses[0]
	team := &scm.Team{
		ID:           1,
		Name:         "test-team",
		Organization: course.OrganizationName,
	}
	s.Teams[1] = team
	tests := []struct {
		name    string
		opt     *scm.UpdateTeamOptions
		wantErr bool
	}{
		{
			name: "valid team and opts",
			opt: &scm.UpdateTeamOptions{
				OrganizationID: course.OrganizationID,
				TeamID:         team.ID,
			},
			wantErr: false,
		},
		{
			name: "missing team ID",
			opt: &scm.UpdateTeamOptions{
				OrganizationID: course.OrganizationID,
			},
			wantErr: true,
		},
		{
			name: "valid team, missing org ID",
			opt: &scm.UpdateTeamOptions{
				TeamID: team.ID,
			},
			wantErr: true,
		},
		{
			name: "invalid team",
			opt: &scm.UpdateTeamOptions{
				TeamID:         123,
				OrganizationID: course.OrganizationID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if err := s.UpdateTeamMembers(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
	}
}

func TestMockCreateIssue(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	s.Repositories = map[uint64]*scm.Repository{
		1: mockRepos[0],
	}
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
				Owner: course.OrganizationName,
			},
			wantIssues: map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1]},
			wantErr:    false,
		},
		{
			name: "delete all issues for 'test-labs' repo (issues 1 and 2)",
			opt: &scm.RepositoryOptions{
				Path:  "test-labs",
				Owner: course.OrganizationName,
			},
			wantIssues: map[uint64]*scm.Issue{3: mockIssues[2]},
			wantErr:    false,
		},
		{
			name: "missing repository, nothing deleted",
			opt: &scm.RepositoryOptions{
				Owner: course.OrganizationName,
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
		wantTeams map[uint64]*scm.Team
		wantErr   bool
	}{
		{
			name: "invalid opt, missing org ID",
			opt: &scm.CourseOptions{
				CourseCreator: user,
			},
			wantRepos: 0,
			wantTeams: map[uint64]*scm.Team{},
			wantErr:   true,
		},
		{
			name: "invalid opt, missing user",
			opt: &scm.CourseOptions{
				OrganizationID: 1,
			},
			wantRepos: 0,
			wantTeams: map[uint64]*scm.Team{},
			wantErr:   true,
		},
		{
			name: "incorrect organization ID",
			opt: &scm.CourseOptions{
				OrganizationID: 123,
				CourseCreator:  user,
			},
			wantRepos: 0,
			wantTeams: map[uint64]*scm.Team{},
			wantErr:   true,
		},
		{
			name: "correct arguments",
			opt: &scm.CourseOptions{
				OrganizationID: 1,
				CourseCreator:  user,
			},
			wantRepos: len(wantRepos),
			wantTeams: map[uint64]*scm.Team{
				1: {
					ID:           1,
					Name:         scm.TeachersTeam,
					Organization: qtest.MockOrg,
				},
			},
			wantErr: false,
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
			if diff := cmp.Diff(tt.wantTeams, s.Teams); diff != "" {
				t.Errorf("mismatch teams (-want +got):\n%s", diff)
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
		Path:  "testgrp",
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
	teamRepos := []*scm.Repository{
		{
			ID:    1,
			Path:  mockTeams[0].Name,
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
		{
			ID:    2,
			Path:  mockTeams[1].Name,
			Owner: qtest.MockOrg,
			OrgID: 1,
		},
	}
	tests := []struct {
		name      string
		opt       *scm.TeamOptions
		wantTeam  *scm.Team
		wantRepo  *scm.Repository
		wantTeams map[uint64]*scm.Team
		wantRepos map[uint64]*scm.Repository
		wantErr   bool
	}{
		{
			name: "invalid opts, missing organization",
			opt: &scm.TeamOptions{
				TeamName: "test-team",
			},
			wantTeam:  nil,
			wantRepo:  nil,
			wantTeams: map[uint64]*scm.Team{},
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "invalid opts, missing team name",
			opt: &scm.TeamOptions{
				Organization: qtest.MockOrg,
			},
			wantTeam:  nil,
			wantRepo:  nil,
			wantTeams: map[uint64]*scm.Team{},
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "organization does not exist",
			opt: &scm.TeamOptions{
				Organization: "some-org",
				TeamName:     "team",
			},
			wantTeam:  nil,
			wantRepo:  nil,
			wantTeams: map[uint64]*scm.Team{},
			wantRepos: map[uint64]*scm.Repository{},
			wantErr:   true,
		},
		{
			name: "add a new group",
			opt: &scm.TeamOptions{
				Organization: qtest.MockOrg,
				TeamName:     mockTeams[0].Name,
				Users:        []string{user},
			},
			wantTeam:  mockTeams[0],
			wantRepo:  teamRepos[0],
			wantTeams: map[uint64]*scm.Team{1: mockTeams[0]},
			wantRepos: map[uint64]*scm.Repository{1: teamRepos[0]},
			wantErr:   false,
		},
		{
			name: "add another group",
			opt: &scm.TeamOptions{
				Organization: qtest.MockOrg,
				TeamName:     mockTeams[1].Name,
				Users:        []string{user},
			},
			wantTeam:  mockTeams[1],
			wantRepo:  teamRepos[1],
			wantTeams: map[uint64]*scm.Team{1: mockTeams[0], 2: mockTeams[1]},
			wantRepos: map[uint64]*scm.Repository{1: teamRepos[0], 2: teamRepos[1]},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		repo, team, err := s.CreateGroup(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantRepo, repo); diff != "" {
			t.Errorf("%s: mismatch repo (-want +got):\n%s", tt.name, diff)
		}
		if diff := cmp.Diff(tt.wantTeam, team); diff != "" {
			t.Errorf("%s: mismatch team (-want +got):\n%s", tt.name, diff)
		}
		if diff := cmp.Diff(tt.wantRepos, s.Repositories); diff != "" {
			t.Errorf("%s: mismatch repos (-want +got):\n%s", tt.name, diff)
		}
		if diff := cmp.Diff(tt.wantTeams, s.Teams); diff != "" {
			t.Errorf("%s: mismatch teams (-want +got):\n%s", tt.name, diff)
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
			Path:  mockTeams[0].Name,
		},
	}
	s.Repositories = map[uint64]*scm.Repository{
		1: repositories[0],
		2: repositories[1],
	}
	s.Teams = map[uint64]*scm.Team{
		1: mockTeams[0],
		2: mockTeams[1],
		3: mockTeams[2],
	}

	tests := []struct {
		name      string
		opt       *scm.GroupOptions
		wantRepos map[uint64]*scm.Repository
		wantTeams map[uint64]*scm.Team
		wantErr   bool
	}{
		{
			name: "invalid opt, missing organization",
			opt: &scm.GroupOptions{
				TeamID:       1,
				RepositoryID: 2,
			},
			wantRepos: map[uint64]*scm.Repository{
				1: repositories[0],
				2: repositories[1],
			},
			wantTeams: map[uint64]*scm.Team{
				1: mockTeams[0],
				2: mockTeams[1],
				3: mockTeams[2],
			},
			wantErr: true,
		},
		{
			name: "invalid opt, missing team ID",
			opt: &scm.GroupOptions{
				OrganizationID: 1,
				RepositoryID:   1,
			},
			wantRepos: map[uint64]*scm.Repository{
				1: repositories[0],
				2: repositories[1],
			},
			wantTeams: map[uint64]*scm.Team{
				1: mockTeams[0],
				2: mockTeams[1],
				3: mockTeams[2],
			},
			wantErr: true,
		},
		{
			name: "invalid opt, missing repo ID",
			opt: &scm.GroupOptions{
				OrganizationID: 1,
				TeamID:         1,
			},
			wantRepos: map[uint64]*scm.Repository{
				1: repositories[0],
				2: repositories[1],
			},
			wantTeams: map[uint64]*scm.Team{
				1: mockTeams[0],
				2: mockTeams[1],
				3: mockTeams[2],
			},
			wantErr: true,
		},
		{
			name: "incorrect organization ID",
			opt: &scm.GroupOptions{
				OrganizationID: 1,
			},
			wantRepos: map[uint64]*scm.Repository{
				1: repositories[0],
				2: repositories[1],
			},
			wantTeams: map[uint64]*scm.Team{
				1: mockTeams[0],
				2: mockTeams[1],
				3: mockTeams[2],
			},
			wantErr: true,
		},
		{
			name: "correct opt, delete group repo with ID 2, team ID 1",
			opt: &scm.GroupOptions{
				OrganizationID: 1,
				TeamID:         1,
				RepositoryID:   2,
			},
			wantRepos: map[uint64]*scm.Repository{
				1: repositories[0],
			},
			wantTeams: map[uint64]*scm.Team{
				2: mockTeams[1],
				3: mockTeams[2],
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
		if diff := cmp.Diff(tt.wantTeams, s.Teams); diff != "" {
			t.Errorf("%s: mismatch teams (-want +got):\n%s", tt.name, diff)
		}
	}
}
