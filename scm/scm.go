package scm

import (
	"context"
	"errors"

	pb "github.com/autograde/aguis/ag"
)

// SCM is a common interface for different source code management solutions,
// i.e., GitHub and GitLab.
type SCM interface {
	// Lists directories which can be used as a course directory.
	ListDirectories(context.Context) ([]*pb.Directory, error)
	// Creates a new directory.
	CreateDirectory(context.Context, *CreateDirectoryOptions) (*pb.Directory, error)
	// Gets a directory.
	GetDirectory(context.Context, uint64) (*pb.Directory, error)
	// Create a new repository.
	CreateRepository(context.Context, *CreateRepositoryOptions) (*Repository, error)
	// Get repositories within directory.
	GetRepositories(context.Context, *pb.Directory) ([]*Repository, error)
	// Update repository settings
	UpdateRepository(context.Context, *Repository) error
	// Delete repository.
	DeleteRepository(context.Context, uint64) error
	// List the webhooks associated with the provided repository.
	ListHooks(context.Context, *Repository) ([]*Hook, error)
	// Creates a new webhook.
	CreateHook(context.Context, *CreateHookOptions) error
	// Create team.
	CreateTeam(context.Context, *CreateTeamOptions) (*Team, error)
	// Add repo to team.
	AddTeamRepo(context.Context, *AddTeamRepoOptions) error
	// AddTeamMember as a member to a team.
	// AddTeamMember(context.Context, *AddMemberOptions) error
	// Retrieves user by remoteID
	GetUserNameByID(context.Context, uint64) (string, error)
	// Returns a provider spesefic clone path.
	CreateCloneURL(context.Context, *CreateClonePathOptions) (string, error)
	// Fetch current payment plan
	GetPaymentPlan(context.Context, uint64) (*PaymentPlan, error)
}

// NewSCMClient returns a new provider client implementing the SCM interface.
func NewSCMClient(provider, token string) (SCM, error) {
	switch provider {
	case "github":
		return NewGithubSCMClient(token), nil
	case "gitlab":
		return NewGitlabSCMClient(token), nil
	case "fake":
		return NewFakeSCMClient(), nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

// Directory represents an entity which is capable of managing source code
// repositories as well as user access to those repositories.
type Directory struct {
	ID     uint64 `json:"id"`
	Path   string `json:"path"`
	Avatar string `json:"avatar,omitempty"`
}

// CreateDirectoryOptions contains information on how a directory should be
// created.
type CreateDirectoryOptions struct {
	Path string
	Name string
}

// Repository represents a git remote repository.
type Repository struct {
	ID   uint64
	Path string

	// Only used by GitHub.
	Owner string

	// Repository website.
	WebURL string
	// SSH clone URL.
	SSHURL string
	// HTTP(S) clone URL.
	HTTPURL string

	DirectoryID uint64
}

// Hook contains information about a webhook for a repository.
type Hook struct {
	ID   uint64
	Name string
	URL  string
}

// CreateRepositoryOptions contains information on how a repository should be
// created.
type CreateRepositoryOptions struct {
	Path      string
	Directory *pb.Directory
	Private   bool
}

// CreateHookOptions contains information on how to create a webhook.
type CreateHookOptions struct {
	URL    string
	Secret string

	Repository *Repository
}

// CreateTeamOptions contains information about the team and the users of the team.
type CreateTeamOptions struct {
	Directory *Directory
	TeamName  string
	Users     []string
}

// ErrNotSupported is returned when the source code management solution used
// does not provide a sufficient API for the method called.
type ErrNotSupported struct {
	SCM    string
	Method string
}

type CreateClonePathOptions struct {
	UserToken  string
	Directory  string
	Repository string
}

// AddTeamRepoOptions contains information about the repos to be added to a team.
type AddTeamRepoOptions struct {
	TeamID uint64
	Repo   string

	// Only used by GitHub.
	Owner string
}

// Team represents a git Team
type Team struct {
	ID   uint64
	Name string
	URL  string
}

type PaymentPlan struct {
	Name         string
	PrivateRepos uint64
}

func (e ErrNotSupported) Error() string {
	return "method" + e.Method + " not supported by " + e.SCM + " SCM"
}
