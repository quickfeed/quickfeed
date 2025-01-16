package scm

import (
	"context"
	"errors"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"go.uber.org/zap"
)

// SCM is the source code management interface for managing courses and users.
type SCM interface {
	// Gets an organization.
	GetOrganization(context.Context, *OrganizationOptions) (*qf.Organization, error)
	// Get repositories within organization.
	GetRepositories(context.Context, string) ([]*Repository, error)
	// Returns true if there are no commits in the given repository
	RepositoryIsEmpty(context.Context, *RepositoryOptions) bool

	// CreateCourse creates repositories for a new course.
	CreateCourse(context.Context, *CourseOptions) ([]*Repository, error)
	// UpdateEnrollment updates organization membership and creates and grants access to user repository.
	UpdateEnrollment(context.Context, *UpdateEnrollmentOptions) (*Repository, error)
	// RejectEnrollment removes user's repository and revokes user's membership in the course organization.
	RejectEnrollment(context.Context, *RejectEnrollmentOptions) error
	// DemoteTeacherToStudent revokes a user's owner status in the organization.
	DemoteTeacherToStudent(context.Context, *UpdateEnrollmentOptions) error
	// CreateGroup creates repository for a new group.
	CreateGroup(context.Context, *GroupOptions) (*Repository, error)
	// UpdateGroupMembers adds or removes members of an existing group.
	UpdateGroupMembers(context.Context, *GroupOptions) error
	// DeleteGroup deletes group's repository.
	DeleteGroup(context.Context, uint64) error

	// Clone clones the given repository and returns the path to the cloned repository.
	// The returned path is the provided destination directory joined with the
	// repository type, e.g., "assignments" or "tests".
	Clone(context.Context, *CloneOptions) (string, error)

	// AcceptInvitations accepts course invites on behalf of the user.
	// A new refresh token for the user is returned, which may be used in subsequent requests.
	AcceptInvitations(context.Context, *InvitationOptions) (string, error)

	// CreateIssue creates an issue.
	CreateIssue(context.Context, *IssueOptions) (*Issue, error)
	// UpdateIssue edits an existing issue.
	UpdateIssue(context.Context, *IssueOptions) (*Issue, error)
	// GetIssue fetches a specific issue.
	GetIssue(context.Context, *RepositoryOptions, int) (*Issue, error)
	// GetIssues fetches all issues in a repository.
	GetIssues(context.Context, *RepositoryOptions) ([]*Issue, error)
	// DeleteIssue deletes the given issue number in the given repository.
	DeleteIssue(context.Context, *RepositoryOptions, int) error
	// DeleteIssues deletes all issues in the given repository.
	DeleteIssues(context.Context, *RepositoryOptions) error
	// CreateIssueComment creates a comment on a SCM issue.
	CreateIssueComment(context.Context, *IssueCommentOptions) (int64, error)
	// UpdateIssueComment edits a comment on a SCM issue.
	UpdateIssueComment(context.Context, *IssueCommentOptions) error
	// RequestReviewers requests reviewers for a pull request.
	RequestReviewers(context.Context, *RequestReviewersOptions) error
}

// NewSCMClient returns a new provider client implementing the SCM interface.
func NewSCMClient(logger *zap.SugaredLogger, token string) (SCM, error) {
	provider := env.ScmProvider()
	switch provider {
	case "github":
		return NewGithubSCMClient(logger, token), nil
	case "fake":
		return NewMockedGithubSCMClient(logger, WithMockOrgs()), nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

func newSCMAppClient(ctx context.Context, logger *zap.SugaredLogger, config *Config, organization string) (SCM, error) {
	provider := env.ScmProvider()
	switch provider {
	case "github":
		return newGithubAppClient(ctx, logger, config, organization)
	case "fake":
		return NewMockedGithubSCMClient(logger, WithMockOrgs()), nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

// Repository represents a git remote repository.
type Repository struct {
	ID      uint64
	Repo    string
	Owner   string // Only used by GitHub.
	HTMLURL string // Repository website.
}

// Issue represents an SCM issue.
type Issue struct {
	ID         int64
	Title      string
	Body       string
	Repository string
	Assignee   string
	Status     string
	Number     int
}
