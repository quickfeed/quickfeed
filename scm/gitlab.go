package scm

import (
	"context"
	"strconv"

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

// ListDirectories implements the SCM interface.
func (s *GitlabSCM) ListDirectories(ctx context.Context) ([]*Directory, error) {
	groups, _, err := s.client.Groups.ListGroups(&gitlab.ListGroupsOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var directories []*Directory
	for _, group := range groups {
		directories = append(directories, &Directory{
			ID:     uint64(group.ID),
			Path:   group.Path,
			Avatar: group.AvatarURL,
		})
	}
	return directories, nil
}

// CreateDirectory implements the SCM interface.
func (s *GitlabSCM) CreateDirectory(ctx context.Context, opt *CreateDirectoryOptions) (*Directory, error) {
	group, _, err := s.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:            &opt.Name,
		Path:            &opt.Path,
		VisibilityLevel: getVisibilityLevel(false),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &Directory{
		ID:     uint64(group.ID),
		Path:   group.Path,
		Avatar: group.AvatarURL,
	}, nil
}

// GetDirectory implements the SCM interface.
func (s *GitlabSCM) GetDirectory(ctx context.Context, id uint64) (*Directory, error) {
	group, _, err := s.client.Groups.GetGroup(strconv.FormatUint(id, 10), gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &Directory{
		ID:     uint64(group.ID),
		Path:   group.Path,
		Avatar: group.AvatarURL,
	}, nil
}

// CreateRepository implements the SCM interface.
func (s *GitlabSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	directoryID := int(opt.Directory.ID)
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
		ID:          uint64(repo.ID),
		Path:        repo.Path,
		WebURL:      repo.WebURL,
		SSHURL:      repo.SSHURLToRepo,
		HTTPURL:     repo.HTTPURLToRepo,
		DirectoryID: opt.Directory.ID,
	}, nil
}

func getVisibilityLevel(private bool) *gitlab.VisibilityLevelValue {
	if private {
		return gitlab.VisibilityLevel(gitlab.PrivateVisibility)
	}
	return gitlab.VisibilityLevel(gitlab.PublicVisibility)
}
