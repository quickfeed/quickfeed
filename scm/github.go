package scm

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	pb "github.com/autograde/aguis/ag"
	"github.com/google/go-github/v26/github"
	"github.com/gosimple/slug"
	"golang.org/x/oauth2"
)

// GithubSCM implements the SCM interface.
type GithubSCM struct {
	logger *zap.SugaredLogger
	client *github.Client
	token  string
}

// NewGithubSCMClient returns a new Github client implementing the SCM interface.
func NewGithubSCMClient(logger *zap.Logger, token string) *GithubSCM {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := github.NewClient(oauth2.NewClient(context.Background(), ts))
	return &GithubSCM{
		logger: logger.Sugar(),
		client: client,
		token:  token,
	}
}

// ListOrganizations implements the SCM interface.
func (s *GithubSCM) ListOrganizations(ctx context.Context) ([]*pb.Organization, error) {
	userOrgs, _, err := s.client.Organizations.ListOrgMemberships(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ListOrganizations: failed to get GitHub organizations: %w", err)
	}

	var orgs []*pb.Organization
	for _, org := range userOrgs {
		// TODO (vera) - with this update we will have to access GitHub for each organization
		// we cannot just use org.Organization.GetPlan() to set payment plan, as it is always null here
		// because of that we need to retrieve every organization from GitHub to get access to its payment plan
		// not sure if it is really an improvement
		userOrg, err := s.GetOrganization(ctx, uint64(org.Organization.GetID()))
		if err != nil {
			return nil, fmt.Errorf("ListOrganizations: failed to get GitHub organization: %w", err)
		}
		orgs = append(orgs, userOrg)
	}
	return orgs, nil
}

// CreateOrganization implements the SCM interface.
func (s *GithubSCM) CreateOrganization(ctx context.Context, opt *CreateOrgOptions) (*pb.Organization, error) {
	return nil, ErrNotSupported{
		SCM:    "github",
		Method: "CreateOrganization",
	}
}

// GetOrganization implements the SCM interface.
func (s *GithubSCM) GetOrganization(ctx context.Context, id uint64) (*pb.Organization, error) {
	org, _, err := s.client.Organizations.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: failed to get GitHub organization: %w", err)
	}
	return &pb.Organization{
		ID:          uint64(org.GetID()),
		Path:        org.GetLogin(),
		Avatar:      org.GetAvatarURL(),
		PaymentPlan: org.GetPlan().GetName(),
	}, nil
}

