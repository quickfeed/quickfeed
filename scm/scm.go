package scm

import (
	"context"
	"errors"

	pb "github.com/autograde/aguis/ag"
	"go.uber.org/zap"
)

const (

	// OrgOwner is organization owner
	OrgOwner = "admin"
	// OrgMember is organization member
	OrgMember = "member"

	// TeamMaintainer can add and delete team users and repos
	TeamMaintainer = "maintainer"
	// TeamMember is a regular member
	TeamMember = "member"

	// repository permission levels for organization

	// OrgPull allows only pull access to organization repositories
	OrgPull = "read"
	// OrgPush allows pull and push acces to organization repositories
	OrgPush = "write"
	// OrgFull allows to pull/push, create, remove and update organization repositories
	OrgFull = "admin"
	// OrgNone allows no access to organization repositories
	OrgNone = "none"

	// repository permission levels for a user

	// RepoPull allows only pull access to repository
	RepoPull = "pull"
	// RepoPush allows pull and push access to repository
	RepoPush = "push"
	// RepoFull allows full access to repository
	RepoFull = "admin"
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
	GetOrganization(context.Context, uint64) (*pb.Organization, error)
	// Create a new repository.
	CreateRepository(context.Context, *CreateRepositoryOptions) (*Repository, error)
	// Get repositories within organization.
	GetRepositories(context.Context, *pb.Organization) ([]*Repository, error)
	// Delete repository.
	DeleteRepository(context.Context, uint64) error
	// Add user as repository collaborator with provided permissions
	UpdateRepoAccess(context.Context, *Repository, string, string) error
	// List the webhooks associated with the provided repository.
	ListHooks(context.Context, *Repository, string) ([]*Hook, error)
	// Creates a new webhook.
	CreateHook(context.Context, *CreateHookOptions) error
	// Create team.
	CreateTeam(context.Context, *CreateTeamOptions) (*Team, error)
	// Delete team.
	DeleteTeam(context.Context, *CreateTeamOptions) error
	// Get a single team by ID or name
	GetTeam(context.Context, *CreateTeamOptions) (*Team, error)
	// Fetch all teams for organization
	GetTeams(context.Context, *pb.Organization) ([]*Team, error)
	// Add repo to team.
	AddTeamRepo(context.Context, *AddTeamRepoOptions) error
	// AddTeamMember adds a member to a team.
	AddTeamMember(context.Context, *TeamMembershipOptions) error
	// RemoveTeamMember removes team member
	RemoveTeamMember(context.Context, *TeamMembershipOptions) error
	// UpdateTeamMembers adds or removes members of an existing team based on list of users in CreateTeamOptions
	UpdateTeamMembers(context.Context, *CreateTeamOptions) error
	// GetUserName returns the currently logged in user's login name.
	GetUserName(context.Context) (string, error)
	// GetUserNameByID returns the login name of user with the given remoteID.
	GetUserNameByID(context.Context, uint64) (string, error)
	// Returns a provider specific clone path.
	CreateCloneURL(*CreateClonePathOptions) string
	// Promotes or demotes organization member, based on Role field in OrgMembership
	UpdateOrgMembership(context.Context, *OrgMembershipOptions) error
	// Lists all authorizations for authenticated user
	GetUserScopes(context.Context) *Authorization
}

// NewSCMClient returns a new provider client implementing the SCM interface.
func NewSCMClient(logger *zap.Logger, provider, token string) (SCM, error) {
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

// Valid checks that required struct fields are set
func (opt CreateOrgOptions) valid() bool {
	return opt.Path != "" &&
		opt.DefaultPermission != ""
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
}

// ValidForHooks checks that repository object can be used in hooks related methods
func (r Repository) ValidForHooks() bool {
	return r.Path != "" && r.Owner != ""
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
	Owner        string // we can create user repositories
	Permission   string // default permission level for the repo. Can be "read", "write", "admin", "none"
}

// CreateHookOptions contains information on how to create a webhook.
type CreateHookOptions struct {
	URL        string
	Secret     string
	Repository *Repository
}

// CreateTeamOptions contains information about the team and the users of the team.
type CreateTeamOptions struct {
	Organization *pb.Organization
	TeamName     string
	TeamID       uint64
	Users        []string
}

// TeamMembershipOptions contain information on organization team and user to be added
type TeamMembershipOptions struct {
	Organization *pb.Organization
	TeamID       int64
	TeamSlug     string // slugified team name
	Username     string // GitHub username
	Role         string // member or maintainer. Maintainer can add, remove and promote team members
}

// OrgMembershipOptions represent user's membership in organization
type OrgMembershipOptions struct {
	Organization *pb.Organization
	Username     string // GitHub username
	Role         string // role can be "admin" (organization owner) or "member"
}

// ErrNotSupported is returned when the source code management solution used
// does not provide a sufficient API for the method called.
type ErrNotSupported struct {
	SCM    string
	Method string
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
	Owner      string // only used by GitHub
	Permission string // permission level for team members. Can be "push", "pull", "admin"
}

// Valid checks that every field is set to avoid unnecessary scm calls with invalid arguments
func (opt AddTeamRepoOptions) Valid() bool {
	return opt.TeamID > 0 &&
		opt.Repo != "" &&
		opt.Owner != "" &&
		opt.Permission != ""
}

// Team represents a git Team
type Team struct {
	ID   uint64
	Name string
	URL  string
}

// Authorization stores information about user scopes
type Authorization struct {
	Token  string
	Scopes []string
}

func (e ErrNotSupported) Error() string {
	return "method " + e.Method + " not supported by " + e.SCM + " SCM"
}

// ErrMissingFields is returned when scm struct validation fails.
// This error only used for development/debugging and never goes to frontend user.
var ErrMissingFields = errors.New("invalid argument: missing required fields")
