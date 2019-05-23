package scm

import (
	"context"
	"errors"
	"strconv"

	pb "github.com/autograde/aguis/ag"
)

// FakeSCM implements the SCM interface.
type FakeSCM struct {
	Repositories map[uint64]*Repository
	Directories  map[uint64]*pb.Directory
	Hooks        map[uint64]int
}

// NewFakeSCMClient returns a new Fake client implementing the SCM interface.
func NewFakeSCMClient() *FakeSCM {
	return &FakeSCM{
		Repositories: make(map[uint64]*Repository),
		Directories:  make(map[uint64]*pb.Directory),
		Hooks:        make(map[uint64]int),
	}
}

// ListDirectories implements the SCM interface.
func (s *FakeSCM) ListDirectories(ctx context.Context) ([]*pb.Directory, error) {
	var dirs []*pb.Directory
	for _, dir := range s.Directories {
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

// CreateDirectory implements the SCM interface.
func (s *FakeSCM) CreateDirectory(ctx context.Context, opt *CreateDirectoryOptions) (*pb.Directory, error) {
	id := len(s.Directories) + 1
	dir := &pb.Directory{
		Id:     uint64(id),
		Path:   opt.Path,
		Avatar: "https://avatars3.githubusercontent.com/u/1000" + strconv.Itoa(id) + "?v=3",
	}
	s.Directories[dir.Id] = dir
	return dir, nil
}

// GetDirectory implements the SCM interface.
func (s *FakeSCM) GetDirectory(ctx context.Context, id uint64) (*pb.Directory, error) {
	dir, ok := s.Directories[id]
	if !ok {
		return nil, errors.New("directory not found")
	}
	return dir, nil
}

// CreateRepoAndTeam implements the SCM interface.
func (s *FakeSCM) CreateRepoAndTeam(ctx context.Context, opt *CreateRepositoryOptions, teamName string, gitUserNames []string) (*Repository, error) {
	repo, err := s.CreateRepository(ctx, opt)
	if err != nil {
		return nil, err
	}

	team, err := s.CreateTeam(ctx, &CreateTeamOptions{
		Directory: opt.Directory,
		TeamName:  teamName,
		Users:     gitUserNames,
	})
	if err != nil {
		return nil, err
	}

	err = s.AddTeamRepo(ctx, &AddTeamRepoOptions{
		TeamID: team.ID,
		Owner:  repo.Owner,
		Repo:   repo.Path,
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// CreateRepository implements the SCM interface.
func (s *FakeSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	id := len(s.Repositories) + 1
	repo := &Repository{
		ID:          uint64(id),
		Path:        opt.Path,
		WebURL:      "https://example.com/" + opt.Directory.Path + "/" + opt.Path,
		SSHURL:      "git@example.com:" + opt.Directory.Path + "/" + opt.Path,
		HTTPURL:     "https://example.com/" + opt.Directory.Path + "/" + opt.Path + ".git",
		DirectoryID: opt.Directory.Id,
	}
	s.Repositories[repo.ID] = repo
	return repo, nil
}

// GetRepositories implements the SCM interface.
func (s *FakeSCM) GetRepositories(ctx context.Context, directory *pb.Directory) ([]*Repository, error) {
	var repos []*Repository
	for _, repo := range s.Repositories {
		if repo.DirectoryID == directory.Id {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// DeleteRepository implements the SCM interface.
func (s *FakeSCM) DeleteRepository(ctx context.Context, id uint64) error {
	if _, ok := s.Repositories[id]; !ok {
		return errors.New("repository not found")
	}
	delete(s.Repositories, id)
	return nil
}

// ListHooks implements the SCM interface.
func (s *FakeSCM) ListHooks(ctx context.Context, repo *Repository) ([]*Hook, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// CreateHook implements the SCM interface.
func (s *FakeSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) error {
	if _, ok := s.Repositories[opt.Repository.ID]; !ok {
		return errors.New("repository not found")
	}
	s.Hooks[opt.Repository.ID]++
	return nil
}

// CreateTeam implements the SCM interface.
func (s *FakeSCM) CreateTeam(ctx context.Context, opt *CreateTeamOptions) (*Team, error) {
	// TODO no implementation provided yet
	return &Team{ID: 1, Name: "", URL: ""}, nil
}

// CreateCloneURL implements the SCM interface.
func (s *FakeSCM) CreateCloneURL(opt *CreateClonePathOptions) string {
	return ""
}

// AddTeamRepo implements the SCM interface.
func (s *FakeSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	return nil
}

// GetUserName implements the SCM interface.
func (s *FakeSCM) GetUserName(ctx context.Context) (string, error) {
	return "", nil
}

// GetUserNameByID implements the SCM interface.
func (s *FakeSCM) GetUserNameByID(ctx context.Context, remoteID uint64) (string, error) {
	return "", nil
}

// GetPaymentPlan implements the SCM interface.
func (s *FakeSCM) GetPaymentPlan(ctx context.Context, orgID uint64) (*PaymentPlan, error) {
	return &PaymentPlan{
		Name:         "Donald Duck",
		PrivateRepos: 5,
	}, nil
}

// UpdateRepository implements the SCM interface.
func (s *FakeSCM) UpdateRepository(ctx context.Context, repo *Repository) error {
	return nil
}
