package scm_test

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func TestMockClone(t *testing.T) {
	dstDir := t.TempDir()
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	cloneTests := []struct {
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
	for _, tt := range cloneTests {
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
		if _, err := s.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.OrganizationID}); err != nil {
			t.Error(err)
		}
		if _, err := s.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.OrganizationName}); err != nil {
			t.Error(err)
		}
		if err := s.UpdateOrganization(ctx, &scm.OrganizationOptions{
			Name:              course.OrganizationName,
			DefaultPermission: "read",
		}); err != nil {
			t.Error(err)
		}
		if err := s.UpdateOrgMembership(ctx, &scm.OrgMembershipOptions{
			Organization: course.OrganizationName,
			Username:     user,
		}); err != nil {
			t.Error(err)
		}
		if err := s.RemoveMember(ctx, &scm.OrgMembershipOptions{
			Organization: course.OrganizationName,
			Username:     user,
		}); err != nil {
			t.Error(err)
		}
	}
	if err := s.UpdateOrganization(ctx, &scm.OrganizationOptions{
		Name: qtest.MockCourses[0].OrganizationName,
	}); err == nil {
		t.Error("expected error 'invalid argument'")
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
		if _, err := s.GetOrganization(ctx, &scm.GetOrgOptions{ID: org.id, Name: org.name}); err == nil {
			t.Errorf("expected error: %s", org.err)
		}
		if err := s.UpdateOrganization(ctx, &scm.OrganizationOptions{
			Name:              org.name,
			DefaultPermission: org.permission,
		}); err == nil {
			t.Errorf("expected error: %s", org.err)
		}
		opt := &scm.OrgMembershipOptions{
			Organization: org.name,
			Username:     org.username,
		}
		if err := s.UpdateOrgMembership(ctx, opt); err == nil {
			t.Errorf("expected error: %s", org.err)
		}
		if err := s.RemoveMember(ctx, opt); err == nil {
			t.Errorf("expected error: %s", org.err)
		}
	}
}

func TestMockRepositories(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	course, course2 := qtest.MockCourses[0], qtest.MockCourses[2]
	repos := []*scm.Repository{
		{
			OrgID: course.OrganizationID,
			Owner: course.OrganizationName,
			Path:  "info",
		},
		{
			OrgID: course.OrganizationID,
			Owner: course.OrganizationName,
			Path:  "tests",
		},
		{
			OrgID: course2.OrganizationID,
			Owner: course2.OrganizationName,
			Path:  "assignments",
		},
		{
			OrgID: course2.OrganizationID,
			Owner: course2.OrganizationName,
			Path:  "tests",
		},
	}

	for _, repo := range repos {
		r, err := s.CreateRepository(ctx, &scm.CreateRepositoryOptions{
			Organization: repo.Owner,
			Path:         repo.Path,
			Owner:        repo.Owner,
			Permission:   "read",
		})
		if err != nil {
			t.Error(err)
		}
		repo.ID = r.ID
		if err := s.UpdateRepoAccess(ctx, repo, "", ""); err != nil {
			t.Error(err)
		}
		gotRepo, err := s.GetRepository(ctx, &scm.RepositoryOptions{
			ID:    repo.ID,
			Path:  repo.Path,
			Owner: repo.Owner,
		})
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(repo, gotRepo, cmpopts.IgnoreFields(scm.Repository{}, "HTMLURL")); diff != "" {
			t.Errorf("Expected same repository, got (-sub +want):\n%s", diff)
		}
	}

	wantRepos := []*scm.Repository{repos[0], repos[1]}
	courseRepos, err := s.GetRepositories(ctx, &qf.Organization{ID: course.OrganizationID})
	if err != nil {
		t.Error(err)
	}
	sort.Slice(courseRepos, func(i, j int) bool {
		return courseRepos[i].ID < courseRepos[j].ID
	})
	if diff := cmp.Diff(wantRepos, courseRepos, cmpopts.IgnoreFields(scm.Repository{}, "HTMLURL")); diff != "" {
		t.Errorf("Expected same repositories, got (-sub +want):\n%s", diff)
	}

	if err := s.DeleteRepository(ctx, &scm.RepositoryOptions{ID: 3}); err != nil {
		t.Error(err)
	}
	courseRepos, err = s.GetRepositories(ctx, &qf.Organization{ID: course2.OrganizationID})
	if err != nil {
		t.Error(err)
	}
	if len(courseRepos) > 1 {
		t.Errorf("expected 1 repository, got %d", len(courseRepos))
	}
	if diff := cmp.Diff(repos[3], courseRepos[0], cmpopts.IgnoreFields(scm.Repository{}, "HTMLURL")); diff != "" {
		t.Errorf("Expected same repository, got (-sub +want):\n%s", diff)
	}
}

