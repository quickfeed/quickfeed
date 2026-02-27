package qf

// These methods implement the ID provider interfaces used for authorization and access control.

// GetCourseID returns the course ID.
func (r *Course) GetCourseID() uint64 {
	return r.GetID()
}

// GetUserID returns the user ID.
func (r *User) GetUserID() uint64 {
	return r.GetID()
}

// GetGroupID returns the group ID.
func (r *Group) GetGroupID() uint64 {
	return r.GetID()
}
