package models

// User represents a user account.
type User struct {
	ID uint64 `json:"id"`

	IsAdmin bool `json:"isadmin"`

	Name      string `json:"name"`
	StudentID string `json:"studentid"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarurl"`

	RemoteIdentities []*RemoteIdentity `json:"remoteidentities,omitempty"`

	Enrollments []*Enrollment
}

// RemoteIdentity represents a third-party identity which can be attached to a
// user account.
type RemoteIdentity struct {
	ID uint64 `json:"id"`

	// TODO: Provider + RemoteID = key.
	Provider string `json:"provider"`
	RemoteID uint64 `json:"remoteid"`

	AccessToken string `json:"-"`

	UserID uint64 `json:"userid"`
}
