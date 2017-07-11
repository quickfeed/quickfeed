package scm

import (
	"context"
	"fmt"
)

// FakeSCM implements the SCM interface.
type FakeSCM struct {
}

// ListDirectories implements the SCM interface.
func (s *FakeSCM) ListDirectories(ctx context.Context) ([]*Directory, error) {
	return []*Directory{
		{
			ID:     6347238,
			Path:   "fake-repo-6347238",
			Avatar: "https://avatars3.githubusercontent.com/u/6347238?v=3",
		},
		{
			ID:     7581319,
			Path:   "fake-repo-7581319",
			Avatar: "https://avatars3.githubusercontent.com/u/7581319?v=3",
		},
		{
			ID:     13813278,
			Path:   "fake-repo-13813278",
			Avatar: "https://avatars3.githubusercontent.com/u/13813278?v=3",
		},
		{
			ID:     14003302,
			Path:   "fake-repo-14003302",
			Avatar: "https://avatars3.githubusercontent.com/u/14003302?v=3",
		},
		{
			ID:     16490855,
			Path:   "fake-repo-16490855",
			Avatar: "https://avatars3.githubusercontent.com/u/16490855?v=3",
		},
		{
			ID:     23650610,
			Path:   "fake-repo-23650610",
			Avatar: "https://avatars3.githubusercontent.com/u/23650610?v=3",
		},
		{
			ID:     29543863,
			Path:   "fake-repo-29543863",
			Avatar: "https://avatars3.githubusercontent.com/u/29543863?v=3",
		},
	}, nil
}

// CreateDirectory implements the SCM interface.
func (s *FakeSCM) CreateDirectory(ctx context.Context, opt *CreateDirectoryOptions) (*Directory, error) {
	return &Directory{
		ID:     999,
		Path:   opt.Path,
		Avatar: "https://avatars3.githubusercontent.com/u/29543863?v=3",
	}, nil
}

// GetDirectory implements the SCM interface.
func (s *FakeSCM) GetDirectory(ctx context.Context, id uint64) (*Directory, error) {
	return &Directory{
		ID:     id,
		Path:   fmt.Sprintf("fake-repo-%d", id),
		Avatar: fmt.Sprintf("https://avatars3.githubusercontent.com/u/%d?v=3", id),
	}, nil
}

// CreateRepository implements the SCM interface.
func (s *FakeSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	return newRepository(123456, opt.Directory), nil
}

// GetRepositories implements the SCM interface.
func (s *FakeSCM) GetRepositories(ctx context.Context, directory *Directory) ([]*Repository, error) {
	return []*Repository{
		newRepository(123456, directory),
		newRepository(234567, directory),
		newRepository(345678, directory),
	}, nil
}

func newRepository(id uint64, dir *Directory) *Repository {
	var dirID uint64
	var dirPath string
	if dir.Path != "" {
		dirID = 999
		dirPath = dir.Path
	} else {
		dirID = dir.ID
		dirPath = fmt.Sprintf("fake-dir-%d", dirID)
	}
	repoPath := fmt.Sprintf("fake-repo-%d", id)
	return &Repository{
		ID:          id,
		Path:        repoPath,
		WebURL:      "https://example.com/" + dirPath + "/" + repoPath,
		SSHURL:      "git@example.com:" + dirPath + "/" + repoPath,
		HTTPURL:     "https://example.com/" + dirPath + "/" + repoPath + ".git",
		DirectoryID: dirID,
	}
}

// DeleteRepository implements the SCM interface.
func (s *FakeSCM) DeleteRepository(context.Context, uint64) error {
	return nil
}
