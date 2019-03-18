package models

// User represents a user account.
type User struct {
	ID uint64 `json:"id"`

	IsAdmin *bool `json:"isadmin"`

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

	Provider string `json:"provider" gorm:"unique_index:uid_provider_remote_id"`
	RemoteID uint64 `json:"remoteid" gorm:"unique_index:uid_provider_remote_id"`

	AccessToken string `json:"-"`

	UserID uint64 `json:"userid"`
}

// GetRemoteIDFor returns the user's remote identity for the given provider.
// If no remote identity for the given provider is found, then nil is returned.
func (user *User) GetRemoteIDFor(provider string) *RemoteIdentity {
	var remoteID *RemoteIdentity
	for _, v := range user.RemoteIdentities {
		if v.Provider == provider {
			remoteID = v
			break
		}
	}
	return remoteID
}

// IAdmin returns true only if this user is admin.
func (user *User) IAdmin() bool {
	return user.IsAdmin != nil && *user.IsAdmin
}
