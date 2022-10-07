package scm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/google/go-github/v45/github"
	"github.com/gosimple/slug"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GithubSCM implements the SCM interface.
type GithubSCM struct {
	logger      *zap.SugaredLogger
	client      *github.Client
	clientV4    *githubv4.Client
	config      *Config
	token       string
	providerURL string
	tokenURL    string
}

// NewGithubSCMClient returns a new Github client implementing the SCM interface.
func NewGithubSCMClient(logger *zap.SugaredLogger, token string) *GithubSCM {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return &GithubSCM{
		logger:      logger,
		client:      github.NewClient(httpClient),
		clientV4:    githubv4.NewClient(httpClient),
		token:       token,
		providerURL: "github.com",
	}
}

// UpdateOrganization implements the SCM interface.
func (s *GithubSCM) UpdateOrganization(ctx context.Context, opt *OrganizationOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateOrganization",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	_, _, err := s.client.Organizations.Edit(ctx, opt.Name, &github.Organization{
		DefaultRepoPermission: &opt.DefaultPermission,
		MembersCanCreateRepos: &opt.RepoPermissions,
	})
	return err
}

// GetOrganization implements the SCM interface.
func (s *GithubSCM) GetOrganization(ctx context.Context, opt *GetOrgOptions) (*qf.Organization, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "GetOrganization",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	var orgNameOrID string
	var gitOrg *github.Organization
	var err error
	// priority is getting the organization by ID
	if opt.ID > 0 {
		orgNameOrID = strconv.Itoa(int(opt.ID))
		gitOrg, _, err = s.client.Organizations.GetByID(ctx, int64(opt.ID))
	} else {
		// if ID not provided, get by name
		orgNameOrID = slug.Make(opt.Name)
		gitOrg, _, err = s.client.Organizations.Get(ctx, slug.Make(opt.Name))
	}
	if err != nil || gitOrg == nil {
		return nil, ErrFailedSCM{
			Method:   "GetOrganization",
			Message:  fmt.Sprintf("could not find github organization %s. Make sure it allows third party access.", orgNameOrID), // this message is logged, never sent to user
			GitError: err,
		}
	}

	// if user name is provided, return the found organization only if the user is one of its owners
	if opt.Username != "" {
		// fetch user membership in that organization, if exists
		membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, slug.Make(opt.Name))
		if err != nil {
			return nil, ErrFailedSCM{
				Method:   "GetOrganization",
				Message:  fmt.Sprintf("Failed to GetOrganization for (%q, %q)", opt.Username, slug.Make(opt.Name)),
				GitError: fmt.Errorf("failed to GetOrgMembership(%q, %q): %w", opt.Username, slug.Make(opt.Name), err),
			}
		}
		// membership role must be "admin", if not, return error (possibly to show user)
		if membership.GetRole() != OrgOwner {
			return nil, ErrNotOwner
		}
	}

	return &qf.Organization{
		ID:          uint64(gitOrg.GetID()),
		Name:        gitOrg.GetLogin(),
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

	// check that repo does not already exist for this user or group
	repo, _, err := s.client.Repositories.Get(ctx, opt.Organization, slug.Make(opt.Path))
	if repo != nil {
		s.logger.Debugf("CreateRepository: found existing repository (skipping creation): %s: %v", opt.Path, repo)
		return toRepository(repo), nil
	}
	// error expected to be 404 Not Found; logging here in case it's a different error
	s.logger.Debugf("CreateRepository: check for repository %s: %s", opt.Path, err)

	// repo does not exist, create it
	s.logger.Debugf("CreateRepository: creating %s", opt.Path)
	repo, _, err = s.client.Repositories.Create(ctx, opt.Organization, &github.Repository{
		Name:    &opt.Path,
		Private: &opt.Private,
	})
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "CreateRepository",
			Message:  fmt.Sprintf("failed to create repository %s, make sure it does not already exist", opt.Path),
			GitError: err,
		}
	}
	s.logger.Debugf("CreateRepository: done creating %s", opt.Path)

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
		return nil, fmt.Errorf("GetRepository failed to fetch repository %d, and path %s: %w", opt.ID, opt.Path, err)
	}

	return toRepository(repo), nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, org *qf.Organization) ([]*Repository, error) {
	if !org.IsValid() {
		return nil, ErrMissingFields{
			Method:  "GetRepositories",
			Message: fmt.Sprintf("%+v", org),
		}
	}
	var path string
	if org.Name != "" {
		path = org.Name
	} else {
		opt := &GetOrgOptions{
			ID: org.ID,
		}
		org, err := s.GetOrganization(ctx, opt)
		if err != nil {
			return nil, err
		}
		path = org.Name
	}

	repos, _, err := s.client.Repositories.ListByOrg(ctx, path, nil)
	if err != nil {
		return nil, ErrFailedSCM{
			GitError: err,
			Method:   "GetRepositories",
			Message:  fmt.Sprintf("failed to access repositories for organization %s", path),
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
				Message:  fmt.Sprintf("failed to fetch repository %d: may not exists in the course organization", opt.ID),
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
	if _, _, err := s.client.Repositories.AddCollaborator(ctx, repo.Owner, repo.Path, user, opt); err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "UpdateRepoAccess",
			Message:  fmt.Sprintf("failed to grant %s permission to user %s for repository %s", opt.Permission, user, repo.Path),
		}
	}
	return nil
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
func (s *GithubSCM) ListHooks(ctx context.Context, repo *Repository, org string) (hooks []*Hook, err error) {
	var githubHooks []*github.Hook

	// we prioritize organization hooks because repository hooks are no longer used.
	switch {
	case org != "":
		orgName := slug.Make(org)
		githubHooks, _, err = s.client.Organizations.ListHooks(ctx, orgName, nil)
		if err != nil {
			return nil, fmt.Errorf("ListHooks: failed to get hooks for organization %q: %w", orgName, err)
		}

	case repo != nil && repo.valid():
		githubHooks, _, err = s.client.Repositories.ListHooks(ctx, repo.Owner, repo.Path, nil)
		if err != nil {
			return nil, fmt.Errorf("ListHooks: failed to get hooks for repository %q: %w", repo, err)
		}

	default:
		return nil, fmt.Errorf("ListHooks: called with missing or incompatible arguments: %q %q", repo, org)
	}

	for _, hook := range githubHooks {
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
	if !opt.valid() || opt.Organization == "" {
		return ErrMissingFields{
			Method:  "CreateHook",
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
		Events: []string{"push", "pull_request", "pull_request_review"},
	}
	_, _, err := s.client.Organizations.CreateHook(ctx, opt.Organization, hook)
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "CreateHook",
			Message:  fmt.Sprintf("failed to create GitHub hook with query: %+v", opt),
		}
	}
	return err
}

// CreateTeam implements the SCM interface.
func (s *GithubSCM) CreateTeam(ctx context.Context, opt *NewTeamOptions) (*Team, error) {
	if !opt.valid() || opt.TeamName == "" || opt.Organization == "" {
		return nil, ErrMissingFields{
			Method:  "CreateTeam",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	// check that the team name does not already exist for this organization
	team, _, err := s.client.Teams.GetTeamBySlug(ctx, slug.Make(opt.Organization), slug.Make(opt.TeamName))
	if err != nil {
		// error expected to be 404 Not Found; logging here in case it's a different error
		s.logger.Debugf("CreateTeam: check for team %s: %s", opt.TeamName, err)
	}

	if team == nil {
		s.logger.Debugf("CreateTeam: creating %s", opt.TeamName)
		team, _, err = s.client.Teams.CreateTeam(ctx, opt.Organization, github.NewTeam{
			Name: opt.TeamName,
		})
		if err != nil {
			if opt.TeamName != TeachersTeam && opt.TeamName != StudentsTeam {
				return nil, ErrFailedSCM{
					Method:   "CreateTeam",
					Message:  fmt.Sprintf("failed to create GitHub team %s, make sure it does not already exist", opt.TeamName),
					GitError: fmt.Errorf("failed to create GitHub team %s: %w", opt.TeamName, err),
				}
			}
			// continue if it is one of standard teacher/student teams. Such teams can be safely reused
			s.logger.Debugf("Team %s already exists on organization %s", opt.TeamName, opt.Organization)
		}
		s.logger.Debugf("CreateTeam: done creating %s", opt.TeamName)
	}
	for _, user := range opt.Users {
		s.logger.Debugf("CreateTeam: adding user %s to %s", user, opt.TeamName)
		_, _, err = s.client.Teams.AddTeamMembershipByID(ctx, team.GetOrganization().GetID(), team.GetID(), user, nil)
		if err != nil {
			return nil, ErrFailedSCM{
				Method:   "CreateTeam",
				Message:  fmt.Sprintf("failed to add user '%s' to GitHub team '%s'", user, team.GetName()),
				GitError: fmt.Errorf("failed to add '%s' to GitHub team '%s': %w", user, team.GetName(), err),
			}
		}
	}
	return &Team{
		ID:           uint64(team.GetID()),
		Name:         team.GetName(),
		Organization: team.GetOrganization().GetLogin(),
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

	var err error
	if opt.TeamID > 0 {
		_, err = s.client.Teams.DeleteTeamByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID))
	} else {
		_, err = s.client.Teams.DeleteTeamBySlug(ctx, slug.Make(opt.Organization), slug.Make(opt.TeamName))
	}

	if err != nil {
		return ErrFailedSCM{
			Method:   "DeleteTeam",
			Message:  fmt.Sprintf("failed to delete GitHub team '%s'", opt.TeamName),
			GitError: fmt.Errorf("failed to get GitHub team '%s': %w", opt.TeamName, err),
		}
	}
	return err
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
		team, _, err = s.client.Teams.GetTeamBySlug(ctx, opt.Organization, slug)
		if err != nil {
			return nil, fmt.Errorf("GetTeam: failed to get GitHub team by slug '%s': %w", slug, err)
		}
	} else {
		team, _, err = s.client.Teams.GetTeamByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID))
		if err != nil {
			return nil, fmt.Errorf("GetTeam: failed to get GitHub team by ID '%d': %w", opt.TeamID, err)
		}
	}
	return &Team{
		ID:           uint64(team.GetID()),
		Name:         team.GetName(),
		Organization: team.Organization.GetLogin(),
	}, nil
}

