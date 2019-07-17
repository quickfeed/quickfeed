package scm

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc/codes"

	pb "github.com/autograde/aguis/ag"
	"github.com/google/go-github/v26/github"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/status"
)

// GithubSCM implements the SCM interface.
type GithubSCM struct {
	client *github.Client
	token  string
}

// NewGithubSCMClient returns a new Github client implementing the SCM interface.
func NewGithubSCMClient(token string) *GithubSCM {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := github.NewClient(oauth2.NewClient(context.Background(), ts))
	return &GithubSCM{
		client: client,
		token:  token,
	}
}

// ListOrganizations implements the SCM interface.
func (s *GithubSCM) ListOrganizations(ctx context.Context) ([]*pb.Organization, error) {
	userOrgs, _, err := s.client.Organizations.ListOrgMemberships(ctx, nil)
	if err != nil {
		return nil, err
	}

	var orgs []*pb.Organization
	for _, org := range userOrgs {
		orgs = append(orgs, &pb.Organization{
			ID:     uint64(org.Organization.GetID()),
			Path:   org.Organization.GetLogin(),
			Avatar: org.Organization.GetAvatarURL(),
		})
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
		return nil, err
	}

	return &pb.Organization{
		ID:     uint64(org.GetID()),
		Path:   org.GetLogin(),
		Avatar: org.GetAvatarURL(),
	}, nil
}

// CreateRepoAndTeam implements the SCM interface.
func (s *GithubSCM) CreateRepoAndTeam(ctx context.Context, opt *CreateRepositoryOptions, teamName string, gitUserNames []string) (*Repository, *Team, error) {
	repo, err := s.CreateRepository(ctx, opt)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, err.Error())
	}

	team, err := s.CreateTeam(ctx, &CreateTeamOptions{
		Organization: opt.Organization,
		TeamName:     teamName,
		Users:        gitUserNames,
	})
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, err.Error())
	}

	err = s.AddTeamRepo(ctx, &AddTeamRepoOptions{
		TeamID: team.ID,
		Owner:  repo.Owner,
		Repo:   repo.Path,
	})
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, err.Error())
	}
	return repo, team, nil
}

