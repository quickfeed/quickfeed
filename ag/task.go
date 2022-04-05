package ag

// HasChanged returns true if task t has a different title or body than the new task.
func (t *Task) HasChanged(newTask *Task) bool {
	return t.Title != newTask.Title || t.Body != newTask.Body
}