// CreateRepository implements the SCM interface.
func (s *GithubSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	repo, _, err := s.client.Repositories.Create(ctx, opt.Organization.Path, &github.Repository{
		Name:    &opt.Path,
		Private: &opt.Private,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateRepositories: failed to create GitHub repository: %w", err)
	}

	return &Repository{
		ID:      uint64(repo.GetID()),
		Path:    repo.GetName(),
		Owner:   repo.Owner.GetLogin(), // this is safe against nil
		WebURL:  repo.GetHTMLURL(),
		SSHURL:  repo.GetSSHURL(),
		HTTPURL: repo.GetCloneURL(),
		OrgID:   opt.Organization.ID,
	}, nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, org *pb.Organization) ([]*Repository, error) {
	var path string
	if org.Path != "" {
		path = org.Path
	} else {
		org, err := s.GetOrganization(ctx, org.ID)
		if err != nil {
			return nil, fmt.Errorf("GetRepositories: failed to get GitHub organization: %w", err)
		}
		path = org.Path
	}

	repos, _, err := s.client.Repositories.ListByOrg(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("GetRepositories: failed to get GitHub repositories: %w", err)
	}

	var repositories []*Repository
	for _, repo := range repos {
		repositories = append(repositories, &Repository{
			ID:      uint64(repo.GetID()),
			Path:    repo.GetName(),
			Owner:   repo.Owner.GetLogin(), // this is safe against nil
			WebURL:  repo.GetHTMLURL(),
			SSHURL:  repo.GetSSHURL(),
			HTTPURL: repo.GetCloneURL(),
			OrgID:   org.ID,
		})
	}

	return repositories, nil
}

// DeleteRepository implements the SCM interface.
func (s *GithubSCM) DeleteRepository(ctx context.Context, id uint64) error {
	repo, _, err := s.client.Repositories.GetByID(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("DeleteRepository: failed to get GitHub repository: %w", err)
	}
	if _, err := s.client.Repositories.Delete(ctx, repo.Owner.GetLogin(), repo.GetName()); err != nil {
		return fmt.Errorf("DeleteRepository: failed to delete GitHub repository: %w", err)
	}
	return nil
}

// ListHooks implements the SCM interface.
func (s *GithubSCM) ListHooks(ctx context.Context, repo *Repository) ([]*Hook, error) {
	githubHooks, _, err := s.client.Repositories.ListHooks(ctx, repo.Owner, repo.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("ListHooks: failed to list GitHub hooks: %w", err)
	}
	var hooks []*Hook
	for _, hook := range githubHooks {
		hooks = append(hooks, &Hook{
			ID:  uint64(hook.GetID()),
			URL: hook.GetURL(),
		})
	}
	return hooks, nil
}

// CreateHook implements the SCM interface.
func (s *GithubSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) error {
	_, _, err := s.client.Repositories.CreateHook(ctx, opt.Repository.Owner, opt.Repository.Path,
		&github.Hook{
			Config: map[string]interface{}{
				"url":          opt.URL,
				"secret":       opt.Secret,
				"content_type": "json",
				"insecure_ssl": "0",
			},
		})
	if err != nil {
		return fmt.Errorf("CreateHook: failed to create GitHub hook for %s: %w", opt.Repository.Path, err)
	}
	return nil
}

// CreateTeam implements the SCM interface.
func (s *GithubSCM) CreateTeam(ctx context.Context, opt *CreateTeamOptions) (*Team, error) {
	t, _, err := s.client.Teams.CreateTeam(ctx, opt.Organization.Path, github.NewTeam{
		Name: opt.TeamName,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateTeam: failed to create GitHub team: %w", err)
	}

	for _, user := range opt.Users {
		_, _, err = s.client.Teams.AddTeamMembership(ctx, t.GetID(), user, nil)
		if err != nil {
			return nil, fmt.Errorf("CreateTeam: failed to add '%s' to GitHub team '%s': %w", user, t.GetName(), err)
		}
	}
	return &Team{
		ID:   uint64(t.GetID()),
		Name: t.GetName(),
		URL:  t.GetURL(),
	}, nil
}

// DeleteTeam implements the SCM interface.
func (s *GithubSCM) DeleteTeam(ctx context.Context, opt *CreateTeamOptions) error {
	team, err := s.GetTeam(ctx, opt)
	if err != nil {
		return fmt.Errorf("DeleteTeam: failed to get GitHub team '%s': %w", opt.TeamName, err)
	}

	if _, err := s.client.Teams.DeleteTeam(ctx, int64(team.ID)); err != nil {
		return fmt.Errorf("DeleteTeam: failed to delete GitHub team '%s': %w", opt.TeamName, err)
	}
	return nil
}

// GetTeam implements the SCM interface
func (s *GithubSCM) GetTeam(ctx context.Context, opt *CreateTeamOptions) (scmTeam *Team, err error) {
	var team *github.Team
	if opt.TeamID < 1 {
		slug := slug.Make(opt.TeamName)
		team, _, err = s.client.Teams.GetTeamBySlug(ctx, opt.Organization.Path, slug)
		if err != nil {
			return nil, fmt.Errorf("GetTeam: failed to get GitHub team by slug '%s': %w", slug, err)
		}
	} else {
		team, _, err = s.client.Teams.GetTeam(ctx, int64(opt.TeamID))
		if err != nil {
			return nil, fmt.Errorf("GetTeam: failed to get GitHub team by ID '%d': %w", opt.TeamID, err)
		}
	}
	return &Team{
		ID:   uint64(team.GetID()),
		Name: team.GetName(),
		URL:  team.GetURL(),
	}, nil
}

// GetTeams implements the scm interface
func (s *GithubSCM) GetTeams(ctx context.Context, org *pb.Organization) ([]*Team, error) {
	gitTeams, _, err := s.client.Teams.ListTeams(ctx, org.Path, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("GetTeams: failed to list GitHub teams: %w", err)
	}
	var teams []*Team
	for _, gitTeam := range gitTeams {
		newTeam := &Team{ID: uint64(gitTeam.GetID()), Name: gitTeam.GetName(), URL: gitTeam.GetURL()}
		teams = append(teams, newTeam)
	}
	return teams, nil
}

// AddTeamMember implements the scm interface
func (s *GithubSCM) AddTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	team, err := s.GetTeam(ctx, &CreateTeamOptions{
		Organization: opt.Organization,
		TeamName:     opt.TeamSlug,
		TeamID:       uint64(opt.TeamID),
	})
	if err != nil {
		// just logging the error; not returning
		s.logger.Debugf("AddTeamMember: failed to get GitHub team '%s': %w", opt.TeamSlug, err)
		return err
	}

	isAlreadyMember, _, err := s.client.Teams.GetTeamMembership(ctx, int64(team.ID), opt.Username)
	if err != nil {
		// this will always return an error when the user is not member of team.
		// this is expected and no error will be returned, but it is still useful
		// to log it in case there were other reasons (invalid token and others).
		s.logger.Debugf("AddTeamMember: failed to get GitHub team membership for '%s' (user %s): %w",
			opt.TeamSlug, opt.Username, err)
	}
	if isAlreadyMember != nil {
		// if already in team, log it, and return without further action
		s.logger.Debugf("AddTeamMember: GitHub user '%s' is already member of team '%s'", opt.Username, opt.TeamSlug)
		return nil
	}
	// otherwise add user to team
	_, _, err = s.client.Teams.AddTeamMembership(ctx, int64(team.ID), opt.Username,
		&github.TeamAddTeamMembershipOptions{Role: opt.Role})
	if err != nil {
		return fmt.Errorf("AddTeamMember: failed to add member '%s' to GitHub team '%s': %w", opt.Username, opt.TeamSlug, err)
	}
	return nil
}

// RemoveTeamMember implements the scm interface
func (s *GithubSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	team, err := s.GetTeam(ctx, &CreateTeamOptions{Organization: opt.Organization, TeamName: opt.TeamSlug, TeamID: uint64(opt.TeamID)})
	if err != nil {
		// just logging the error; not returning
		s.logger.Debugf("RemoveTeamMember: failed to get GitHub team '%s': %w", opt.TeamSlug, err)
	}

	isMember, _, err := s.client.Teams.GetTeamMembership(ctx, int64(team.ID), opt.Username)
	if err != nil {
		// this will always return an error when the user is not member of team.
		// this is expected and no error will be returned, but it is still useful
		// to log it in case there were other reasons (invalid token and others).
		s.logger.Debugf("RemoveTeamMember: failed to get GitHub team membership for '%s' (user %s): %w",
			opt.TeamSlug, opt.Username, err)
	}
	if isMember == nil {
		// user is not in this team, log it, and return without further action
		s.logger.Debugf("RemoveTeamMember: GitHub user '%s' is not member of team '%s'", opt.Username, opt.TeamSlug)
		return nil
	}
	// otherwise remove user from team
	_, err = s.client.Teams.RemoveTeamMembership(ctx, int64(team.ID), opt.Username)
	if err != nil {
		return fmt.Errorf("RemoveTeamMember: failed to remove member '%s' from GitHub team '%s': %w", opt.Username, opt.TeamSlug, err)
	}
	return nil
}

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *CreateTeamOptions) error {
	groupTeam, _, err := s.client.Teams.GetTeam(ctx, int64(opt.TeamID))
	if err != nil {
		return fmt.Errorf("UpdateTeamMember: failed to get GitHub team '%s': %w", opt.TeamName, err)
	}

	// find current team members
	oldUsers, _, err := s.client.Teams.ListTeamMembers(ctx, groupTeam.GetID(), nil)
	if err != nil {
		return fmt.Errorf("UpdateTeamMember: failed to get members for GitHub team '%s': %w", opt.TeamName, err)
	}

	// check whether group members are already in team; add missing members
	for _, member := range opt.Users {
		_, _, err = s.client.Teams.AddTeamMembership(ctx, groupTeam.GetID(), member, nil)
		if err != nil {
			return fmt.Errorf("UpdateTeamMember: failed to add user '%s' to GitHub team '%s': %w", member, opt.TeamName, err)
		}

	}

	// check if all the team members are in the new group;
	for _, teamMember := range oldUsers {
		toRemove := true
		for _, groupMember := range opt.Users {
			if teamMember.GetLogin() == groupMember {
				toRemove = false
			}
		}
		if toRemove {
			_, err = s.client.Teams.RemoveTeamMembership(ctx, groupTeam.GetID(), teamMember.GetLogin())
			if err != nil {
				return fmt.Errorf("UpdateTeamMember: failed to remove user '%s' from GitHub team '%s': %w", teamMember.GetLogin(), opt.TeamName, err)
			}
		}
	}
	return nil
}

