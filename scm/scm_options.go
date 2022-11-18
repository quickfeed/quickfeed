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

// GroupOptions contain information about group.
type GroupOptions struct {
	OrganizationID uint64
	RepositoryID   uint64
	TeamID         uint64
}

func (opt *GroupOptions) valid() bool {
	return opt.OrganizationID > 0 && opt.RepositoryID > 0 && opt.TeamID > 0
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

func (opt *RejectEnrollmentOptions) valid() bool {
	return opt.OrganizationID > 0 && opt.RepositoryID > 0 && opt.User != ""
}

// OrganizationOptions contain information about organization.
type OrganizationOptions struct {
	ID   uint64
	Name string
	// Username field is used to filter organizations
	// where the given user has a certain role.
	Username  string
	NewCourse bool // Get organization for a new course
}

func (opt OrganizationOptions) valid() bool {
	return opt.ID != 0 || opt.Name != ""
}

// RepositoryOptions is used to fetch a single repository by ID or name.
// Either ID or both Path and Owner fields must be set.
type RepositoryOptions struct {
	ID    uint64
	Path  string
	Owner string
}

func (opt RepositoryOptions) valid() bool {
	return opt.ID > 0 || (opt.Path != "" && opt.Owner != "")
}

// CreateRepositoryOptions contains information on how a repository should be created.
type CreateRepositoryOptions struct {
	Organization string
	Path         string
	Private      bool
	Permission   string // Default permission level for the given repo. Can be "read", "write", "admin", "none".
}

func (opt CreateRepositoryOptions) valid() bool {
	return opt.Organization != "" && opt.Path != ""
}

// TeamOptions used when creating a new team
type TeamOptions struct {
	Organization string
	TeamName     string
	Users        []string
}

func (opt TeamOptions) valid() bool {
	return opt.TeamName != "" && opt.Organization != ""
}

// UpdateTeamOptions used when updating team members.
type UpdateTeamOptions struct {
	OrganizationID uint64
	TeamID         uint64
	Users          []string
}

func (opt UpdateTeamOptions) valid() bool {
	return opt.TeamID > 0 && opt.OrganizationID > 0
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

func (opt *IssueOptions) valid() bool {
	return opt.Organization != "" && opt.Repository != "" && opt.Title != "" && opt.Body != ""
}

// RequestReviewersOptions contains information on how to create or edit a pull request comment.
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
