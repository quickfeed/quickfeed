package scm

import (
	"context"
	"strconv"

	pb "github.com/autograde/quickfeed/ag"
	gitlab "github.com/xanzy/go-gitlab"
)

// GitlabSCM implements the SCM interface.
type GitlabSCM struct {
	client *gitlab.Client
}

// NewGitlabSCMClient returns a new GitLab client implementing the SCM interface.
func NewGitlabSCMClient(token string) *GitlabSCM {
	cli, _ := gitlab.NewOAuthClient(token, gitlab.WithoutRetries())
	return &GitlabSCM{
		client: cli,
	}
}

// CreateOrganization implements the SCM interface.
func (s *GitlabSCM) CreateOrganization(ctx context.Context, opt *OrganizationOptions) (*pb.Organization, error) {
	group, _, err := s.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:       &opt.Name,
		Path:       &opt.Path,
		Visibility: getVisibilityLevel(false),
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

// UpdateOrganization implements the SCM interface.
func (s *GitlabSCM) UpdateOrganization(ctx context.Context, opt *OrganizationOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "UpdateOrganization",
	}
}

// GetOrganization implements the SCM interface.
func (s *GitlabSCM) GetOrganization(ctx context.Context, opt *GetOrgOptions) (*pb.Organization, error) {
	group, _, err := s.client.Groups.GetGroup(strconv.FormatUint(opt.ID, 10), gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &pb.Organization{
		ID:     uint64(group.ID),
		Path:   group.Path,
		Avatar: group.AvatarURL,
	}, nil
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

// GetRepository implements the SCM interface.
func (s *GitlabSCM) GetRepository(cts context.Context, opt *RepositoryOptions) (*Repository, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "gitlab",
		Method: "GetRepository",
	}
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
func (s *GitlabSCM) DeleteRepository(ctx context.Context, opt *RepositoryOptions) (err error) {
	_, err = s.client.Projects.DeleteProject(strconv.FormatUint(opt.ID, 10), gitlab.WithContext(ctx))
	return
}

// UpdateRepoAccess implements the SCM interface.
func (s *GitlabSCM) UpdateRepoAccess(ctx context.Context, repo *Repository, user, permission string) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "UpdateRepoAccess",
	}
}

// RepositoryIsEmpty implements the SCM interface
func (s *GitlabSCM) RepositoryIsEmpty(ctx context.Context, opt *RepositoryOptions) bool {
	// TODO no implementation provided yet
	return false
}

// ListHooks implements the SCM interface.
func (s *GitlabSCM) ListHooks(ctx context.Context, repo *Repository, org string) ([]*Hook, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "gitlab",
		Method: "ListHooks",
	}
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
func (s *GitlabSCM) CreateTeam(ctx context.Context, opt *NewTeamOptions) (*Team, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "gitlab",
		Method: "CreateTeam",
	}
}

// DeleteTeam implements the SCM interface.
func (s *GitlabSCM) DeleteTeam(ctx context.Context, opt *TeamOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "DeleteTeam",
	}
}

// GetTeam implements the SCM interface
func (s *GitlabSCM) GetTeam(ctx context.Context, opt *TeamOptions) (*Team, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "gitlab",
		Method: "GetTeam",
	}
}

// GetTeams implements the SCM interface
func (s *GitlabSCM) GetTeams(ctx context.Context, org *pb.Organization) ([]*Team, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "gitlab",
		Method: "GetTeam",
	}
}

// AddTeamMember implements the scm interface
func (s *GitlabSCM) AddTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "AddTeamMember",
	}
}

// RemoveTeamMember implements the scm interface
func (s *GitlabSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "RemoveTeamMember",
	}
}

// UpdateTeamMembers implements the SCM interface
func (s *GitlabSCM) UpdateTeamMembers(context.Context, *UpdateTeamOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "UpdateTeamMembers",
	}
}

// AddTeamRepo implements the SCM interface.
func (s *GitlabSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "AddTeamRepo",
	}
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

func getVisibilityLevel(private bool) *gitlab.VisibilityValue {
	if private {
		return gitlab.Visibility(gitlab.PrivateVisibility)
	}
	return gitlab.Visibility(gitlab.PublicVisibility)
}

// UpdateOrgMembership implements the SCM interface
func (s *GitlabSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "UpdateOrgMembership",
	}
}

// RemoveMember implements the SCM interface
func (s *GitlabSCM) RemoveMember(ctx context.Context, opt *OrgMembershipOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "gitlab",
		Method: "RemoveMember",
	}
}

// GetUserScopes implements the SCM interface
func (s *GitlabSCM) GetUserScopes(ctx context.Context) *Authorization {
	// TODO no implementation provided yet
	return nil
}

// GetFileContent implements the SCM interface
func (s *GitlabSCM) GetFileContent(context.Context, *FileOptions) (string, error) {
	// TODO no implementation provided yet
	return "", ErrNotSupported{
		SCM:    "gitlab",
		Method: "GetFileContent",
	}
}
