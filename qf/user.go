package qf

// IsOwner returns true if the current user is the same as the given user ID.
func (u *User) IsOwner(userID uint64) bool {
	return u.GetID() == userID
}

// UserIDs returns the user ID of this user.
func (u *User) UserIDs() []uint64 {
	return []uint64{u.GetID()}
}