// CreateRepository implements the SCM interface.
func (s *GithubSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	repo, _, err := s.client.Repositories.Create(ctx, opt.Organization.Path, &github.Repository{
		Name:    &opt.Path,
		Private: &opt.Private,
	})
	if err != nil {
		return nil, err
	}

	owner := ""
	if repo.Owner != nil {
		owner = repo.Owner.GetLogin()
	}

	return &Repository{
		ID:      uint64(repo.GetID()),
		Path:    repo.GetName(),
		Owner:   owner,
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
			return nil, err
		}
		path = org.Path
	}

	repos, _, err := s.client.Repositories.ListByOrg(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var repositories []*Repository
	for _, repo := range repos {

		owner := ""
		if repo.Owner != nil {
			owner = repo.Owner.GetLogin()
		}

		repositories = append(repositories, &Repository{
			ID:      uint64(repo.GetID()),
			Path:    repo.GetName(),
			Owner:   owner,
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
		return err
	}
	if _, err := s.client.Repositories.Delete(ctx, repo.Owner.GetLogin(), repo.GetName()); err != nil {
		return err
	}
	return nil
}

// ListHooks implements the SCM interface.
func (s *GithubSCM) ListHooks(ctx context.Context, repo *Repository) ([]*Hook, error) {
	githubHooks, _, err := s.client.Repositories.ListHooks(ctx, repo.Owner, repo.Path, nil)
	var hooks []*Hook
	for _, hook := range githubHooks {
		hooks = append(hooks, &Hook{
			ID:  uint64(hook.GetID()),
			URL: hook.GetURL(),
		})
	}
	return hooks, err
}

// CreateHook implements the SCM interface.
func (s *GithubSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) (err error) {
	_, _, err = s.client.Repositories.CreateHook(ctx, opt.Repository.Owner, opt.Repository.Path, &github.Hook{
		Config: map[string]interface{}{
			"url":          opt.URL,
			"secret":       opt.Secret,
			"content_type": "json",
			"insecure_ssl": "0",
		},
	})
	if err != nil {
		log.Println("GitHub SCM: CreateHook for repository ", opt.Repository.Path, " resulted in error: ", err.Error())
	}
	return
}

// CreateTeam implements the SCM interface.
func (s *GithubSCM) CreateTeam(ctx context.Context, opt *CreateTeamOptions) (*Team, error) {
	t, _, err := s.client.Teams.CreateTeam(ctx, opt.Organization.Path, github.NewTeam{
		Name: opt.TeamName,
	})
	if err != nil {
		log.Println("GitHub CreateTeam failed: ", err.Error())
		return nil, err
	}

	for _, user := range opt.Users {
		_, _, err = s.client.Teams.AddTeamMembership(ctx, t.GetID(), user, nil)
		if err != nil {
			log.Println("GitHub CreateTeam failed to add membership for user ", user, ": ", err.Error())
			return nil, err
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
		log.Println("GitHub DeleteTeam failed to get team: ", err.Error())
	}

	if _, err := s.client.Teams.DeleteTeam(ctx, int64(team.ID)); err != nil {
		log.Println("GitHub DeleteTeam failed: ", err.Error())
		return err
	}
	return nil
}

// GetTeam implements the SCM interface
func (s *GithubSCM) GetTeam(ctx context.Context, opt *CreateTeamOptions) (*Team, error) {
	if opt.TeamID < 1 {
		slug := strings.ToLower(opt.TeamName)
		team, _, err := s.client.Teams.GetTeamBySlug(ctx, opt.Organization.Path, slug)
		if err != nil {
			log.Println("GitHub GetTeam: could not get team by slug")
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		scmTeam := &Team{
			ID:   uint64(team.GetID()),
			Name: team.GetName(),
			URL:  team.GetURL(),
		}

		return scmTeam, nil
	}
	team, _, err := s.client.Teams.GetTeam(ctx, int64(opt.TeamID))
	if err != nil {
		log.Println("GitHub GetTeam: could not get team by ID")
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	scmTeam := &Team{
		ID:   uint64(team.GetID()),
		Name: team.GetName(),
		URL:  team.GetURL(),
	}
	return scmTeam, nil
}

// GetTeams implements the scm interface
func (s *GithubSCM) GetTeams(ctx context.Context, org *pb.Organization) ([]*Team, error) {
	gitTeams, _, err := s.client.Teams.ListTeams(ctx, org.Path, &github.ListOptions{})
	if err != nil {
		log.Println("GitHub GetTeams: failed to list teams ")
		return nil, err
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
	team, err := s.GetTeam(ctx, &CreateTeamOptions{TeamName: opt.TeamSlug, TeamID: uint64(opt.TeamID)})
	if err != nil {
		log.Println("GitHub AddTeamMember failed to get team: ", err.Error())
	}

	isAlreadyMember, _, err := s.client.Teams.GetTeamMembership(ctx, int64(team.ID), opt.Username)
	if err != nil {
		// will always return an error when user is not a team member, but this is expected, no error will be returned
		// but it is useful to log it in case there were other reasons (invalid token and others)
		log.Println("GitHub AddTeamMember: team membership not found: ", err.Error())
	}
	// if already in team , take no action
	if isAlreadyMember != nil {
		log.Println("GitHub adding team member: user ", opt.Username, " is already in the team ", opt.TeamSlug)
		return nil
	}
	// otherwise add user as team member
	_, _, err = s.client.Teams.AddTeamMembership(ctx, int64(team.ID), opt.Username, &github.TeamAddTeamMembershipOptions{Role: opt.Role})
	return err
}

// RemoveTeamMember implements the scm interface
func (s *GithubSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	team, err := s.GetTeam(ctx, &CreateTeamOptions{TeamName: opt.TeamSlug, TeamID: uint64(opt.TeamID)})
	if err != nil {
		log.Println("GitHub RemoveTeamMember failed to get team: ", err.Error())
	}

	isMember, _, err := s.client.Teams.GetTeamMembership(ctx, int64(team.ID), opt.Username)
	if isMember == nil {
		log.Println("GitHub removing team member: user ", opt.Username, " is not a member of team ", opt.TeamSlug)
		// user is not in this team
		return nil
	}
	// TODO(vera): check for errors other than not found

	_, err = s.client.Teams.RemoveTeamMembership(ctx, opt.TeamID, opt.Username)
	return err
}

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *CreateTeamOptions) error {
	groupTeam, _, err := s.client.Teams.GetTeam(ctx, int64(opt.TeamID))
	if err != nil {
		log.Println("GitHub UpdateTeamMember: failed to get team: ", err.Error())
		return err
	}

	// check whether group members are already in team; add missing members
	for _, member := range opt.Users {
		isMember, _, err := s.client.Teams.GetTeamMembership(ctx, groupTeam.GetID(), member)
		// TODO(vera): error check (other than not found)
		if isMember == nil {
			_, _, err = s.client.Teams.AddTeamMembership(ctx, groupTeam.GetID(), member, nil)
			if err != nil {
				log.Println("GitHub UpdateTeamMembers could not add user ", member, " to the team ", groupTeam.GetName(), ": ", err.Error())
				return err
			}
		}
	}

	// find current team members
	oldUsers, _, err := s.client.Teams.ListTeamMembers(ctx, groupTeam.GetID(), nil)
	if err != nil {
		log.Println("GitHub UpdateTeamMembers could not list team members: ", err.Error())
		return err
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
				log.Println("GitHub UpdateTeamMembers could not remove user ", teamMember.GetLogin(), " from team ", groupTeam.GetName(), ": ", err.Error())
				return err
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
	_, err := s.client.Teams.AddTeamRepo(ctx, int64(opt.TeamID), opt.Owner, opt.Repo, &github.TeamAddTeamRepoOptions{
		Permission: "push", // This make sure that users can pull and push
	})
	return err
}

// GetUserName implements the SCM interface.
func (s *GithubSCM) GetUserName(ctx context.Context) (string, error) {
	user, _, err := s.client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

// GetUserNameByID implements the SCM interface.
func (s *GithubSCM) GetUserNameByID(ctx context.Context, remoteID uint64) (string, error) {
	user, _, err := s.client.Users.GetByID(ctx, int64(remoteID))
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

// GetPaymentPlan implements the SCM interface.
func (s *GithubSCM) GetPaymentPlan(ctx context.Context, orgID uint64) (*PaymentPlan, error) {
	org, _, err := s.client.Organizations.GetByID(ctx, int64(orgID))
	if err != nil {
		return nil, err
	}
	plan := &PaymentPlan{
		Name:         org.Plan.GetName(),
		PrivateRepos: uint64(org.Plan.GetPrivateRepos()),
	}
	return plan, nil
}

// UpdateRepository implements the SCM interface
func (s *GithubSCM) UpdateRepository(ctx context.Context, repo *Repository) error {
	// TODO - make this more flexible rather than only making stuff private.
	gitRepo, _, err := s.client.Repositories.GetByID(ctx, int64(repo.ID))
	if err != nil {
		return err
	}

	*gitRepo.Private = true
	_, _, err = s.client.Repositories.Edit(ctx, gitRepo.Owner.GetLogin(), gitRepo.GetName(), gitRepo)
	if err != nil {
		return err
	}

	return nil
}

// GetOrgMembership implements the SCM interface
func (s *GithubSCM) GetOrgMembership(ctx context.Context, opt *OrgMembership) (*OrgMembership, error) {

	gitOrg, _, err := s.client.Organizations.GetByID(ctx, int64(opt.OrgID))
	if err != nil {
		log.Println("GitHub GetOrgMembership could not get organization: ", err.Error())
		return nil, err
	}

	membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, gitOrg.GetLogin())
	if err != nil {
		log.Println("scms: GetOrgMembership could not get membership: ", err.Error())
		return nil, err
	}
	opt.Role = membership.GetRole()

	return opt, nil
}

// UpdateOrgMembership implements the SCM interface
func (s *GithubSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembership) error {
	log.Println("scms: UpdateOrgMembership startedwith options: ", opt)
	gitOrg, _, err := s.client.Organizations.GetByID(ctx, int64(opt.OrgID))
	if err != nil {
		log.Println("GitHub UpdateOrgMembership could not get org: ", err.Error())
		return err
	}

	isMember, _, err := s.client.Organizations.IsMember(ctx, gitOrg.GetLogin(), opt.Username)
	if err != nil {
		log.Println("GitHub UpdateOrgMembership could not check if member: ", err.Error())
		return err
	}
	if !isMember {
		return status.Errorf(codes.NotFound, "membership not found")
	}
	gitMembership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, gitOrg.GetLogin())
	if err != nil {
		log.Println("GitHub UpdateOrgMembership could not get org membership: ", err.Error())
		return err
	}
	if opt.Role != "admin" && opt.Role != "member" {
		return status.Errorf(codes.InvalidArgument, "invalid role")
	}
	gitMembership.Role = &opt.Role
	newMembership, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.Username, gitOrg.GetLogin(), gitMembership)
	if err != nil {
		log.Println("GitHub UpdateOrgMembership could not edit org membership: ", err.Error())
		return err
	}
	if newMembership.GetRole() != opt.Role {
		return status.Errorf(codes.Canceled, "failed to update membership")
	}
	return nil
}

// GetUserScopes implements the SCM interface
func (s *GithubSCM) GetUserScopes(ctx context.Context) *Authorization {
	// this method will always return error as it is OAut2 API, but will also return scopes info in response headers
	_, resp, _ := s.client.Authorizations.List(ctx, &github.ListOptions{})
	// header contains a single string with all scopes for authenticated user
	stringScopes := resp.Header.Get("X-OAuth-Scopes")
	// we split the string to check against the global slice of required scopes
	gitScopes := strings.Split(stringScopes, ", ")
	return &Authorization{Scopes: gitScopes}
}
