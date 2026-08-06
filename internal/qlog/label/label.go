// Package label defines canonical structured logging attribute keys.
//
// Use these keys for fields that are queried or shared across backend packages;
// keep component-specific or one-off attribute names local to the component.
//
// For RPC handlers, the logging interceptors already attach RPCMethod, and,
// once access control has accepted the request, UserID for the calling user and
// CourseID for the requested course. Handlers must not repeat those attributes,
// since slog records every attribute it is given, including duplicate keys.
// When a handler acts on some other user than the caller, use TargetUser and
// TargetUserID to keep the two apart.
package label

const (
	Assignment     = "assignment"
	AssignmentID   = "assignment_id"
	Branch         = "branch"
	Code           = "code" // Connect error code of a completed RPC.
	Commit         = "commit"
	CourseCode     = "course_code"
	CourseID       = "course_id"
	Duration       = "duration"
	Error          = "error"
	Group          = "group"
	GroupID        = "group_id"
	Organization   = "organization" // SCM organization name.
	Owner          = "owner"        // SCM owner (organization or user) of a repository.
	Path           = "path"
	PullRequest    = "pull_request"
	RemoteID       = "remote_id" // SCM user ID, i.e., User.ScmRemoteID; distinct from UserID.
	Repository     = "repository"
	RepositoryType = "repository_type"
	RPCMethod      = "rpc_method"
	SubmissionID   = "submission_id"
	TargetUser     = "target_user"    // SCM login of the user acted upon.
	TargetUserID   = "target_user_id" // QuickFeed database ID of the user acted upon.
	User           = "user"           // SCM login or another human-readable user identifier.
	UserID         = "user_id"        // QuickFeed database user ID.
)