// GetTeams implements the scm interface
func (s *GithubSCM) GetTeams(ctx context.Context, org *qf.Organization) ([]*Team, error) {
	if !org.IsValid() {
		return nil, ErrMissingFields{
			Method:  "GetTeams",
			Message: fmt.Sprintf("%+v", org),
		}
	}
	gitTeams, _, err := s.client.Teams.ListTeams(ctx, org.Name, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("GetTeams: failed to list GitHub teams: %w", err)
	}
	var teams []*Team
	for _, gitTeam := range gitTeams {
		newTeam := &Team{ID: uint64(gitTeam.GetID()), Name: gitTeam.GetName(), Organization: gitTeam.Organization.GetLogin()}
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

	var err error
	if opt.TeamID < 1 {
		_, _, err = s.client.Teams.AddTeamMembershipBySlug(ctx, opt.Organization, slug.Make(opt.TeamName), opt.Username,
			&github.TeamAddTeamMembershipOptions{Role: opt.Role})
	} else {
		_, _, err = s.client.Teams.AddTeamMembershipByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), opt.Username,
			&github.TeamAddTeamMembershipOptions{Role: opt.Role})
	}

	if err != nil {
		err = ErrFailedSCM{
			GitError: err,
			Method:   "AddTeamMember",
			Message:  fmt.Sprintf("failed to add user (%s) to team (ID %d, team name: %s) with role %s", opt.Username, opt.TeamID, opt.TeamName, opt.Role),
		}
	}
	return err
}

