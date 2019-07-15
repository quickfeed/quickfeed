package ag

// RemoveRemoteID removes user's remote identity before transmitting to client.
func (u *User) RemoveRemoteID() {
	if u != nil {
		voidIDs := make([]*RemoteIdentity, 0)
		u.RemoteIdentities = voidIDs
		for _, enrollment := range u.GetEnrollments() {
			if enrollment.User != nil && enrollment.User.RemoteIdentities != nil {
				enrollment.User.RemoteIdentities = voidIDs
			}
		}
	}
}

// RemoveRemoteIDs nullifies remote identities of all users
func (u *Users) RemoveRemoteIDs() {
	for _, user := range u.GetUsers() {
		user.RemoveRemoteID()
	}
}

// RemoveRemoteIDs nullifies remote identities of all users in a group
func (g *Group) RemoveRemoteIDs() {
	if g != nil {
		for _, user := range g.GetUsers() {
			user.RemoveRemoteID()
		}
		for _, enrollment := range g.GetEnrollments() {
			enrollment.RemoveRemoteID()
		}
	}
}

// RemoveRemoteIDs nullifies remote identities of all users in every group
func (g *Groups) RemoveRemoteIDs() {
	for _, group := range g.GetGroups() {
		group.RemoveRemoteIDs()
	}
}

// RemoveRemoteID removes remote identity of the enrolled user
func (e *Enrollment) RemoveRemoteID() {
	if e != nil && e.User != nil {
		e.User.RemoveRemoteID()
	}
}

// RemoveRemoteIDs removes remote identities for every enrollment
func (e *Enrollments) RemoveRemoteIDs() {
	for _, enr := range e.GetEnrollments() {
		enr.RemoveRemoteID()
	}
}
