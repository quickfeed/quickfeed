package scm

import (
	"context"
	"errors"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"go.uber.org/zap"
)

// SCM is a common interface for different source code management solutions,
// i.e., GitHub.
type SCM interface {
	// Updates an organization
	UpdateOrganization(context.Context, *OrganizationOptions) error
	// Gets an organization.
	GetOrganization(context.Context, *GetOrgOptions) (*qf.Organization, error)
	// Create a new repository.
	CreateRepository(context.Context, *CreateRepositoryOptions) (*Repository, error)
	// Get repository by ID or name
	GetRepository(context.Context, *RepositoryOptions) (*Repository, error)
	// Get repositories within organization.
	GetRepositories(context.Context, *qf.Organization) ([]*Repository, error)
	// Delete repository.
	DeleteRepository(context.Context, *RepositoryOptions) error
	// Add user as repository collaborator with provided permissions
	UpdateRepoAccess(context.Context, *Repository, string, string) error
	// Returns true if there are no commits in the given repository
	RepositoryIsEmpty(context.Context, *RepositoryOptions) bool
	// List the webhooks associated with the provided repository or organization.
	ListHooks(context.Context, *Repository, string) ([]*Hook, error)
	// Create a new webhook at the organization level.
	CreateHook(context.Context, *CreateHookOptions) error
	// Create team.
	CreateTeam(context.Context, *NewTeamOptions) (*Team, error)
	// Delete team.
	DeleteTeam(context.Context, *TeamOptions) error
	// Get a single team by ID or name.
	GetTeam(context.Context, *TeamOptions) (*Team, error)
	// Fetch all teams for organization.
	GetTeams(context.Context, *qf.Organization) ([]*Team, error)
	// Add repo to team.
	AddTeamRepo(context.Context, *AddTeamRepoOptions) error
	// AddTeamMember adds a member to a team.
	AddTeamMember(context.Context, *TeamMembershipOptions) error
	// RemoveTeamMember removes team member.
	RemoveTeamMember(context.Context, *TeamMembershipOptions) error
	// UpdateTeamMembers adds or removes members of an existing team based on list of users in TeamOptions.
	UpdateTeamMembers(context.Context, *UpdateTeamOptions) error
	// GetUserName returns the currently logged in user's login name.
	GetUserName(context.Context) (string, error)
	// GetUserNameByID returns the login name of user with the given remoteID.
	GetUserNameByID(context.Context, uint64) (string, error)
	// Returns a provider specific clone path.
	CreateCloneURL(*URLPathOptions) string
	// Promote or demote organization member based on Role field in OrgMembership.
	UpdateOrgMembership(context.Context, *OrgMembershipOptions) error
	// RemoveMember removes user from the organization.
	RemoveMember(context.Context, *OrgMembershipOptions) error

	// Clone clones the given repository and returns the path to the cloned repository.
	// The returned path is the provided destination directory joined with the
	// repository type, e.g., "assignments" or "tests".
	Clone(context.Context, *CloneOptions) (string, error)

	// CreateIssue creates an issue.
	CreateIssue(context.Context, *IssueOptions) (*Issue, error)
	// UpdateIssue edits an existing issue.
	UpdateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error)
	// GetIssue fetches a specific issue.
	GetIssue(ctx context.Context, opt *RepositoryOptions, number int) (*Issue, error)
	// GetIssues fetches all issues in a repository.
	GetIssues(ctx context.Context, opt *RepositoryOptions) ([]*Issue, error)
	// DeleteIssue deletes the given issue number in the given repository.
	DeleteIssue(context.Context, *RepositoryOptions, int) error
	// DeleteIssues deletes all issues in the given repository.
	DeleteIssues(context.Context, *RepositoryOptions) error

	// CreateIssueComment creates a comment on a SCM issue.
	CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error)
	// UpdateIssueComment edits a comment on a SCM issue.
	UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error

	// RequestReviewers requests reviewers for a pull request.
	RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error

	// Accepts repository invites.
	AcceptRepositoryInvites(context.Context, *RepositoryInvitationOptions) error
}

