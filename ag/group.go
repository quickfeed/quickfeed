package ag

// UserNames returns the SCM user names of the group.
func (g *Group) UserNames() []string {
	var gitUserNames []string
	for _, user := range g.GetUsers() {
		gitUserNames = append(gitUserNames, user.GetLogin())
	}
	return gitUserNames
}
