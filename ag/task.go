package ag

// TODO(Espeland): Add method description, and test
func (t *Task) HasChanged(new *Task) bool {
	return t.Title != new.Title || t.Body != new.Body
}
