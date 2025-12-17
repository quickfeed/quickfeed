package scm

import "github.com/quickfeed/quickfeed/qf"

// CourseOptions contain information about new course.
type CourseOptions struct {
	OrganizationID uint64
	CourseCreator  string
}

func (opt CourseOptions) valid() bool {
	return opt.OrganizationID > 0 && opt.CourseCreator != ""
}

// UpdateEnrollmentOptions contain information about enrollment.
type UpdateEnrollmentOptions struct {
	Organization string
	User         string
	Status       qf.Enrollment_UserStatus
}

func (opt UpdateEnrollmentOptions) valid() bool {
	return opt.Organization != "" && opt.User != ""
}

// RejectEnrollmentOptions contain information about enrollment.
type RejectEnrollmentOptions struct {
	OrganizationID uint64
	RepositoryID   uint64
	User           string
}

func (opt RejectEnrollmentOptions) valid() bool {
	return opt.OrganizationID > 0 && opt.RepositoryID > 0 && opt.User != ""
}

// OrganizationOptions contain information about organization.
type OrganizationOptions struct {
	ID        uint64
	Name      string
	Username  string // Username, if provide, must be owner of the organization
	NewCourse bool   // Get organization for a new course
}

func (opt OrganizationOptions) valid() bool {
	return opt.ID != 0 || opt.Name != ""
}

// RepositoryOptions is used to fetch a single repository by ID or name.
// Either ID or both Path and Owner fields must be set.
type RepositoryOptions struct {
	ID    uint64
	Repo  string
	Owner string
}

func (opt RepositoryOptions) valid() bool {
	return opt.ID > 0 || (opt.Repo != "" && opt.Owner != "")
}

// CreateRepositoryOptions contains information on how a repository should be created.
type CreateRepositoryOptions struct {
	Owner   string
	Repo    string
	Private bool
}

func (opt CreateRepositoryOptions) valid() bool {
	return opt.Owner != "" && opt.Repo != ""
}

// GroupOptions is used when creating or modifying a group.
type GroupOptions struct {
	Organization string   // Organization is the owner of the repository
	GroupName    string   // GroupName is the name of the repository
	Users        []string // Users are group collaborators (GitHub usernames)
}

func (opt GroupOptions) valid() bool {
	return opt.GroupName != "" && opt.Organization != ""
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

func (opt IssueOptions) valid() bool {
	return opt.Organization != "" && opt.Repository != "" && opt.Title != "" && opt.Body != ""
}

// IssueCommentOptions contains information for creating or updating an IssueComment.
type IssueCommentOptions struct {
	Organization string
	Repository   string
	Body         string
	Number       int
	CommentID    int64
}

func (opt IssueCommentOptions) valid() bool {
	return opt.Organization != "" && opt.Repository != "" && opt.Body != ""
}

// RequestReviewersOptions contains information on how to assign reviewers to a pull request.
type RequestReviewersOptions struct {
	Organization string
	Repository   string
	Number       int
	Reviewers    []string // Reviewers is a slice of github usernames
}

func (opt RequestReviewersOptions) valid() bool {
	return opt.Organization != "" && opt.Repository != "" && opt.Number > 0 && len(opt.Reviewers) != 0
}

// SyncForkOptions contains information for syncing a forked repository with its upstream.
type SyncForkOptions struct {
	Organization string
	Repository   string
	Branch       string
}

func (opt SyncForkOptions) valid() bool {
	return opt.Organization != "" && opt.Repository != "" && opt.Branch != ""
}
