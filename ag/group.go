package ag

import "reflect"

// UserNames returns the SCM user names of the group.
func (g *Group) UserNames() []string {
	var gitUserNames []string
	for _, user := range g.GetUsers() {
		gitUserNames = append(gitUserNames, user.GetLogin())
	}
	return gitUserNames
}

// Contains returns true if the given user is in the group.
func (g *Group) Contains(user *User) bool {
	for _, u := range g.GetUsers() {
		if user.ID == u.ID {
			return true
		}
	}
	return false
}

// ContainsAll compares group members
func (g *Group) ContainsAll(group *Group) bool {
	return reflect.DeepEqual(g.Users, group.Users)
}

// SetSlipDays sets number of remaining slip days for each enrollment
func (g *Group) SetSlipDays(c *Course) {
	for _, e := range g.Enrollments {
		e.SetSlipDays(c)
	}
}

// GetUsersExcept returns a list of all users in a group, except the one with the given userID.
func (g *Group) GetUsersExcept(userID uint64) []*User {
	subset := []*User{}
	for _, user := range g.Users {
		if user.GetID() == userID {
			continue
		}
		subset = append(subset, user)
	}
	return subset
}
