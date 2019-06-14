package scm

import (
	"context"
	"log"

	pb "github.com/autograde/aguis/ag"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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

// ListDirectories implements the SCM interface.
func (s *GithubSCM) ListDirectories(ctx context.Context) ([]*pb.Directory, error) {
	orgs, _, err := s.client.Organizations.ListOrgMemberships(ctx, nil)
	if err != nil {
		return nil, err
	}

	var directories []*pb.Directory
	for _, org := range orgs {
		directories = append(directories, &pb.Directory{
			ID:     uint64(org.Organization.GetID()),
			Path:   org.Organization.GetLogin(),
			Avatar: org.Organization.GetAvatarURL(),
		})
	}
	return directories, nil
}

// CreateDirectory implements the SCM interface.
func (s *GithubSCM) CreateDirectory(ctx context.Context, opt *CreateDirectoryOptions) (*pb.Directory, error) {
	return nil, ErrNotSupported{
		SCM:    "github",
		Method: "CreateDirectory",
	}
}

// GetDirectory implements the SCM interface.
func (s *GithubSCM) GetDirectory(ctx context.Context, id uint64) (*pb.Directory, error) {
	org, _, err := s.client.Organizations.GetByID(ctx, int(id))
	if err != nil {
		return nil, err
	}

	return &pb.Directory{
		ID:     uint64(org.GetID()),
		Path:   org.GetLogin(),
		Avatar: org.GetAvatarURL(),
	}, nil
}

// CreateRepoAndTeam implements the SCM interface.
func (s *GithubSCM) CreateRepoAndTeam(ctx context.Context, opt *CreateRepositoryOptions, teamName string, gitUserNames []string) (*Repository, *Team, error) {
	log.Println("scm: createRepoAndTeam starts")
	repo, err := s.CreateRepository(ctx, opt)
	if err != nil {
		log.Println("scm: createRepoAndTeam - error creating repo: ", err.Error())
		return nil, nil, err
	}

	team, err := s.CreateTeam(ctx, &CreateTeamOptions{
		Directory: opt.Directory,
		TeamName:  teamName,
		Users:     gitUserNames,
	})
	if err != nil {
		log.Println("scm: createRepoAndTeam - error creating team: ", err.Error())
		return nil, nil, err
	}

	err = s.AddTeamRepo(ctx, &AddTeamRepoOptions{
		TeamID: team.ID,
		Owner:  repo.Owner,
		Repo:   repo.Path,
	})
	if err != nil {
		log.Println("scm: createRepoAndTeam - error adding team repo: ", err.Error())
		return nil, nil, err
	}
	return repo, team, nil
}

// CreateRepository implements the SCM interface.
func (s *GithubSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	repo, _, err := s.client.Repositories.Create(ctx, opt.Directory.Path, &github.Repository{
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
		ID:          uint64(repo.GetID()),
		Path:        repo.GetName(),
		Owner:       owner,
		WebURL:      repo.GetHTMLURL(),
		SSHURL:      repo.GetSSHURL(),
		HTTPURL:     repo.GetCloneURL(),
		DirectoryID: opt.Directory.ID,
	}, nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, directory *pb.Directory) ([]*Repository, error) {
	var path string
	if directory.Path != "" {
		path = directory.Path
	} else {
		directory, err := s.GetDirectory(ctx, directory.ID)
		if err != nil {
			return nil, err
		}
		path = directory.Path
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
			ID:          uint64(repo.GetID()),
			Path:        repo.GetName(),
			Owner:       owner,
			WebURL:      repo.GetHTMLURL(),
			SSHURL:      repo.GetSSHURL(),
			HTTPURL:     repo.GetCloneURL(),
			DirectoryID: directory.ID,
		})
	}

	return repositories, nil
}

// DeleteRepository implements the SCM interface.
func (s *GithubSCM) DeleteRepository(ctx context.Context, id uint64) error {
	repo, _, err := s.client.Repositories.GetByID(ctx, int(id))
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
			ID:   uint64(hook.GetID()),
			Name: hook.GetName(),
			URL:  hook.GetURL(),
		})
	}
	return hooks, err
}

const autograderHookName = "web"

// CreateHook implements the SCM interface.
func (s *GithubSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) (err error) {
	name := autograderHookName
	_, _, err = s.client.Repositories.CreateHook(ctx, opt.Repository.Owner, opt.Repository.Path, &github.Hook{
		Name: &name,
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
	t, _, err := s.client.Organizations.CreateTeam(ctx, opt.Directory.Path, &github.Team{
		Name: &opt.TeamName,
	})
	if err != nil {
		return nil, err
	}
	for _, user := range opt.Users {
		_, _, err = s.client.Organizations.AddTeamMembership(ctx, t.GetID(), user, nil)
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

	if _, err := s.client.Organizations.DeleteTeam(ctx, int(teamID)); err != nil {
		return err
	}
	return nil
}

// GetTeams implements the scm interface
func (s *GithubSCM) GetTeams(ctx context.Context, org *pb.Directory) ([]*Team, error) {
	gitTeams, _, err := s.client.Organizations.ListTeams(ctx, org.Path, &github.ListOptions{})
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

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *CreateTeamOptions) error {
	groupTeam, _, err := s.client.Organizations.GetTeam(ctx, int(opt.TeamID))
	if err != nil {
		return err
	}

	// check whether group members are already in team; add missing members
	for _, member := range opt.Users {
		isMember, _, err := s.client.Organizations.IsTeamMember(ctx, groupTeam.GetID(), member)
		if err != nil {
			log.Println("GitHub UpdateTeamMembers could not check user against the team: ", err.Error())
			return err
		}
		if !isMember {
			_, _, err = s.client.Organizations.AddTeamMembership(ctx, groupTeam.GetID(), member, nil)
			if err != nil {
				log.Println("GitHub UpdateTeamMembers could not add user ", member, " to the team ", groupTeam.GetName(), ": ", err.Error())
				return err
			}
		}
	}

	// find current team members
	oldUsers, _, err := s.client.Organizations.ListTeamMembers(ctx, groupTeam.GetID(), nil)
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
			_, err = s.client.Organizations.RemoveTeamMembership(ctx, groupTeam.GetID(), teamMember.GetLogin())
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
	return "https://" + token + "@github.com/" + opt.Directory + "/" + opt.Repository + ".git"
}

// AddTeamRepo implements the SCM interface.
func (s *GithubSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	_, err := s.client.Organizations.AddTeamRepo(ctx, int(opt.TeamID), opt.Owner, opt.Repo, &github.OrganizationAddTeamRepoOptions{
		Permission: "push", // This make sure that users can pull and push
	})
	if err != nil {
		return err
	}
	return nil
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
	user, _, err := s.client.Users.GetByID(ctx, int(remoteID))
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

// GetPaymentPlan implements the SCM interface.
func (s *GithubSCM) GetPaymentPlan(ctx context.Context, orgID uint64) (*PaymentPlan, error) {
	org, _, err := s.client.Organizations.GetByID(ctx, int(orgID))
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
	gitRepo, _, err := s.client.Repositories.GetByID(ctx, int(repo.ID))
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

	gitOrg, _, err := s.client.Organizations.GetByID(ctx, int(opt.OrgID))
	if err != nil {
		return nil, err
	}
	membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, gitOrg.GetName())
	if err != nil {
		return nil, err
	}
	opt.Role = membership.GetRole()

	return opt, nil
}
