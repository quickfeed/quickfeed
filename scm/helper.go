package scm

import "errors"

const (
	// Organization roles //

	// OrgOwner is organization's owner
	OrgOwner = "admin"
	// OrgMember is organization's member
	OrgMember = "member"

	// Team roles //

	// TeamMaintainer can add and delete team users and repos
	TeamMaintainer = "maintainer"
	// TeamMember is a regular member
	TeamMember = "member"

	// Repository permission levels for organization //

	// OrgPull allows only pull access to organization repositories
	OrgPull = "read"
	// OrgPush allows pull and push access to organization repositories
	OrgPush = "write"
	// OrgFull allows to pull/push, create, remove and update organization repositories
	OrgFull = "admin"
	// OrgNone allows no access to organization repositories
	OrgNone = "none"

	// Repository permission levels for a user //

	// RepoPull allows only pull access to repository
	RepoPull = "pull"
	// RepoPush allows pull and push access to repository
	RepoPush = "push"
	// RepoFull allows full access to repository
	RepoFull = "admin"

	// Standard team names

	// TeachersTeam is the team with all teachers and teaching assistants of a course.
	TeachersTeam = "allteachers"
	// StudentsTeam is the team with all students of a course.
	StudentsTeam = "allstudents"
)

var (
	// ErrNotMember indicates that the requested organization exists, but the current user
	// is not its member.
	ErrNotMember = errors.New("user is not a member of the organization")
	// ErrNotOwner indicates that user has no admin rights in the requested organization.
	ErrNotOwner = errors.New("user is not an owner of the organization")
)

// Validators //

func (opt OrganizationOptions) valid() bool {
	return opt.Path != "" && opt.DefaultPermission != ""
}

func (opt GetOrgOptions) valid() bool {
	return opt.ID != 0 || opt.Name != ""
}

func (r Repository) valid() bool {
	return r.Path != "" && r.Owner != ""
}

func (opt AddTeamRepoOptions) valid() bool {
	return opt.TeamID > 0 &&
		opt.OrganizationID > 0 &&
		opt.Repo != "" &&
		opt.Owner != "" &&
		opt.Permission != ""
}

func (opt UpdateTeamOptions) valid() bool {
	return opt.TeamID > 0 && opt.OrganizationID > 0
}

func (opt CreateRepositoryOptions) valid() bool {
	return opt.Organization != nil && opt.Path != ""
}

func (opt CreateHookOptions) valid() bool {
	return opt.URL != "" &&
		(opt.Organization != "" || (opt.Repository != nil &&
			opt.Repository.valid()))
}

func (opt TeamOptions) valid() bool {
	return opt.TeamName != "" && opt.Organization != "" ||
		opt.TeamID > 0 && opt.OrganizationID > 0
}

func (opt NewTeamOptions) valid() bool {
	return opt.TeamName != "" && opt.Organization != ""
}

func (opt TeamMembershipOptions) valid() bool {
	return (opt.TeamID > 0 && opt.OrganizationID > 0 ||
		opt.TeamName != "" && opt.Organization != "") &&
		opt.Username != ""
}

func (opt OrgMembershipOptions) valid() bool {
	return opt.Organization != "" && opt.Username != ""
}

func (opt RepositoryOptions) valid() bool {
	return opt.ID > 0 || (opt.Path != "" && opt.Owner != "")
}

func (opt FileOptions) valid() bool {
	return opt.Owner != "" &&
		opt.Path != "" && opt.Repository != ""
}

// Errors //

// ErrNotSupported is returned when the source code management solution used
// does not provide a sufficient API for the method called.
type ErrNotSupported struct {
	SCM    string
	Method string
}

func (e ErrNotSupported) Error() string {
	return "method " + e.Method + " not supported by " + e.SCM + " SCM"
}

// ErrMissingFields is returned when scm struct validation fails.
// This error only used for development/debugging and never goes to frontend user.
type ErrMissingFields struct {
	Message string
	Method  string
}

func (e ErrMissingFields) Error() string {
	return "github method " + e.Method + " called with missing required fields: " + e.Message
}

// ErrFailedSCM is returned to provide detailed information
// to user about source of the error and possible solution
type ErrFailedSCM struct {
	Method   string
	Message  string
	GitError error
}

// Error message includes name of the failed method and the original error message
// from GitHub, to make it suitable for informative back-end logging
func (e ErrFailedSCM) Error() string {
	return "github method " + e.Method + " failed: " + e.GitError.Error() + "\n" + e.Message
}
