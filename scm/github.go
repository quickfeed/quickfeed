package scm

import (
	"context"
	"log"

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
		return nil, err
	}
	for _, user := range opt.Users {
		_, _, err = s.client.Teams.AddTeamMembership(ctx, t.GetID(), user, nil)
		if err != nil {
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
func (s *GithubSCM) DeleteTeam(ctx context.Context, teamID uint64) error {

	if _, err := s.client.Teams.DeleteTeam(ctx, int64(teamID)); err != nil {
		return err
	}
	return nil
}

// GetTeams implements the scm interface
func (s *GithubSCM) GetTeams(ctx context.Context, org *pb.Organization) ([]*Team, error) {
	gitTeams, _, err := s.client.Teams.ListTeams(ctx, org.Path, &github.ListOptions{})
	if err != nil {
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
	// if no id provided get it from github by slug
	if opt.TeamID < 1 {
		team, _, err := s.client.Teams.GetTeamBySlug(ctx, opt.Organization.Path, opt.TeamSlug)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		opt.TeamID = team.GetID()
	}

	isAlreadyMember, _, err := s.client.Teams.IsTeamMember(ctx, opt.TeamID, opt.Username)
	if err != nil {
		return err
	}
	if isAlreadyMember {
		return nil
	}
	_, _, err = s.client.Teams.AddTeamMembership(ctx, opt.TeamID, opt.Username, &github.TeamAddTeamMembershipOptions{})
	return err
}

// RemoveTeamMember implements the scm interface
func (s *GithubSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	// if no id provided get it from github by slug
	if opt.TeamID < 1 {
		team, _, err := s.client.Teams.GetTeamBySlug(ctx, opt.Organization.Path, opt.TeamSlug)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		opt.TeamID = team.GetID()
	}

	isMember, _, err := s.client.Teams.IsTeamMember(ctx, opt.TeamID, opt.Username)
	if err != nil {
		return err
	}
	if !isMember {
		return nil
	}

	_, err = s.client.Teams.RemoveTeamMembership(ctx, opt.TeamID, opt.Username)

	return err
}

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *CreateTeamOptions) error {
	groupTeam, _, err := s.client.Teams.GetTeam(ctx, int64(opt.TeamID))
	if err != nil {
		return err
	}

	// check whether group members are already in team; add missing members
	for _, member := range opt.Users {
		isMember, _, err := s.client.Teams.IsTeamMember(ctx, groupTeam.GetID(), member)
		if err != nil {
			log.Println("GitHub UpdateTeamMembers could not check user against the team: ", err.Error())
			return err
		}
		if !isMember {
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

	log.Println("scms: GetOrgMembership started with options: ", opt)
	gitOrg, _, err := s.client.Organizations.GetByID(ctx, int64(opt.OrgID))
	if err != nil {
		log.Println("scms: GetOrgMembership could not get organization: ", err.Error())
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
		log.Println("scms: UpdateOrgMembership could not get org: ", err.Error())
		return err
	}

	isMember, _, err := s.client.Organizations.IsMember(ctx, gitOrg.GetLogin(), opt.Username)
	if err != nil {
		log.Println("scms: UpdateOrgMembership could not check if member: ", err.Error())
		return err
	}
	if !isMember {
		return status.Errorf(codes.NotFound, "membership not found")
	}
	gitMembership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, gitOrg.GetLogin())
	if err != nil {
		log.Println("scms: UpdateOrgMembership could not get org membership: ", err.Error())
		return err
	}
	if opt.Role != "admin" && opt.Role != "member" {
		return status.Errorf(codes.InvalidArgument, "invalid role")
	}
	gitMembership.Role = &opt.Role
	newMembership, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.Username, gitOrg.GetLogin(), gitMembership)
	if err != nil {
		log.Println("scms: UpdateOrgMembership could not edit org membership: ", err.Error())
		return err
	}
	if newMembership.GetRole() != opt.Role {
		return status.Errorf(codes.Canceled, "failed to update membership")
	}
	return nil
}

// CreateOrgMembership implements the SCM interface
func (s *GithubSCM) CreateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	log.Println("scms: CreateOrgMembership startedwith options: ", opt)
	// check that organization is valid
	gitOrg, _, err := s.client.Organizations.GetByID(ctx, int64(opt.Organization.ID))
	if err != nil {
		log.Println("scms: CreateOrgMembership could not get org: ", err.Error())
		return err
	}
	// check that user is not already org member
	isMember, _, err := s.client.Organizations.IsMember(ctx, gitOrg.GetLogin(), opt.Username)
	if err != nil {
		log.Println("scms: CreateOrgMembership could not check if member: ", err.Error())
		return err
	}
	// if not member - issue an invitation
	if !isMember {
		// we can invite by github login (recommended, provided on user authentication with github, i.e. valid),
		// or by mail (provided by student, not necessary connected to github acount)
		invitation := github.CreateOrgInvitationOptions{}
		// if login is set - use it to get user's GitHub ID
		if opt.Username != "" {
			gitUser, _, err := s.client.Users.Get(ctx, opt.Username)
			if err != nil {
				log.Println("scms: CreateOrgMembership could not get user ", opt.Username)
				return status.Errorf(codes.InvalidArgument, "github user not found")
			}
			invitation.InviteeID = gitUser.ID
			log.Println("scms: CreateOrgMembership will use user ID: ", invitation.GetInviteeID())
		} else {
			// if no username and no email provided, method must fail
			if opt.Email == "" {
				log.Println("scms: CreateOrgMembership got neither username nor email")
				return status.Errorf(codes.InvalidArgument, "to invite user provide username or email")
			}
			invitation.Email = &opt.Email
		}
		// we want to use default values for other option fields (or we will get null field errors from github)
		role := "direct_member"
		invitation.Role = &role
		invitation.TeamID = make([]int64, 0)
		// issue an invitation. Github wil use user ID if provided and send invitation to account email,
		// otherwise will use email provided in options
		inv, _, err := s.client.Organizations.CreateOrgInvitation(ctx, gitOrg.GetLogin(), &invitation)
		if err != nil {
			log.Println("scms: CreateOrgMembership could not create invitation: ", err.Error())
			return status.Errorf(codes.Internal, "could not issue github invitation")
		}
		log.Println("scms: CreateOrgMembership sent invitation to user ", inv.GetLogin())
	}
	return nil
}
