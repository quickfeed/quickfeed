package scm

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

func TestIDs(t *testing.T) {
	repos := make(map[uint64]*Repository)
	repos[1] = &Repository{ID: 1}
	repos[3] = &Repository{ID: 3}
	id := generateID(repos)
	if id != 2 {
		t.Errorf("expected id = 2, got %d", id)
	}
	repos[id] = &Repository{ID: id}
	id = generateID(repos)
	if id != 4 {
		t.Errorf("expected id = 4, got %d", id)
	}

	organizations := make(map[uint64]*qf.Organization)
	organizations[2] = &qf.Organization{}
	id = generateID(organizations)
	if id != 1 {
		t.Errorf("expected id = 1, got %d", id)
	}
}