// RemoveTeamMember implements the scm interface
func (s *GithubSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RemoveTeamMember",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	var err error
	if opt.TeamID < 1 {
		_, err = s.client.Teams.RemoveTeamMembershipBySlug(ctx, opt.Organization, opt.TeamName, opt.Username)
	} else {
		_, err = s.client.Teams.RemoveTeamMembershipByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), opt.Username)
	}

	if err != nil {
		err = ErrFailedSCM{
			GitError: err,
			Method:   "RemoveTeamMember",
			Message:  fmt.Sprintf("failed to remove user %s from team ID %d", opt.Username, opt.TeamID),
		}
	}
	return err
}

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *UpdateTeamOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateTeamMembers",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	// find current team members
	oldUsers, _, err := s.client.Teams.ListTeamMembersByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), nil)
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "UpdateTeamMember",
			Message:  fmt.Sprintf("failed to get members for team ID %d", opt.TeamID),
		}
	}

	// check whether group members are already in team; add missing members
	for _, member := range opt.Users {
		_, _, err = s.client.Teams.AddTeamMembershipByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), member, nil)
		if err != nil {
			return ErrFailedSCM{
				GitError: err,
				Method:   "UpdateTeamMember",
				Message:  fmt.Sprintf("failed to add user %s to team ID %d", member, opt.TeamID),
			}
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
			_, err = s.client.Teams.RemoveTeamMembershipByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), teamMember.GetLogin())
			if err != nil {
				return ErrFailedSCM{
					GitError: err,
					Method:   "UpdateTeamMember",
					Message:  fmt.Sprintf("failed to remove user %s from team ID %d", teamMember.GetLogin(), opt.TeamID),
				}
			}
		}
	}
	return nil
}

