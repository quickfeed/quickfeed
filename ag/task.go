package ag

// HasChanged returns true if a task has a different body or title than another.
func (t *Task) HasChanged(new *Task) bool {
	return t.Title != new.Title || t.Body != new.Body
}
