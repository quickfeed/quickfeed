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
func NewGithubSCMClient(logger *zap.SugaredLogger, token string) *GithubSCM {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := github.NewClient(oauth2.NewClient(context.Background(), ts))
	return &GithubSCM{
		logger: logger,
		client: client,
		token:  token,
	}
}

// ListOrganizations implements the SCM interface.
func (s *GithubSCM) ListOrganizations(ctx context.Context) ([]*pb.Organization, error) {

	memberships, _, err := s.client.Organizations.ListOrgMemberships(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ListOrganizations: failed to get GitHub memberships: %w", err)
	}

	var orgs []*pb.Organization
	for _, membership := range memberships {
		// limit scm requests to organizations where the user is owner ("admin")
		// owner membership role is required to create a course
		if membership.GetRole() == OrgOwner {
			opt := &GetOrgOptions{
				ID: uint64(membership.Organization.GetID()),
			}
			userOrg, err := s.GetOrganization(ctx, opt)
			if err != nil {
				return nil, fmt.Errorf("ListOrganizations: failed to get GitHub organization %s: %w", membership.Organization.GetLogin(), err)
			}
			orgs = append(orgs, userOrg)
		}
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

// UpdateOrganization implements the SCM interface.
func (s *GithubSCM) UpdateOrganization(ctx context.Context, opt *CreateOrgOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateOrganization",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	_, _, err := s.client.Organizations.Edit(ctx, opt.Path, &github.Organization{DefaultRepoPermission: &opt.DefaultPermission})
	return err
}

// GetOrganization implements the SCM interface.
func (s *GithubSCM) GetOrganization(ctx context.Context, opt *GetOrgOptions) (*pb.Organization, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "GetOrganization",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	var gitOrg *github.Organization
	var err error
	// priority is getting the organization by ID
	if opt.ID > 0 {
		gitOrg, _, err = s.client.Organizations.GetByID(ctx, int64(opt.ID))
		// if no ID provided, get by name
	} else {
		gitOrg, _, err = s.client.Organizations.Get(ctx, slug.Make(opt.Name))

	}
	if err != nil || gitOrg == nil {
		return nil, ErrFailedSCM{
			Method:   "GetOrganization",
			Message:  fmt.Sprintf("failed to get github organization, make sure it allows third party access"),
			GitError: err,
		}
	}

	// if user name is provided, return the found organization only if the user is one of its owners
	if opt.Username != "" {
		// fetch user membersip in that organization, if exists
		membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, slug.Make(opt.Name))
		if err != nil {
			return nil, ErrNotMember
		}
		// membership role must be "admin", if not, return error (possibly to show user)
		if membership.GetRole() != OrgOwner {
			return nil, ErrNotOwner
		}
	}

	return &pb.Organization{
		ID:          uint64(gitOrg.GetID()),
		Path:        gitOrg.GetLogin(),
		Avatar:      gitOrg.GetAvatarURL(),
		PaymentPlan: gitOrg.GetPlan().GetName(),
	}, nil
}

// CreateRepository implements the SCM interface.
func (s *GithubSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "CreateRepository",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	// first make sure that repo does not already exist for this user or group
	repo, _, err := s.client.Repositories.Get(ctx, opt.Owner, opt.Path)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "CreateRepository",
			Message:  fmt.Sprintf("failed to create repository %s, make sure it does not already exist", opt.Path),
			GitError: err,
		}
	}
	return toRepository(repo), nil
}

// GetRepository implements the SCM interface.
func (s *GithubSCM) GetRepository(ctx context.Context, opt *RepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "GetRepository",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	var repo *github.Repository
	var err error
	// if ID is set, get by ID
	if opt.ID > 0 {
		repo, _, err = s.client.Repositories.GetByID(ctx, int64(opt.ID))
	} else {
		// otherwise get by repo name and owner (usually owner = organization name)
		repo, _, err = s.client.Repositories.Get(ctx, opt.Owner, opt.Path)
	}
	if err != nil {
		return nil, fmt.Errorf("GetRepository failed to fetch a repo with ID %d and path % s: %w", opt.ID, opt.Path, err)
	}

	return toRepository(repo), nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, org *pb.Organization) ([]*Repository, error) {
	if !org.IsValid() {
		return nil, ErrMissingFields{
			Method:  "GetRepositories",
			Message: fmt.Sprintf("%+v", org),
		}
	}
	var path string
	if org.Path != "" {
		path = org.Path
	} else {
		opt := &GetOrgOptions{
			ID: org.ID,
		}
		org, err := s.GetOrganization(ctx, opt)
		if err != nil {
			return nil, err
		}
		path = org.Path
	}

	repos, _, err := s.client.Repositories.ListByOrg(ctx, path, nil)
	if err != nil {
		return nil, ErrFailedSCM{
			GitError: err,
			Method:   "GetRepositories",
			Message:  fmt.Sprintf("failed to get repositories for organization %s", path),
		}
	}

	var repositories []*Repository
	for _, repo := range repos {
		repositories = append(repositories, toRepository(repo))
	}
	return repositories, nil
}