// CreateCloneURL implements the SCM interface.
func (s *GithubSCM) CreateCloneURL(opt *URLPathOptions) string {
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

	_, err := s.client.Teams.AddTeamRepoByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), opt.Owner, opt.Repo,
		&github.TeamAddTeamRepoOptions{
			Permission: opt.Permission, // make sure users can pull and push
		})
	if err != nil {
		return ErrFailedSCM{
			GitError: fmt.Errorf("failed to make GitHub repository '%s' a team repository for team %d: %w", opt.Repo, opt.TeamID, err),
			Method:   "AddTeamRepo",
			Message:  fmt.Sprintf("failed to make GitHub repository '%s' a team repository", opt.Repo),
		}
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
	newMembership, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.Username, opt.Organization, &github.Membership{Role: &opt.Role})
	if err != nil || newMembership.GetRole() != opt.Role {
		// Note: the error here is potentially nil
		return ErrFailedSCM{
			GitError: fmt.Errorf("failed to update membership for user %s in organization %s: %w", opt.Username, opt.Organization, err),
			Method:   "UpdateOrgMembership",
			Message:  fmt.Sprintf("failed to update membership for user %s", opt.Username),
		}
	}
	return nil
}

// RemoveMember implements the SCM interface
func (s *GithubSCM) RemoveMember(ctx context.Context, opt *OrgMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RemoveMember",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	// remove user from the organization and all teams
	_, err := s.client.Organizations.RemoveMember(ctx, opt.Organization, opt.Username)
	if err != nil {
		return ErrFailedSCM{
			Method:   "RemoveMember",
			GitError: fmt.Errorf("failed to remove user %s from organization %s: %w", opt.Username, opt.Organization, err),
			Message:  fmt.Sprintf("failed to remove user %s from the organization", opt.Username),
		}
	}
	return nil
}

// CreateIssue implements the SCM interface
func (s *GithubSCM) CreateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "CreateIssue",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	newIssue := &github.IssueRequest{
		Title:     &opt.Title,
		Body:      &opt.Body,
		Assignee:  opt.Assignee,
		Assignees: opt.Assignees,
	}

	s.logger.Debugf("Creating issue %q on %s", opt.Title, opt.Repository)
	issue, _, err := s.client.Issues.Create(ctx, opt.Organization, opt.Repository, newIssue)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "CreateIssue",
			Message:  fmt.Sprintf("failed to create issue %q", opt.Title),
			GitError: err,
		}
	}
	s.logger.Debugf("Created issue %q", opt.Title)

	return toIssue(issue), nil
}

