package ag

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
