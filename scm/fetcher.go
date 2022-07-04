package scm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	pb "github.com/quickfeed/quickfeed/ag"
)

const authUserName = "quickfeed" // can be anything except an empty string

func (s GithubSCM) Clone(opt *CloneOptions) (string, error) {
	cloneDir := filepath.Join(opt.DestDir, repoDir(opt))
	_, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		Auth:     &http.BasicAuth{Username: authUserName, Password: s.token},
		URL:      s.cloneURL(opt),
		Progress: os.Stdout,
		// ReferenceName: plumbing.ReferenceName(opt.Branch),
	})
	if err != nil {
		return "", err
	}
	return cloneDir, nil
}

func repoDir(opt *CloneOptions) string {
	if pb.RepoType(opt.Repository).IsStudentRepo() {
		return pb.AssignmentRepo
	}
	return pb.TestsRepo
}

// cloneURL returns the URL to clone the given repository.
func (s GithubSCM) cloneURL(opt *CloneOptions) string {
	return fmt.Sprintf("https://%s/%s/%s.git", s.providerURL, opt.Organization, opt.Repository)
}

type CloneOptions struct {
	Organization string
	Repository   string
	Branch       string
	DestDir      string
}

// TODO return this from Clone()
type Path struct {
	Assignments string
	Tests       string
}

type TestEnv struct{}

func (TestEnv) TestDir() string {
	// TODO(meling): return the path to the test directory
	return "test"
}

func (TestEnv) Validate() error {
	// TODO(meling): validate the test environment
	// validate that assignment folder exists
	// validate that the tests folder exists
	// validate that the student code contains code
	// validate that the student code doesn't contain the secret string
	// check student code for plagiarism
	return nil
}

// TODO(meling): check if {{ .AssignmentName }} exists in both tests and assignment directories