// CreateCloneURL implements the SCM interface.
func (s *GithubSCM) CreateCloneURL(opt *CreateClonePathOptions) string {
	token := s.token
	if len(opt.UserToken) > 0 {
		token = opt.UserToken
	}
	return "https://" + token + "@github.com/" + opt.Organization + "/" + opt.Repository + ".git"
}

// AddTeamRepo implements the SCM interface.
func (s *GithubSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	_, err := s.client.Teams.AddTeamRepo(ctx, int64(opt.TeamID), opt.Owner, opt.Repo,
		&github.TeamAddTeamRepoOptions{
			Permission: "push", // make sure users can pull and push
		})
	if err != nil {
		return fmt.Errorf("AddTeamRepo: failed to add team '%d' to GitHub repository '%s': %w",
			opt.TeamID, opt.Repo, err)
	}
	return nil
}

// GetUserName implements the SCM interface.
func (s *GithubSCM) GetUserName(ctx context.Context) (string, error) {
	user, _, err := s.client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("GetUserName: failed to get GitHub user: %w", err)
	}
	return user.GetLogin(), nil
}

// GetUserNameByID implements the SCM interface.
func (s *GithubSCM) GetUserNameByID(ctx context.Context, remoteID uint64) (string, error) {
	user, _, err := s.client.Users.GetByID(ctx, int64(remoteID))
	if err != nil {
		return "", fmt.Errorf("GetUserNameByID: failed to get GitHub user '%d': %w", remoteID, err)
	}
	return user.GetLogin(), nil
}