// DeleteRepository implements the SCM interface.
func (s *GithubSCM) DeleteRepository(ctx context.Context, opt *RepositoryOptions) error {
	if !opt.valid() {
		s.logger.Errorf("DeleteRepository got invalid RepositoryOptions: %+v", opt)
	}

	// if ID provided, get path and owner from github
	if opt.ID > 0 {
		repo, _, err := s.client.Repositories.GetByID(ctx, int64(opt.ID))
		if err != nil {
			return ErrFailedSCM{
				GitError: err,
				Method:   "DeleteRepository",
				Message:  fmt.Sprintf("repository not found, make sure it exists in the course organization"),
			}
		}
		opt.Path = repo.GetName()
		opt.Owner = repo.Owner.GetLogin()
	}

	if _, err := s.client.Repositories.Delete(ctx, opt.Owner, opt.Path); err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "DeleteRepository",
			Message:  fmt.Sprintf("failed to delete repository %s", opt.Path),
		}
	}
	return nil
}

// UpdateRepoAccess implements the SCM interface.
func (s *GithubSCM) UpdateRepoAccess(ctx context.Context, repo *Repository, user, permission string) error {
	if repo == nil || !repo.valid() {
		return ErrMissingFields{
			Method:  "UpdateRepoAccess",
			Message: fmt.Sprintf("%+v", repo),
		}
	}
	opt := &github.RepositoryAddCollaboratorOptions{
		Permission: permission,
	}
	_, err := s.client.Repositories.AddCollaborator(ctx, repo.Owner, repo.Path, user, opt)
	return ErrFailedSCM{
		GitError: err,
		Method:   "UpdateRepoAccess",
		Message:  fmt.Sprintf("failed to grant %s permission to user %s for repository %s", opt.Permission, user, repo.Path),
	}
}

// RepositoryIsEmpty implements the SCM interface
func (s *GithubSCM) RepositoryIsEmpty(ctx context.Context, opt *RepositoryOptions) bool {

	repo, err := s.GetRepository(ctx, opt)
	if err != nil {
		return false
	}

	// test to check how repo commits look like
	_, _, err = s.client.Repositories.ListCommits(ctx, repo.Owner, repo.Path, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Git Repository is empty") {
			return true
		}
	}
	return false
}

// ListHooks implements the SCM interface.
func (s *GithubSCM) ListHooks(ctx context.Context, repo *Repository, org string) ([]*Hook, error) {
	var gitHooks []*github.Hook
	var hooks []*Hook

	if repo == nil || !repo.valid() {
		return nil, ErrMissingFields{
			Method:  "ListHooks",
			Message: fmt.Sprintf("%+v", org),
		}
	}

	githubHooks, _, err := s.client.Repositories.ListHooks(ctx, repo.Owner, repo.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("ListHooks: failed to list GitHub hooks: %w", err)
	}
	gitHooks = githubHooks

	// if org name provided, get all hooks existing on that organization
	if org != "" {
		orgName := slug.Make(org)
		githubHooks, _, err := s.client.Organizations.ListHooks(ctx, orgName, nil)
		if err != nil {
			return nil, fmt.Errorf("ListHooks: failed to list GitHub hooks for organization %s: %w", orgName, err)
		}
		gitHooks = githubHooks
	}
	if len(gitHooks) < 1 {
		s.logger.Debugf("ListHooks: invalid payload. Repo: %v, org: %s", repo, org)
		return nil, fmt.Errorf("ListHooks: found no hooks")
	}
	for _, hook := range gitHooks {
		s.logger.Infof("Found hook with events: %s", hook.Events)
		hooks = append(hooks, &Hook{
			ID:     uint64(hook.GetID()),
			URL:    hook.GetURL(),
			Events: hook.Events,
		})
	}
	return hooks, nil
}

