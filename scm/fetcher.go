package scm

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const authUserName = "quickfeed" // can be anything except an empty string

// Clone clones the given repository and returns the path to the cloned repository.
// If the repository already exists, it is updated using git pull.
// The returned path is the provided destination directory joined with the repository.
func (s *GithubSCM) Clone(ctx context.Context, opt *CloneOptions) (string, error) {
	if s.config != nil {
		// GitHubSCM is being used as a GitHub App, and since the go-git library requires
		// token-based authentication, we first refresh the token since it may have expired.
		if err := s.refreshToken(opt.Organization); err != nil {
			return "", err
		}
	}

	authInfo := &http.BasicAuth{Username: authUserName, Password: s.token}

	cloneDir := filepath.Join(opt.DestDir, opt.Repository)
	r, err := git.PlainOpen(cloneDir)
	if err == nil {
		// Repository already exists, pull the latest changes
		s.logger.Debugf("Pulling(%s)", s.cloneURL(opt))
		w, err := r.Worktree()
		if err != nil {
			return "", err
		}
		err = w.Pull(&git.PullOptions{
			Auth:       authInfo,
			RemoteName: "origin",
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", err
		}
		return cloneDir, nil
	} else if err != git.ErrRepositoryNotExists {
		return "", err
	}

	s.logger.Debugf("Clone(%s)", s.cloneURL(opt))
	var branch plumbing.ReferenceName
	if opt.Branch != "" {
		branch = plumbing.NewBranchReferenceName(opt.Branch)
	}
	_, err = git.PlainCloneContext(ctx, cloneDir, false, &git.CloneOptions{
		Auth:          authInfo,
		URL:           s.cloneURL(opt),
		ReferenceName: branch,
	})
	if err != nil {
		return "", err
	}
	s.logger.Debugf("CloneDir = %s", cloneDir)
	return cloneDir, nil
}

// cloneURL returns the URL to clone the given repository.
func (s *GithubSCM) cloneURL(opt *CloneOptions) string {
	return fmt.Sprintf("%s/%s/%s", s.providerURL, opt.Organization, opt.Repository)
}

type CloneOptions struct {
	Organization string
	Repository   string
	Branch       string
	DestDir      string
}
