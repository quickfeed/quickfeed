package ag

// IsOwner returns true if the current user is the same as the given user ID.
func (u *User) IsOwner(userID uint64) bool {
	return u.GetID() == userID
}