// GetOrgMembership implements the SCM interface
func (s *GithubSCM) GetOrgMembership(ctx context.Context, opt *OrgMembershipOptions) (*OrgMembershipOptions, error) {
	membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, opt.Organization.Path)
	if err != nil {
		return nil, fmt.Errorf("GetOrgMembership: failed to get GitHub org membership for user '%s': %w", opt.Username, err)
	}
	opt.Role = membership.GetRole()
	return opt, nil
}

// UpdateOrgMembership implements the SCM interface
func (s *GithubSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	isMember, _, err := s.client.Organizations.IsMember(ctx, opt.Organization.Path, opt.Username)
	if err != nil {
		return fmt.Errorf("GetOrgMembership: failed to check if user '%s' is member of GitHub organization '%s': %w",
			opt.Username, opt.Organization.Path, err)
	}
	if !isMember {
		return fmt.Errorf("GetOrgMembership: user '%s' is not member of GitHub organization '%s': %w",
			opt.Username, opt.Organization.Path, err)
	}
	gitMembership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, opt.Organization.Path)
	if err != nil {
		return fmt.Errorf("UpdateOrgMembership: failed to get GitHub org membership for user '%s': %w", opt.Username, err)
	}
	if opt.Role != "admin" && opt.Role != "member" {
		return fmt.Errorf("UpdateOrgMembership: invalid role '%s'", opt.Role)
	}
	gitMembership.Role = &opt.Role
	newMembership, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.Username, opt.Organization.Path, gitMembership)
	if err != nil || newMembership.GetRole() != opt.Role {
		return fmt.Errorf("UpdateOrgMembership: failed to edit GitHub org membership for user '%s': %w", opt.Username, err)
	}
	return nil
}

// GetUserScopes implements the SCM interface
func (s *GithubSCM) GetUserScopes(ctx context.Context) *Authorization {
	// Authorizations.List method will always return nill, response struct and error,
	// we are only interested in user scopes that are always contained in response header
	// TODO(meling) @Vera: the above comment needs to be clarified a little more.
	_, resp, _ := s.client.Authorizations.List(ctx, &github.ListOptions{})
	// header contains a single string with all scopes for authenticated user
	stringScopes := resp.Header.Get("X-OAuth-Scopes")
	gitScopes := strings.Split(stringScopes, ", ")
	return &Authorization{Scopes: gitScopes}
}
