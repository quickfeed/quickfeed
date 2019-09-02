package scm

const (
	// Organization roles //

	// OrgOwner is organization owner
	OrgOwner = "admin"
	// OrgMember is organization member
	OrgMember = "member"

	// Team roles //

	// TeamMaintainer can add and delete team users and repos
	TeamMaintainer = "maintainer"
	// TeamMember is a regular member
	TeamMember = "member"

	// Repository permission levels for organization //

	// OrgPull allows only pull access to organization repositories
	OrgPull = "read"
	// OrgPush allows pull and push acces to organization repositories
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

// Validators //

func (opt CreateOrgOptions) valid() bool {
	return opt.Path != "" && opt.DefaultPermission != ""
}

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
		opt.URL != ""
}

func (opt OrgHookOptions) valid() bool {
	return opt.Organization != nil &&
		opt.Organization.IsValid() &&
		opt.URL != ""
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

func (opt RepositoryOptions) valid() bool {
	return opt.ID > 0 ||
		(opt.Path != "" && opt.Owner != "")
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
// TODO(vera): this error can be a struct just as ErrNotSupported with method and interface fields
// then we cn skip logging it every time it occures. The question is if it even is reasonable to pass a failng struct back to
// ag_service for logging when it can be logged right here
type ErrMissingFields struct {
	Message string
	Method  string
}

func (e ErrMissingFields) Error() string {
	return "github method " + e.Method + " got argument with some of required fields missing: " + e.Message
}

// = fmt.Errorf("invalid argument: missing required fields")
