package qf

// RemoveRemoteID removes user's remote identity before transmitting to client.
func (u *User) RemoveRemoteID() {
	if u != nil {
		u.RefreshToken = ""
		u.ScmRemoteID = 0
	}
}

// RemoveRemoteID nullifies remote identities of all users
func (u *Users) RemoveRemoteID() {
	for _, user := range u.GetUsers() {
		user.RemoveRemoteID()
	}
}

// RemoveRemoteID nullifies remote identities of all users in a group
func (g *Group) RemoveRemoteID() {
	for _, user := range g.GetUsers() {
		user.RemoveRemoteID()
	}
	for _, enrollment := range g.GetEnrollments() {
		enrollment.RemoveRemoteID()
	}
}

// RemoveRemoteID nullifies remote identities of all users in every group
func (g *Groups) RemoveRemoteID() {
	for _, group := range g.GetGroups() {
		group.RemoveRemoteID()
	}
}

// RemoveRemoteID removes remote identity of the enrolled user
func (e *Enrollment) RemoveRemoteID() {
	e.GetUser().RemoveRemoteID()
	e.GetGroup().RemoveRemoteID()
	e.GetCourse().RemoveRemoteID()
}

// RemoveRemoteID removes remote identities for every enrollment
func (e *Enrollments) RemoveRemoteID() {
	for _, enr := range e.GetEnrollments() {
		enr.RemoveRemoteID()
	}
}

// RemoveRemoteID removes remote identities for all course groups and enrollments
func (c *Course) RemoveRemoteID() {
	if c != nil {
		c.DockerfileDigest = ""
	}
	for _, enr := range c.GetEnrollments() {
		enr.RemoveRemoteID()
	}
	for _, grp := range c.GetGroups() {
		grp.RemoveRemoteID()
	}
}

// RemoveRemoteID removes remote identities for groups and enrollments in every course
func (c *Courses) RemoveRemoteID() {
	for _, crs := range c.GetCourses() {
		crs.RemoveRemoteID()
	}
}
