package models

// User represents a user account.
type User struct {
	ID uint64

	RemoteIdentities []RemoteIdentity `json:"remoteidentities,omitempty"`
}

// RemoteIdentity represents a third-party identity which can be attached to a
// user account.
type RemoteIdentity struct {
	ID uint64

	// TODO: Provider + RemoteID = key.
	Provider string
	RemoteID uint64

	AccessToken string `json:"-"`

	UserID uint64
}