// UpdateIssue implements the SCM interface
func (s *GithubSCM) UpdateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "UpdateIssue",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	issueReq := &github.IssueRequest{
		Title:     &opt.Title,
		Body:      &opt.Body,
		State:     &opt.State,
		Assignee:  opt.Assignee,
		Assignees: opt.Assignees,
	}
	s.logger.Debugf("Updating issue %d on %s", opt.Number, opt.Repository)
	issue, _, err := s.client.Issues.Edit(ctx, opt.Organization, opt.Repository, opt.Number, issueReq)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "UpdateIssue",
			Message:  fmt.Sprintf("failed to update issue %d on %s/%s", opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	s.logger.Debugf("Updated issue number %d", opt.Number)
	return toIssue(issue), nil
}

// GetIssue implements the SCM interface
func (s *GithubSCM) GetIssue(ctx context.Context, opt *RepositoryOptions, number int) (*Issue, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "GetIssue",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	issue, _, err := s.client.Issues.Get(ctx, opt.Owner, opt.Path, number)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "GetIssue",
			Message:  fmt.Sprintf("failed to get issue %d", number),
			GitError: err,
		}
	}
	return toIssue(issue), nil
}

// GetIssues implements the SCM interface
func (s *GithubSCM) GetIssues(ctx context.Context, opt *RepositoryOptions) ([]*Issue, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "GetIssues",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	issueList, _, err := s.client.Issues.ListByRepo(ctx, opt.Owner, opt.Path, &github.IssueListByRepoOptions{})
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "GetIssues",
			Message:  fmt.Sprintf("failed to get issues for %s", opt.Path),
			GitError: err,
		}
	}
	var issues []*Issue
	for _, issue := range issueList {
		issues = append(issues, toIssue(issue))
	}

	return issues, nil
}

// RequestReviewers implements the SCM interface
func (s *GithubSCM) RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RequestReviewers",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	reviewersRequest := github.ReviewersRequest{
		Reviewers: opt.Reviewers,
	}
	if _, _, err := s.client.PullRequests.RequestReviewers(ctx, opt.Organization, opt.Repository, opt.Number, reviewersRequest); err != nil {
		return ErrFailedSCM{
			Method:   "RequestReviewers",
			Message:  fmt.Sprintf("failed to request reviewers for pull request #%d on %s/%s", opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	return nil
}

// CreateIssueComment implements the SCM interface
func (s *GithubSCM) CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error) {
	if !opt.valid() {
		return 0, ErrMissingFields{
			Method:  "CreateIssueComment",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	createdComment, _, err := s.client.Issues.CreateComment(ctx, opt.Organization, opt.Repository, opt.Number, &github.IssueComment{Body: &opt.Body})
	if err != nil {
		return 0, ErrFailedSCM{
			Method:   "CreateIssueComment",
			Message:  fmt.Sprintf("failed to create comment for issue #%d, in repository: %s, for organization: %s", opt.Number, opt.Repository, opt.Organization),
			GitError: err,
		}
	}
	return createdComment.GetID(), nil
}

// UpdateIssueComment implements the SCM interface
func (s *GithubSCM) UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateIssueComment",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	if _, _, err := s.client.Issues.EditComment(ctx, opt.Organization, opt.Repository, opt.CommentID, &github.IssueComment{Body: &opt.Body}); err != nil {
		return ErrFailedSCM{
			Method:   "UpdateIssueComment",
			Message:  fmt.Sprintf("failed to edit comment in repository: %s, for organization: %s", opt.Repository, opt.Organization),
			GitError: err,
		}
	}
	return nil
}

func toRepository(repo *github.Repository) *Repository {
	return &Repository{
		ID:      uint64(repo.GetID()),
		Path:    repo.GetName(),
		Owner:   repo.Owner.GetLogin(),
		HTMLURL: repo.GetHTMLURL(),
		OrgID:   uint64(repo.Organization.GetID()),
		Size:    uint64(repo.GetSize()),
	}
}

func toIssue(issue *github.Issue) *Issue {
	return &Issue{
		ID:         uint64(issue.GetID()),
		Title:      issue.GetTitle(),
		Body:       issue.GetBody(),
		Repository: issue.Repository.GetName(),
		Assignee:   issue.Assignee.GetName(),
		Number:     issue.GetNumber(),
		Status:     issue.GetState(),
	}
}
