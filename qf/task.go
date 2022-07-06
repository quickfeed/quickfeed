package qf

import "strings"

const (
	// Message added to the body of an issue when closing it, since there is no support for deleting issues.
	deleteMsg = "\n**The task associated with this issue has been deleted by the teaching staff.**\n"
	// Prefix added to the title of a deleted task.
	deleted = "DELETED"
)

// HasChanged returns true if task t has a different title or body than the new task.
func (t *Task) HasChanged(newTask *Task) bool {
	return t.Title != newTask.Title || t.Body != newTask.Body
}

func (t *Task) MarkDeleted() {
	t.Title = deleted + " " + t.Title
	t.Body = deleteMsg + t.Body
}

func (t *Task) IsDeleted() bool {
	return strings.HasPrefix(t.Title, deleted)
}
