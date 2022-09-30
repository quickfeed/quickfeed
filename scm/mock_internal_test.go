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

	teams := make(map[uint64]*Team)
	id = generateID(teams)
	if id != 1 {
		t.Errorf("expected id = 1, got %d", id)
	}
	teams[id] = &Team{}
	id = generateID(teams)
	if id != 2 {
		t.Errorf("expected id = 2, got %d", id)
	}

	organizations := make(map[uint64]*qf.Organization)
	organizations[2] = &qf.Organization{}
	id = generateID(organizations)
	if id != 1 {
		t.Errorf("expected id = 1, got %d", id)
	}

	hooks := make(map[uint64]*Hook)
	for i := 1; i <= 10; i++ {
		hooks[uint64(i)] = &Hook{ID: uint64(i)}
	}
	id = generateID(hooks)
	if id != 11 {
		t.Errorf("expected id = 11, got %d", id)
	}
}
