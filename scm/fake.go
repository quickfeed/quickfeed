package scm

import (
	"context"
	"errors"
	"strconv"
)

// FakeSCM implements the SCM interface.
type FakeSCM struct {
	Repositories map[uint64]*Repository
	Directories  map[uint64]*Directory
}

// NewFakeSCMClient returns a new Fake client implementing the SCM interface.
func NewFakeSCMClient() *FakeSCM {
	return &FakeSCM{
		Repositories: make(map[uint64]*Repository),
		Directories:  make(map[uint64]*Directory),
	}
}

// ListDirectories implements the SCM interface.
func (s *FakeSCM) ListDirectories(ctx context.Context) ([]*Directory, error) {
	var dirs []*Directory
	for _, dir := range s.Directories {
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

// CreateDirectory implements the SCM interface.
func (s *FakeSCM) CreateDirectory(ctx context.Context, opt *CreateDirectoryOptions) (*Directory, error) {
	id := len(s.Directories) + 1
	dir := &Directory{
		ID:     uint64(id),
		Path:   opt.Path,
		Avatar: "https://avatars3.githubusercontent.com/u/1000" + strconv.Itoa(id) + "?v=3",
	}
	s.Directories[dir.ID] = dir
	return dir, nil
}

// GetDirectory implements the SCM interface.
func (s *FakeSCM) GetDirectory(ctx context.Context, id uint64) (*Directory, error) {
	dir, ok := s.Directories[id]
	if !ok {
		return nil, errors.New("directory not found")
	}
	return dir, nil
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
		DirectoryID: opt.Directory.ID,
	}
	s.Repositories[repo.ID] = repo
	return repo, nil
}

// GetRepositories implements the SCM interface.
func (s *FakeSCM) GetRepositories(ctx context.Context, directory *Directory) ([]*Repository, error) {
	var repos []*Repository
	for _, repo := range s.Repositories {
		if repo.DirectoryID == directory.ID {
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

// CreateHook implements the SCM interface.
func (s *FakeSCM) CreateHook(context.Context, *CreateHookOptions) (err error) { return }
