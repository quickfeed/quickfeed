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

func TestMockClone(t *testing.T) {
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
			},
			wantPath: filepath.Join("testdata", "assignments"),
			wantErr:  false,
		},
		{
			name: "tests repository",
			opt: &scm.CloneOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.TestsRepo,
			},
			wantPath: filepath.Join("testdata", "tests"),
			wantErr:  false,
		},
		{
			name: "missing organization info",
			opt: &scm.CloneOptions{
				Repository: qf.StudentRepoName("user"),
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
			Organization: &qf.Organization{ID: repo.OrgID, Name: repo.Owner},
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

func TestMockHooks(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	for _, course := range qtest.MockCourses {
		if err := s.CreateHook(ctx, &scm.CreateHookOptions{
			Organization: course.OrganizationName,
			URL:          "/test/hook",
		}); err != nil {
			t.Error(err)
		}
	}
	hooks, err := s.ListHooks(ctx, &scm.Repository{}, "")
	if err != nil {
		t.Error(err)
	}
	if len(hooks) != len(qtest.MockCourses) {
		t.Errorf("expected %d hooks, got %d", len(qtest.MockCourses), len(hooks))
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
	teamMemberTests := []struct {
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

	for _, tt := range teamMemberTests {
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
	teamMemberTests := []struct {
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
	for _, tt := range teamMemberTests {
		if err := s.UpdateTeamMembers(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
	}
}

func TestTeamRepo(t *testing.T) {
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

	teamRepoTests := []struct {
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

	for _, tt := range teamRepoTests {
		if err := s.AddTeamRepo(ctx, tt.opt); (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
	}
}
