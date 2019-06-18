package scm

import (
	"context"
	"strconv"

	pb "github.com/autograde/aguis/ag"
	gitlab "github.com/xanzy/go-gitlab"
)

// GitlabSCM implements the SCM interface.
type GitlabSCM struct {
	client *gitlab.Client
}

// NewGitlabSCMClient returns a new GitLab client implementing the SCM interface.
func NewGitlabSCMClient(token string) *GitlabSCM {
	return &GitlabSCM{
		client: gitlab.NewOAuthClient(nil, token),
	}
}

// ListOrganizations implements the SCM interface.
func (s *GitlabSCM) ListOrganizations(ctx context.Context) ([]*pb.Organization, error) {
	groups, _, err := s.client.Groups.ListGroups(nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var directories []*pb.Organization
	for _, group := range groups {
		directories = append(directories, &pb.Organization{
			ID:     uint64(group.ID),
			Path:   group.Path,
			Avatar: group.AvatarURL,
		})
	}
	return directories, nil
}

// CreateOrganization implements the SCM interface.
func (s *GitlabSCM) CreateOrganization(ctx context.Context, opt *CreateOrgOptions) (*pb.Organization, error) {
	group, _, err := s.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:            &opt.Name,
		Path:            &opt.Path,
		VisibilityLevel: getVisibilityLevel(false),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &pb.Organization{
		ID:     uint64(group.ID),
		Path:   group.Path,
		Avatar: group.AvatarURL,
	}, nil
}

// GetOrganization implements the SCM interface.
func (s *GitlabSCM) GetOrganization(ctx context.Context, id uint64) (*pb.Organization, error) {
	group, _, err := s.client.Groups.GetGroup(strconv.FormatUint(id, 10), gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &pb.Organization{
		ID:     uint64(group.ID),
		Path:   group.Path,
		Avatar: group.AvatarURL,
	}, nil
}

// CreateRepoAndTeam implements the SCM interface.
func (s *GitlabSCM) CreateRepoAndTeam(ctx context.Context, opt *CreateRepositoryOptions, teamName string, gitUserNames []string) (*Repository, *Team, error) {
	// TODO no implementation provided yet
	return nil, nil, nil
}

// CreateRepository implements the SCM interface.
func (s *GitlabSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	directoryID := int(opt.Organization.ID)
	repo, _, err := s.client.Projects.CreateProject(
		&gitlab.CreateProjectOptions{
			Path:        &opt.Path,
			NamespaceID: &directoryID,
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	return &Repository{
		ID:      uint64(repo.ID),
		Path:    repo.Path,
		WebURL:  repo.WebURL,
		SSHURL:  repo.SSHURLToRepo,
		HTTPURL: repo.HTTPURLToRepo,
		OrgID:   opt.Organization.ID,
	}, nil
}

// GetRepositories implements the SCM interface.
func (s *GitlabSCM) GetRepositories(ctx context.Context, directory *pb.Organization) ([]*Repository, error) {
	var gid interface{}
	if directory.Path != "" {
		gid = directory.Path
	} else {
		gid = strconv.FormatUint(directory.ID, 10)
	}

	repos, _, err := s.client.Groups.ListGroupProjects(gid, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var repositories []*Repository
	for _, repo := range repos {
		repositories = append(repositories, &Repository{
			ID:      uint64(repo.ID),
			Path:    repo.Path,
			WebURL:  repo.WebURL,
			SSHURL:  repo.SSHURLToRepo,
			HTTPURL: repo.HTTPURLToRepo,
			OrgID:   directory.ID,
		})
	}

	return repositories, nil
}

// DeleteRepository implements the SCM interface.
func (s *GitlabSCM) DeleteRepository(ctx context.Context, id uint64) (err error) {
	_, err = s.client.Projects.DeleteProject(strconv.FormatUint(id, 10), gitlab.WithContext(ctx))
	return
}

// ListHooks implements the SCM interface.
func (s *GitlabSCM) ListHooks(ctx context.Context, repo *Repository) ([]*Hook, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// CreateHook implements the SCM interface.
func (s *GitlabSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) (err error) {
	_, _, err = s.client.Projects.AddProjectHook(strconv.FormatUint(opt.Repository.ID, 10), &gitlab.AddProjectHookOptions{
		URL:   &opt.URL,
		Token: &opt.Secret,
	}, gitlab.WithContext(ctx))
	return
}

// CreateTeam implements the SCM interface.
func (s *GitlabSCM) CreateTeam(ctx context.Context, opt *CreateTeamOptions) (*Team, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// DeleteTeam implements the SCM interface.
func (s *GitlabSCM) DeleteTeam(ctx context.Context, teamID uint64) error {
	// TODO no implementation provided yet
	return nil
}

// GetTeams implements the SCM interface
func (s *GitlabSCM) GetTeams(ctx context.Context, org *pb.Organization) ([]*Team, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// AddTeamMember implements the scm interface
func (s *GitlabSCM) AddTeamMember(ctx context.Context, opt *AddMemberOptions) error {
	// TODO no implementation provided yet
	return nil
}

// UpdateTeamMembers implements the SCM interface
func (s *GitlabSCM) UpdateTeamMembers(context.Context, *CreateTeamOptions) error {
	// TODO no implementation provided yet
	return nil
}

// AddTeamRepo implements the SCM interface.
func (s *GitlabSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	// TODO no implementation provided yet
	return nil
}

// GetUserName implements the SCM interface.
func (s *GitlabSCM) GetUserName(ctx context.Context) (string, error) {
	return "", nil
}

// GetUserNameByID implements the SCM interface.
func (s *GitlabSCM) GetUserNameByID(ctx context.Context, remoteID uint64) (string, error) {
	return "", nil
}

// CreateCloneURL implements the SCM interface.
func (s *GitlabSCM) CreateCloneURL(opt *CreateClonePathOptions) string {
	return ""
}

// GetPaymentPlan implements the SCM interface.
func (s *GitlabSCM) GetPaymentPlan(ctx context.Context, orgID uint64) (*PaymentPlan, error) {
	return nil, nil
}

// UpdateRepository implements the SCM interface.
func (s *GitlabSCM) UpdateRepository(ctx context.Context, repo *Repository) error {
	return nil
}

// GetOrgMembership implements the SCM interface
func (s *GitlabSCM) GetOrgMembership(ctx context.Context, opt *OrgMembership) (*OrgMembership, error) {
	// TODO no implementation provided yet
	return nil, nil
}

func getVisibilityLevel(private bool) *gitlab.VisibilityLevelValue {
	if private {
		return gitlab.VisibilityLevel(gitlab.PrivateVisibility)
	}
	return gitlab.VisibilityLevel(gitlab.PublicVisibility)
}

// UpdateOrgMembership implements the SCM interface
func (s *GitlabSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembership) error {
	// TODO no implementation provided yet
	return nil
}
