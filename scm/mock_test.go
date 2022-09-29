package scm_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestMockOrganizations(t *testing.T) {
	testUser := "test_user"
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	// All organizations must be retrievable by ID and by name.
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
			Username:     testUser,
		}); err != nil {
			t.Error(err)
		}
		if err := s.RemoveMember(ctx, &scm.OrgMembershipOptions{
			Organization: course.OrganizationName,
			Username:     testUser,
		}); err != nil {
			t.Error(err)
		}
	}
	if err := s.UpdateOrganization(ctx, &scm.OrganizationOptions{
		Name: qtest.MockCourses[0].OrganizationName}); err == nil {
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
		{id: 123, name: "test_missing_org", username: testUser, permission: "read", err: "organization not found"},
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
	repos := []*scm.Repository{
		{
			OrgID: qtest.MockCourses[0].OrganizationID,
			Owner: qtest.MockCourses[0].OrganizationName,
			Path:  "info",
		},
		{
			OrgID: qtest.MockCourses[0].OrganizationID,
			Owner: qtest.MockCourses[0].OrganizationName,
			Path:  "tests",
		},
		{
			OrgID: qtest.MockCourses[2].OrganizationID,
			Owner: qtest.MockCourses[2].OrganizationName,
			Path:  "assignments",
		},
		{
			OrgID: qtest.MockCourses[2].OrganizationID,
			Owner: qtest.MockCourses[2].OrganizationName,
			Path:  "tests"},
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
	}

	for _, repo := range repos {
		r, err := s.GetRepository(ctx, &scm.RepositoryOptions{
			ID:    repo.ID,
			Path:  repo.Path,
			Owner: repo.Owner,
		})
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(repo, r, cmpopts.IgnoreFields(scm.Repository{}, "HTMLURL")); diff != "" {
			t.Errorf("Expected same repository, got (-sub +want):\n%s", diff)
		}
	}

	wantRepos := []*scm.Repository{repos[0], repos[1]}
	courseRepos, err := s.GetRepositories(ctx, &qf.Organization{ID: qtest.MockCourses[0].OrganizationID})
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
	courseRepos, err = s.GetRepositories(ctx, &qf.Organization{ID: qtest.MockCourses[2].OrganizationID})
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
		newTeam, err := s.CreateTeam(ctx, &scm.NewTeamOptions{
			Organization: team.Organization,
			TeamName:     team.Name,
		})
		if err != nil {
			t.Error(err)
		}
		team.ID = newTeam.ID
	}

	opts := []*scm.TeamOptions{
		{
			TeamID:         2,
			OrganizationID: course.OrganizationID,
		},
		{
			TeamName:     teams[1].Name,
			Organization: teams[1].Organization,
		},
	}
	for _, opt := range opts {
		team, err := s.GetTeam(ctx, opt)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(team, teams[1]); diff != "" {
			t.Errorf("Expected same team, got (-sub +want):\n%s", diff)
		}
	}

	if err := s.DeleteTeam(ctx, opts[0]); err != nil {
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

func TestMockTeamMembers(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	course := qtest.MockCourses[0]
	user := "test_user"
	team, err := s.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: course.OrganizationName,
		TeamName:     "test_team",
	})
	if err != nil {
		t.Fatal(err)
	}
	invalidTeamOptions := []*scm.TeamMembershipOptions{
		{},
		{
			Organization:   course.OrganizationName,
			OrganizationID: course.OrganizationID,
			Username:       user,
		},
		{
			TeamID:   team.ID,
			TeamName: team.Name,
			Username: user,
		},
		{
			Organization:   course.OrganizationName,
			OrganizationID: course.OrganizationID,
			TeamID:         team.ID,
			TeamName:       team.Name,
		},
	}
	for _, opt := range invalidTeamOptions {
		if err := s.AddTeamMember(ctx, opt); err == nil {
			t.Error("Expected error: invalid argument")
		}
		if err := s.RemoveTeamMember(ctx, opt); err == nil {
			t.Error("Expected error: invalid argument")
		}
		if err := s.UpdateTeamMembers(ctx, &scm.UpdateTeamOptions{
			TeamID: opt.TeamID,
		}); err == nil {
			t.Error("Expected error: invalid argument")
		}
		if err := s.UpdateTeamMembers(ctx, &scm.UpdateTeamOptions{
			OrganizationID: course.OrganizationID,
		}); err == nil {
			t.Error("Expected error: invalid argument")
		}
	}

	validTeamOptions := []*scm.TeamMembershipOptions{
		{
			TeamID:         team.ID,
			OrganizationID: course.OrganizationID,
			Username:       user,
		},
		{
			TeamName:     team.Name,
			Organization: team.Organization,
			Username:     user,
		},
	}
	for _, opt := range validTeamOptions {
		if err := s.AddTeamMember(ctx, opt); err != nil {
			t.Error(err)
		}
		if err := s.RemoveTeamMember(ctx, opt); err != nil {
			t.Error(err)
		}
		if err := s.UpdateTeamMembers(ctx, &scm.UpdateTeamOptions{
			TeamID:         team.ID,
			OrganizationID: course.OrganizationID,
		}); err != nil {
			t.Error(err)
		}
	}
}
