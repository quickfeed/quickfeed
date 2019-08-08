package scm

import "fmt"

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

// Validators //

func (opt CreateOrgOptions) valid() bool {
	return opt.Path != "" && opt.DefaultPermission != ""
}

// validForHooks checks that repository object can be used in hooks related methods
func (r Repository) valid() bool {
	return r.Path != "" && r.Owner != ""
}

func (opt AddTeamRepoOptions) valid() bool {
	return opt.TeamID > 0 &&
		opt.Repo != "" &&
		opt.Owner != "" &&
		opt.Permission != ""
}

func (opt CreateRepositoryOptions) valid() bool {
	return opt.Organization != nil && opt.Path != ""
}

func (opt CreateHookOptions) valid() bool {
	return opt.Repository != nil &&
		opt.Repository.valid() &&
		opt.URL != "" &&
		opt.Secret != ""
}

func (opt CreateTeamOptions) validWithOrg() bool {
	return opt.Organization != nil &&
		opt.Organization.IsValid() &&
		opt.valid()
}

func (opt CreateTeamOptions) valid() bool {
	return opt.TeamName != "" || opt.TeamID > 0
}

func (opt TeamMembershipOptions) valid() bool {
	return opt.Organization != nil &&
		opt.Organization.IsValid() &&
		(opt.TeamID > 0 || opt.TeamSlug != "")
}

func (opt OrgMembershipOptions) valid() bool {
	return opt.Organization != nil &&
		opt.Organization.IsValid() &&
		opt.Username != ""
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
var ErrMissingFields = fmt.Errorf("invalid argument: missing required fields")