// NewSCMClient returns a new provider client implementing the SCM interface.
func NewSCMClient(logger *zap.SugaredLogger, token string) (SCM, error) {
	provider := env.ScmProvider()
	switch provider {
	case "github":
		return NewGithubSCMClient(logger, token), nil
	case "fake":
		return NewMockSCMClient(), nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

func newSCMAppClient(ctx context.Context, logger *zap.SugaredLogger, config *Config, organization string) (SCM, error) {
	provider := env.ScmProvider()
	switch provider {
	case "github":
		return newGithubAppClient(ctx, logger, config, organization)
	case "fake":
		return NewMockSCMClient(), nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

// OrganizationOptions contains information on how an organization should be
// created.
type OrganizationOptions struct {
	Name              string
	DefaultPermission string
	// prohibit students from creating new repos
	// on the course organization
	RepoPermissions bool
}

// GetOrgOptions contains information on the organization to fetch
type GetOrgOptions struct {
	ID   uint64
	Name string
	// Username field is used to filter organizations
	// where the given user has a certain role.
	Username string
}

// Repository represents a git remote repository.
type Repository struct {
	ID      uint64
	Path    string
	Owner   string // Only used by GitHub.
	HTMLURL string // Repository website.
	OrgID   uint64
	Size    uint64
}

// RepositoryOptions is used to fetch a single repository by ID or name.
// Either ID or both Path and Owner fields must be set.
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
	Organization string
	Path         string
	Private      bool
	Owner        string // The owner of an organization's repo is always the organization itself.
	Permission   string // Default permission level for the given repo. Can be "read", "write", "admin", "none".
}

// CreateHookOptions contains information on how to create a webhook.
// On GitHub, a single webhook is created on the organization level.
type CreateHookOptions struct {
	URL          string
	Secret       string
	Organization string
}

// TeamOptions contains information about the team and the organization it belongs to.
// It must include either both IDs or both names for the team and organization/
type TeamOptions struct {
	Organization   string
	OrganizationID uint64
	TeamName       string
	TeamID         uint64
}

// NewTeamOptions used when creating a new team
type NewTeamOptions struct {
	Organization string
	TeamName     string
	Users        []string
}

// UpdateTeamOptions used when updating team members.
type UpdateTeamOptions struct {
	OrganizationID uint64
	TeamID         uint64
	Users          []string
}

// TeamMembershipOptions contain information on organization team and associated user.
// Username and either team ID, or names of both the team and organization must be provided.
type TeamMembershipOptions struct {
	Organization   string
	OrganizationID uint64
	TeamID         uint64
	TeamName       string
	Username       string // GitHub username.
	Role           string // "Member" or "maintainer". A maintainer can add, remove and promote team members.
}

// OrgMembershipOptions represent user's membership in organization
type OrgMembershipOptions struct {
	Organization string
	Username     string // GitHub username.
	Role         string // Role can be "admin" (organization owner) or "member".
}

// URLPathOptions holds elements used when constructing a clone URL string.
type URLPathOptions struct {
	UserToken    string
	Organization string
	Repository   string
}

// AddTeamRepoOptions contains information about the repos to be added to a team.
// All fields must be provided.
type AddTeamRepoOptions struct {
	OrganizationID uint64
	TeamID         uint64
	Repo           string
	Owner          string // Name of the organization. Only used by GitHub.
	Permission     string // Permission level for team members. Can be "push", "pull", "admin".
}

// Team represents a git Team
type Team struct {
	ID           uint64
	Name         string
	Organization string
}

// Authorization stores information about user scopes
type Authorization struct {
	Scopes []string
}

// Issue represents an SCM issue.
type Issue struct {
	ID         uint64
	Title      string
	Body       string
	Repository string
	Assignee   string
	Status     string
	Number     int
}

// IssueOptions contains information for creating or updating an Issue.
type IssueOptions struct {
	Organization string
	Repository   string
	Title        string
	Body         string
	State        string
	Labels       *[]string
	Assignee     *string
	Assignees    *[]string
	Number       int
}

// RequestReviewersOptions contains information on how to create or edit a pull request comment.
type IssueCommentOptions struct {
	Organization string
	Repository   string
	Body         string
	Number       int
	CommentID    int64
}

// RequestReviewersOptions contains information on how to assign reviewers to a pull request.
type RequestReviewersOptions struct {
	Organization string
	Repository   string
	Number       int
	Reviewers    []string // Reviewers is a slice of github usernames
}
