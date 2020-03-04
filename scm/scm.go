package scm

import (
	"context"
	"errors"

	pb "github.com/autograde/aguis/ag"
	"go.uber.org/zap"
)

// SCM is a common interface for different source code management solutions,
// i.e., GitHub and GitLab.
type SCM interface {
	// Lists organizations (for logged in user) which can be used as a course directory.
	ListOrganizations(context.Context) ([]*pb.Organization, error)
	// Creates a new organization.
	CreateOrganization(context.Context, *CreateOrgOptions) (*pb.Organization, error)
	// Updates an organization
	UpdateOrganization(context.Context, *CreateOrgOptions) error
	// Gets an organization.
	GetOrganization(context.Context, *GetOrgOptions) (*pb.Organization, error)
	// Create a new repository.
	CreateRepository(context.Context, *CreateRepositoryOptions) (*Repository, error)
	// Get repository by ID or name
	GetRepository(context.Context, *RepositoryOptions) (*Repository, error)
	// Get repositories within organization.
	GetRepositories(context.Context, *pb.Organization) ([]*Repository, error)
	// Delete repository.
	DeleteRepository(context.Context, *RepositoryOptions) error
	// Add user as repository collaborator with provided permissions
	UpdateRepoAccess(context.Context, *Repository, string, string) error
	// Returns true if there are no commits in the given repository
	RepositoryIsEmpty(context.Context, *RepositoryOptions) bool
	// List the webhooks associated with the provided repository.
	ListHooks(context.Context, *Repository, string) ([]*Hook, error)
	// Creates a new webhook.
	CreateHook(context.Context, *CreateHookOptions) error
	// Create an organization level webhook
	CreateOrgHook(context.Context, *OrgHookOptions) error
	// Create team.
	CreateTeam(context.Context, *TeamOptions) (*Team, error)
	// Delete team.
	DeleteTeam(context.Context, *TeamOptions) error
	// Get a single team by ID or name
	GetTeam(context.Context, *TeamOptions) (*Team, error)
	// Fetch all teams for organization
	GetTeams(context.Context, *pb.Organization) ([]*Team, error)
	// Add repo to team.
	AddTeamRepo(context.Context, *AddTeamRepoOptions) error
	// AddTeamMember adds a member to a team.
	AddTeamMember(context.Context, *TeamMembershipOptions) error
	// RemoveTeamMember removes team member
	RemoveTeamMember(context.Context, *TeamMembershipOptions) error
	// UpdateTeamMembers adds or removes members of an existing team based on list of users in TeamOptions
	UpdateTeamMembers(context.Context, *TeamOptions) error
	// GetUserName returns the currently logged in user's login name.
	GetUserName(context.Context) (string, error)
	// GetUserNameByID returns the login name of user with the given remoteID.
	GetUserNameByID(context.Context, uint64) (string, error)
	// Returns a provider specific clone path.
	CreateCloneURL(*CreateClonePathOptions) string
	// Promotes or demotes organization member, based on Role field in OrgMembership
	UpdateOrgMembership(context.Context, *OrgMembershipOptions) error
	// RevokeOrgMembership removes user from the organization
	RemoveMember(context.Context, *OrgMembershipOptions) error
	// Lists all authorizations for authenticated user
	GetUserScopes(context.Context) *Authorization
}

// NewSCMClient returns a new provider client implementing the SCM interface.
func NewSCMClient(logger *zap.SugaredLogger, provider, token string) (SCM, error) {
	switch provider {
	case "github":
		return NewGithubSCMClient(logger, token), nil
	case "gitlab":
		return NewGitlabSCMClient(token), nil
	case "fake":
		return NewFakeSCMClient(), nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

// CreateOrgOptions contains information on how an organization should be
// created.
type CreateOrgOptions struct {
	Path              string
	Name              string
	DefaultPermission string
}

// GetOrgOptions contains information on the organization to fetch
type GetOrgOptions struct {
	ID       uint64
	Name     string
	Username string
}

// Repository represents a git remote repository.
type Repository struct {
	ID      uint64
	Path    string
	Owner   string // Only used by GitHub.
	WebURL  string // Repository website.
	SSHURL  string // SSH clone URL.
	HTTPURL string // HTTP(S) clone URL.
	OrgID   uint64
	Size    uint64
}

// RepositoryOptions used to fetch a single repository by ID or name
// either ID or both Path and Owner info must be provided
type RepositoryOptions struct {
	ID    uint64
	Path  string
	Owner string
}

// Hook contains information about a webhook for a repository.
type Hook struct {
	ID     uint64
	Name   string
	URL    string
	Events []string
}

// CreateRepositoryOptions contains information on how a repository should be created.
type CreateRepositoryOptions struct {
	Organization *pb.Organization
	Path         string
	Private      bool
	Owner        string // we can create user repositories. Default owner is github organization
	Permission   string // default permission level for the repo. Can be "read", "write", "admin", "none"
}

// CreateHookOptions contains information on how to create a webhook.
type CreateHookOptions struct {
	URL        string
	Secret     string
	Repository *Repository
}

// OrgHookOptions contains information about an organization level hook
type OrgHookOptions struct {
	URL          string
	Secret       string
	Organization string
}

// TeamOptions contains information about the team and the users of the team.
type TeamOptions struct {
	Organization string
	TeamName     string
	TeamID       uint64
	Users        []string
}

// TeamMembershipOptions contain information on organization team and associated user
type TeamMembershipOptions struct {
	Organization string
	TeamID       int64
	TeamSlug     string // slugified team name
	Username     string // GitHub username
	Role         string // member or maintainer. Maintainer can add, remove and promote team members
}

// OrgMembershipOptions represent user's membership in organization
type OrgMembershipOptions struct {
	Organization string
	Username     string // GitHub username
	Role         string // role can be "admin" (organization owner) or "member"
}

// CreateClonePathOptions holds elements used when constructing a clone URL string.
type CreateClonePathOptions struct {
	UserToken    string
	Organization string
	Repository   string
}

// AddTeamRepoOptions contains information about the repos to be added to a team.
type AddTeamRepoOptions struct {
	TeamID     uint64
	Repo       string
	Owner      string // Name of the team to associate repo with. Only used by GitHub.
	Permission string // permission level for team members. Can be "push", "pull", "admin"
}

// Team represents a git Team
type Team struct {
	ID   uint64
	Name string
	URL  string
}

// Authorization stores information about user scopes
type Authorization struct {
	Scopes []string
}
