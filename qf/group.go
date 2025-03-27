package qf

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
		if user.GetID() == u.GetID() {
			return true
		}
	}
	return false
}

// ContainsAll compares group members
func (g *Group) ContainsAll(group *Group) bool {
	if len(g.GetUsers()) != len(group.GetUsers()) {
		return false
	}
	for _, u := range group.GetUsers() {
		if !g.Contains(u) {
			return false
		}
	}
	return true
}

// GetUsersExcept returns a list of all users in a group, except the one with the given userID.
func (g *Group) GetUsersExcept(userID uint64) []*User {
	var subset []*User
	for _, user := range g.GetUsers() {
		if user.GetID() == userID {
			continue
		}
		subset = append(subset, user)
	}
	return subset
}

// UserIDs returns the user IDs of this group.
func (g *Group) UserIDs() []uint64 {
	userIDs := make([]uint64, 0, len(g.GetUsers()))
	for _, user := range g.GetUsers() {
		userIDs = append(userIDs, user.GetID())
	}
	return userIDs
}

// Dummy implementation of the interceptor.userIDs interface.
// Marks this message type to be evaluated for token refresh.
func (*GroupRequest) UserIDs() []uint64 {
	return []uint64{}
}
