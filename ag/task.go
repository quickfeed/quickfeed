package ag

// HasChanged returns true if task t has a different title or body than the new task
func (t *Task) HasChanged(new *Task) bool {
	return t.Title != new.Title || t.Body != new.Body
}