// CreateHook implements the SCM interface.
func (s *GithubSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "CreateHook",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
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

// CreateOrgHook implements the scm interface
func (s *GithubSCM) CreateOrgHook(ctx context.Context, opt *OrgHookOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "CreateOrgHook",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	hook := &github.Hook{
		Config: map[string]interface{}{
			"url":          opt.URL,
			"secret":       opt.Secret,
			"content_type": "json",
			"insecure_ssl": "0",
		},
	}

	_, _, err := s.client.Organizations.CreateHook(ctx, opt.Organization.GetPath(), hook)
	if err != nil {
		return fmt.Errorf("CreateOrgHook: failed to create GitHub hook for org %s: %w", opt.Organization.Path, err)
	}
	return nil
}

// CreateTeam implements the SCM interface.
func (s *GithubSCM) CreateTeam(ctx context.Context, opt *TeamOptions) (*Team, error) {
	if !opt.validWithOrg() {
		return nil, ErrMissingFields{
			Method:  "CreateTeam",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	t, _, err := s.client.Teams.CreateTeam(ctx, opt.Organization.Path, github.NewTeam{
		Name: opt.TeamName,
	})
	if err != nil {
		if opt.TeamName != TeachersTeam && opt.TeamName != StudentsTeam {
			return nil, fmt.Errorf("CreateTeam: failed to create GitHub team %s: %w", opt.TeamName, err)
		}
		// continue if it is one of standard teacher/student teams. Such teams can be safely reused
		s.logger.Infof("Team %s already exists on organization %s", opt.TeamName, opt.Organization.Path)
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
func (s *GithubSCM) DeleteTeam(ctx context.Context, opt *TeamOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "DeleteTeam",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
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
func (s *GithubSCM) GetTeam(ctx context.Context, opt *TeamOptions) (scmTeam *Team, err error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "GetTeam",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
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
	if !org.IsValid() {
		return nil, ErrMissingFields{
			Method:  "GetTeams",
			Message: fmt.Sprintf("%+v", org),
		}
	}
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
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "AddTeamMember",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	team, err := s.GetTeam(ctx, &TeamOptions{
		Organization: opt.Organization,
		TeamName:     opt.TeamSlug,
		TeamID:       uint64(opt.TeamID),
	})
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "AddTeamMember",
			Message:  fmt.Sprintf("team %s not found", opt.TeamSlug),
		}
	}

	_, _, err = s.client.Teams.AddTeamMembership(ctx, int64(team.ID), opt.Username,
		&github.TeamAddTeamMembershipOptions{Role: opt.Role})
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "AddTeamMember",
			Message:  fmt.Sprintf("failed to add user %s to team %s with role %s", opt.Username, opt.TeamSlug, opt.Role),
		}
	}
	return nil
}

// RemoveTeamMember implements the scm interface
func (s *GithubSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RemoveTeamMember",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	team, err := s.GetTeam(ctx, &TeamOptions{Organization: opt.Organization, TeamName: opt.TeamSlug, TeamID: uint64(opt.TeamID)})
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "RemoveTeamMember",
			Message:  fmt.Sprintf("team %s not found", opt.TeamSlug),
		}
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
		return ErrFailedSCM{
			GitError: err,
			Method:   "RemoveTeamMember",
			Message:  fmt.Sprintf("failed to remove user %s from team %s", opt.Username, opt.TeamSlug),
		}
	}
	return nil
}

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *TeamOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateTeamMembers",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	groupTeam, _, err := s.client.Teams.GetTeam(ctx, int64(opt.TeamID))
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "UpdateTeamMember",
			Message:  fmt.Sprintf("team with ID %d not found", opt.TeamID),
		}
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
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "AddTeamRepo",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	_, err := s.client.Teams.AddTeamRepo(ctx, int64(opt.TeamID), opt.Owner, opt.Repo,
		&github.TeamAddTeamRepoOptions{
			Permission: opt.Permission, // make sure users can pull and push
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

// UpdateOrgMembership implements the SCM interface
func (s *GithubSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateOrgMembership",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	newMembership, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.Username, opt.Organization.Path, &github.Membership{Role: &opt.Role})
	if err != nil || newMembership.GetRole() != opt.Role {
		// we should not wrap error here because it is potentially nil
		return fmt.Errorf("UpdateOrgMembership: failed to edit GitHub org membership for user '%s'", opt.Username)
	}
	return nil
}

// RemoveMember implements the SCM interface
func (s *GithubSCM) RemoveMember(ctx context.Context, opt *OrgMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RevokeOrgMembership",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	// remove user from the organization and all teams
	_, err := s.client.Organizations.RemoveMember(ctx, opt.Organization.GetPath(), opt.Username)
	if err != nil {
		return ErrFailedSCM{
			Method:   "RemoveMember",
			GitError: err,
			Message:  fmt.Sprintf("failed to remove user %s from organization %s", opt.Username, opt.Organization.GetPath()),
		}
	}
	return nil
}

// GetUserScopes implements the SCM interface
func (s *GithubSCM) GetUserScopes(ctx context.Context) *Authorization {
	// Authorizations.List method will always return nill, response struct and error,
	// we are only interested in response. Its header will contain all scopes for current user
	// TODO(meling) @Vera: the above comment needs to be clarified a little more.
	_, resp, _ := s.client.Authorizations.List(ctx, &github.ListOptions{})
	if resp == nil {
		s.logger.Errorf("GetUserScopes: got no scopes: no authorized user")
		tmpScopes := make([]string, 0)
		return &Authorization{Scopes: tmpScopes}
	}
	// header contains a single string with all scopes for authenticated user
	stringScopes := resp.Header.Get("X-OAuth-Scopes")
	if stringScopes == "" {
		s.logger.Errorf("GetUserScopes: header was empty")
		tmpScopes := make([]string, 0)
		return &Authorization{Scopes: tmpScopes}
	}
	s.logger.Debugf("GetUserScopes got scopes: %v", stringScopes)
	gitScopes := strings.Split(stringScopes, ", ")
	return &Authorization{Scopes: gitScopes}
}

func toRepository(repo *github.Repository) *Repository {
	return &Repository{
		ID:      uint64(repo.GetID()),
		Path:    repo.GetName(),
		Owner:   repo.Owner.GetLogin(),
		WebURL:  repo.GetHTMLURL(),
		SSHURL:  repo.GetSSHURL(),
		HTTPURL: repo.GetCloneURL(),
		OrgID:   uint64(repo.Organization.GetID()),
		Size:    uint64(repo.GetSize()),
	}
}
