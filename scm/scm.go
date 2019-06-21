package scm

import (
	"context"
	"errors"

	pb "github.com/autograde/aguis/ag"
)

// SCM is a common interface for different source code management solutions,
// i.e., GitHub and GitLab.
type SCM interface {
	// Lists organizations which can be used as a course directory.
	ListOrganizations(context.Context) ([]*pb.Organization, error)
	// Creates a new organization.
	CreateOrganization(context.Context, *CreateOrgOptions) (*pb.Organization, error)
	// Gets an organization.
	GetOrganization(context.Context, uint64) (*pb.Organization, error)
	// CreateRepoAndTeam invokes the SCM to create a repository and team for the
	// specified namespace (typically the course name), the path of the repository
	// (typically the name of the student with a '-labs' suffix or the group name).
	// The team name is usually the student name or group name, whereas the git
	// user names are the members of the team. For single student repositories,
	// the git user names are typically just the one student.
	CreateRepoAndTeam(ctx context.Context, opt *CreateRepositoryOptions, teamName string, gitUserNames []string) (*Repository, *Team, error)
	// Create a new repository.
	CreateRepository(context.Context, *CreateRepositoryOptions) (*Repository, error)
	// Get repositories within organization.
	GetRepositories(context.Context, *pb.Organization) ([]*Repository, error)
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
	// Delete team.
	DeleteTeam(context.Context, uint64) error
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
	// Returns a provider specific clone path.
	CreateCloneURL(*CreateClonePathOptions) string
	// Fetch current payment plan
	GetPaymentPlan(context.Context, uint64) (*PaymentPlan, error)
	// Get user's membership in organization
	GetOrgMembership(context.Context, *OrgMembership) (*OrgMembership, error)
	// Promotes or demotes organization member, based on Role field in OrgMembership
	UpdateOrgMembership(context.Context, *OrgMembership) error
	// Invite user to course organization
	CreateOrgMembership(context.Context, *OrgMembershipOptions) error
	// Lists all authorizations for authenticated user
	GetUserScopes(context.Context) *Authorization
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

// CreateOrgOptions contains information on how an organization should be
// created.
type CreateOrgOptions struct {
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

	OrgID uint64
}

// Hook contains information about a webhook for a repository.
type Hook struct {
	ID   uint64
	Name string
	URL  string
}

// CreateRepositoryOptions contains information on how a repository should be created.
type CreateRepositoryOptions struct {
	Organization *pb.Organization
	Path         string
	Private      bool
	Owner        string // we can create user repositories
}

// CreateHookOptions contains information on how to create a webhook.
type CreateHookOptions struct {
	URL    string
	Secret string

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
	TeamSlug     string // team name in all lowercase
	Username     string // GitHub username
	Role         string // member or maintainer. Maintainer can add, remove and promote team members
}

// OrgMembershipOptions provides information on user to be invited to organization
type OrgMembershipOptions struct {
	Organization *pb.Organization
	Username     string // GitHub username
	Email        string // we can also send invites by email
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
	TeamID uint64
	Repo   string
	Owner  string // only used by GitHub
}

// Team represents a git Team
type Team struct {
	ID   uint64
	Name string
	URL  string
}

// PaymentPlan represents the payment plan to use.
type PaymentPlan struct {
	Name         string
	PrivateRepos uint64 // max number allowed on the org
}

// OrgMembership represents user's membership in organization
type OrgMembership struct {
	Username string
	OrgID    uint64
	Role     string // role can be "admin" (organization owner) or "member"
}

// Authorization stores information about user scopes
type Authorization struct {
	Token  string
	Scopes []string
}

func (e ErrNotSupported) Error() string {
	return "method " + e.Method + " not supported by " + e.SCM + " SCM"
}
