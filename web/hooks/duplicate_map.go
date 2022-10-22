package hooks

import "sync"

// Duplicates is a map of active/duplicate commit IDs.
type Duplicates struct {
	dup map[string]struct{} // map of duplicate events: CommitID -> struct{}
	mu  sync.Mutex          // protects dup
}

// NewDuplicateMap creates a new DuplicateMap.
func NewDuplicateMap() *Duplicates {
	return &Duplicates{
		dup: make(map[string]struct{}),
	}
}

// Duplicate returns true if the commitID is a duplicate.
func (dm *Duplicates) Duplicate(commitID string) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if _, ok := dm.dup[commitID]; ok {
		return true
	}
	dm.dup[commitID] = struct{}{}
	return false
}

// Remove removes the commitID from the duplicate map to avoid ever growing map.
// Should be called after the push event has been processed.
func (dm *Duplicates) Remove(commitID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.dup, commitID)
}