func TestMockTeams(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	course := qtest.MockCourses[0]
	teams := []*scm.Team{
		{
			Organization: course.OrganizationName,
			Name:         "a_team",
		},
		{
			Organization: course.OrganizationName,
			Name:         "another_team",
		},
		{
			Organization: course.OrganizationName,
			Name:         "best_team",
		},
	}
	for _, team := range teams {
		wantTeam, err := s.CreateTeam(ctx, &scm.NewTeamOptions{
			Organization: team.Organization,
			TeamName:     team.Name,
		})
		if err != nil {
			t.Error(err)
		}
		team.ID = wantTeam.ID
		opts := []*scm.TeamOptions{
			{
				TeamID:         team.ID,
				OrganizationID: course.OrganizationID,
			},
			{
				TeamName:     team.Name,
				Organization: team.Organization,
			},
		}
		for _, opt := range opts {
			gotTeam, err := s.GetTeam(ctx, opt)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(wantTeam, gotTeam); diff != "" {
				t.Errorf("Expected same team, got (-sub +want):\n%s", diff)
			}
		}
	}
	if err := s.DeleteTeam(ctx, &scm.TeamOptions{
		TeamID:         2,
		OrganizationID: course.OrganizationID,
	}); err != nil {
		t.Error(err)
	}
	wantTeams := []*scm.Team{teams[0], teams[2]}
	courseTeams, err := s.GetTeams(ctx, &qf.Organization{
		ID:   course.OrganizationID,
		Name: course.OrganizationName,
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(courseTeams, func(i, j int) bool {
		return courseTeams[i].ID < courseTeams[j].ID
	})
	if diff := cmp.Diff(wantTeams, courseTeams); diff != "" {
		t.Errorf("Expected same teams, got (-sub +want):\n%s", diff)
	}
}

func TestAddRemoveMockTeamMembers(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	course := qtest.MockCourses[0]
	team, err := s.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: course.OrganizationName,
		TeamName:     "test_team",
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		opt     *scm.TeamMembershipOptions
		wantErr bool
	}{
		{
			name: "valid team, with team and organization ID",
			opt: &scm.TeamMembershipOptions{
				TeamID:         team.ID,
				OrganizationID: course.OrganizationID,
				Username:       user,
			},
			wantErr: false,
		},
		{
			name: "valid team, with team and organization name",
			opt: &scm.TeamMembershipOptions{
				TeamName:     team.Name,
				Organization: team.Organization,
				Username:     user,
			},
			wantErr: false,
		},
		{
			name: "valid team, missing team info",
			opt: &scm.TeamMembershipOptions{
				Organization:   course.OrganizationName,
				OrganizationID: course.OrganizationID,
				Username:       user,
			},
			wantErr: true,
		},
		{
			name: "valid team, missing organization info",
			opt: &scm.TeamMembershipOptions{
				TeamID:   team.ID,
				TeamName: team.Name,
				Username: user,
			},
			wantErr: true,
		},
		{
			name: "valid team, missing username",
			opt: &scm.TeamMembershipOptions{
				TeamID:         team.ID,
				TeamName:       team.Name,
				Organization:   course.OrganizationName,
				OrganizationID: course.OrganizationID,
			},
			wantErr: true,
		},
		{
			name: "invalid team",
			opt: &scm.TeamMembershipOptions{
				TeamID:         123,
				TeamName:       "not-a-team",
				Organization:   course.OrganizationName,
				OrganizationID: course.OrganizationID,
				Username:       user,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if err := s.AddTeamMember(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
		if err := s.RemoveTeamMember(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
	}
}

func TestUpdateMockTeamMembers(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	course := qtest.MockCourses[0]
	team, err := s.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: course.OrganizationName,
		TeamName:     "test_team",
	})
	if err != nil {
		t.Fatal(err)
	}
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

func TestMockTeamRepo(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	course := qtest.MockCourses[0]
	team, err := s.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: course.OrganizationName,
		TeamName:     "test_team",
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opt     *scm.AddTeamRepoOptions
		wantErr bool
	}{
		{
			name: "correct team",
			opt: &scm.AddTeamRepoOptions{
				OrganizationID: course.OrganizationID,
				TeamID:         team.ID,
				Repo:           team.Name,
				Owner:          course.OrganizationName,
				Permission:     "push",
			},
			wantErr: false,
		},
		{
			name: "correct team, incorrect organization",
			opt: &scm.AddTeamRepoOptions{
				OrganizationID: 123,
				TeamID:         team.ID,
				Repo:           team.Name,
				Owner:          "not_found",
				Permission:     "push",
			},
			wantErr: false,
		},
		{
			name: "missing organization ID",
			opt: &scm.AddTeamRepoOptions{
				TeamID:     team.ID,
				Repo:       team.Name,
				Owner:      course.OrganizationName,
				Permission: "push",
			},
			wantErr: true,
		},
		{
			name: "missing team ID",
			opt: &scm.AddTeamRepoOptions{
				OrganizationID: course.OrganizationID,
				Repo:           team.Name,
				Owner:          course.OrganizationName,
				Permission:     "push",
			},
			wantErr: true,
		},
		{
			name: "missing repository name",
			opt: &scm.AddTeamRepoOptions{
				OrganizationID: course.OrganizationID,
				TeamID:         team.ID,
				Owner:          course.OrganizationName,
				Permission:     "push",
			},
			wantErr: true,
		},
		{
			name: "missing repository owner",
			opt: &scm.AddTeamRepoOptions{
				OrganizationID: course.OrganizationID,
				TeamID:         team.ID,
				Repo:           team.Name,
				Permission:     "push",
			},
			wantErr: true,
		},
		{
			name: "missing permissions",
			opt: &scm.AddTeamRepoOptions{
				OrganizationID: course.OrganizationID,
				TeamID:         team.ID,
				Repo:           team.Name,
				Owner:          course.OrganizationName,
			},
			wantErr: true,
		},
		{
			name: "team does no exist",
			opt: &scm.AddTeamRepoOptions{
				OrganizationID: course.OrganizationID,
				TeamID:         15,
				Repo:           "not_found",
				Owner:          course.OrganizationName,
				Permission:     "push",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if err := s.AddTeamRepo(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v", tt.name, tt.wantErr, err)
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
			"correct options",
			&scm.IssueOptions{
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			issue,
			false,
		},
		{
			"incorrect organization",
			&scm.IssueOptions{
				Organization: "another-organization",
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			nil,
			true,
		},
		{
			"missing repository",
			&scm.IssueOptions{
				Organization: qtest.MockOrg,
				Title:        issue.Title,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			nil,
			true,
		},
		{
			"missing title",
			&scm.IssueOptions{
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Body:         issue.Body,
				Assignee:     &issue.Assignee,
			},
			nil,
			true,
		},
		{
			"missing body",
			&scm.IssueOptions{
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        issue.Title,
				Assignee:     &issue.Assignee,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.CreateIssue(ctx, tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
			}
			if diff := cmp.Diff(tt.wantIssue, got); diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", tt.name, diff)
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
			"correct issue, no updates",
			&scm.IssueOptions{
				Number:       issue.Number,
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			issue,
			false,
		},
		{
			"correct issue, update title and body",
			&scm.IssueOptions{
				Number:       issue.Number,
				Organization: qtest.MockOrg,
				Repository:   issue.Repository,
				Title:        "New Title",
				Body:         "New Body",
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			&scm.Issue{
				ID:         issue.ID,
				Number:     issue.Number,
				Title:      "New Title",
				Body:       "New Body",
				Repository: issue.Repository,
				Status:     issue.Status,
				Assignee:   issue.Assignee,
			},
			false,
		},
		{
			"incorrect organization",
			&scm.IssueOptions{
				Number:       issue.Number,
				Organization: "some-org",
				Repository:   issue.Repository,
				Title:        issue.Title,
				Body:         issue.Body,
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			nil,
			true,
		},
		{
			"invalid opts",
			&scm.IssueOptions{
				Number:       issue.Number,
				Organization: qtest.MockOrg,
				Title:        issue.Title,
				Body:         issue.Body,
				State:        issue.Status,
				Assignee:     &issue.Assignee,
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		gotIssue, err := s.UpdateIssue(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantIssue, gotIssue); diff != "" {
			t.Errorf("%s mismatch (-want +got):\n%s", tt.name, diff)
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
			"correct issue",
			&scm.RepositoryOptions{
				Path:  issue.Repository,
				Owner: qtest.MockOrg,
			},
			issue.Number,
			issue,
			false,
		},
		{
			"incorrect issue number",
			&scm.RepositoryOptions{
				Path:  issue.Repository,
				Owner: qtest.MockOrg,
			},
			13,
			nil,
			true,
		},
		{
			"incorrect organization name",
			&scm.RepositoryOptions{
				Path:  issue.Repository,
				Owner: "some-org",
			},
			issue.Number,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		gotIssue, err := s.GetIssue(ctx, tt.opt, tt.number)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantIssue, gotIssue); diff != "" {
			t.Errorf("%s mismatch (-want +got):\n%s", tt.name, diff)
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
			"issues for 'test-labs' repo",
			&scm.RepositoryOptions{
				Owner: qtest.MockOrg,
				Path:  mockIssues[0].Repository,
			},
			[]*scm.Issue{mockIssues[0], mockIssues[1]},
			false,
		},
		{
			"issues for 'user-labs' repo",
			&scm.RepositoryOptions{
				Owner: qtest.MockOrg,
				Path:  mockIssues[2].Repository,
			},
			[]*scm.Issue{mockIssues[2]},
			false,
		},
		{
			"incorrect repository",
			&scm.RepositoryOptions{
				Owner: qtest.MockOrg,
				Path:  "unknown-labs",
			},
			nil,
			true,
		},
		{
			"incorrect organization",
			&scm.RepositoryOptions{
				Owner: "some-org",
				Path:  mockIssues[0].Repository,
			},
			nil,
			true,
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
			t.Errorf("%s mismatch (-want +got):\n%s", tt.name, diff)
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
			"delete all issues for 'user-labs' repo (issue 3)",
			&scm.RepositoryOptions{
				Path:  qf.StudentRepoName(user),
				Owner: course.OrganizationName,
			},
			map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1]},
			false,
		},
		{
			"delete all issues for 'test-labs' repo (issues 1 and 2)",
			&scm.RepositoryOptions{
				Path:  "test-labs",
				Owner: course.OrganizationName,
			},
			map[uint64]*scm.Issue{3: mockIssues[2]},
			false,
		},
		{
			"missing repository, nothing deleted",
			&scm.RepositoryOptions{
				Owner: course.OrganizationName,
				Path:  "some-labs",
			},
			map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1], 3: mockIssues[2]},
			true,
		},
		{
			"incorrect organization name",
			&scm.RepositoryOptions{
				Owner: "organization",
				Path:  "test-labs",
			},
			map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1], 3: mockIssues[2]},
			true,
		},
		{
			"invalid opt",
			&scm.RepositoryOptions{},
			map[uint64]*scm.Issue{1: mockIssues[0], 2: mockIssues[1], 3: mockIssues[2]},
			true,
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
			t.Errorf("%s mismatch (-want +got):\n%s", tt.name, diff)
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
			"comment 1 for issue 1",
			&scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       1,
			},
			1,
			false,
		},
		{
			"comment 2 for issue 1",
			&scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       1,
			},
			2,
			false,
		},
		{
			"comment 1 for issue 2",
			&scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       2,
			},
			3,
			false,
		},
		{
			"comment 2 for issue 2",
			&scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("test"),
				Body:         "Comment",
				Number:       2,
			},
			4,
			false,
		},
		{
			"invalid opts, missing organization",
			&scm.IssueCommentOptions{
				Repository: qf.StudentRepoName("test"),
				Body:       "Comment",
				Number:     1,
			},
			0,
			true,
		},
		{
			"invalid opts, missing repository",
			&scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Body:         "Comment",
				Number:       1,
			},
			0,
			true,
		},
		{
			"invalid opts, missing comment body",
			&scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName(user),
				Number:       1,
			},
			0,
			true,
		},
		{
			"incorrect organization name",
			&scm.IssueCommentOptions{
				Organization: "organization",
				Repository:   qf.StudentRepoName(user),
				Body:         "Comment",
				Number:       1,
			},
			0,
			true,
		},
		{
			"incorrect issue number",
			&scm.IssueCommentOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName(user),
				Body:         "Comment",
				Number:       5,
			},
			0,
			true,
		},
	}
	for _, tt := range tests {
		commentNumber, err := s.CreateIssueComment(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(tt.wantNumber, commentNumber); diff != "" {
			t.Errorf("%s mismatch (-want +got):\n%s", tt.name, diff)
		}
	}
}
