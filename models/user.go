package models

// User represents a user account.
type User struct {
	ID uint64

	RemoteIdentities []RemoteIdentity
}

// RemoteIdentity represents a third-party identity which can be attached to a
// user account.
type RemoteIdentity struct {
	ID uint64

	// TODO: Provider + RemoteID = key.
	Provider string
	RemoteID uint64

	AccessToken string

	UserID uint64
}
