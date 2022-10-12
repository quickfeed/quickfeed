package scm

// InvitationOptions contains information on which organization and user to accept invitations for.
type InvitationOptions struct {
	Login string // GitHub username.
	Owner string // Name of the organization.
	Token string // Access token for the user.
}
