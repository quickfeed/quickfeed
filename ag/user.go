package ag

import fmt "fmt"

// IsOwner returns true if the current user is the same as the given user ID.
func (u *User) IsOwner(userID uint64) bool {
	return u.GetID() == userID
}

// GetRemoteIDFor returns the user's remote identity for the given provider.
// If no remote identity for the given provider is found, then nil is returned.
func (u *User) GetRemoteIDFor(provider string) *RemoteIdentity {
	var remoteID *RemoteIdentity
	for _, v := range u.RemoteIdentities {
		if v.Provider == provider {
			remoteID = v
			break
		}
	}
	return remoteID
}

// GetAccessToken returns the user's access token for the given provider.
func (u *User) GetAccessToken(provider string) (string, error) {
	remoteID := u.GetRemoteIDFor(provider)
	if remoteID == nil {
		return "", fmt.Errorf("found no %s access token for user %s", provider, u.GetName())
	}
	return remoteID.GetAccessToken(), nil
}

// SetSlipDays sets number of remaining slip days for each enrollment
func (u *User) SetSlipDays(c *Course) {
	for _, e := range u.Enrollments {
		e.SetSlipDays(c)
	}
}
