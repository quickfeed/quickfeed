package scm_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestIDs(t *testing.T) {
	repos := make(map[uint64]*scm.Repository)
	repos[1] = &scm.Repository{ID: 1}
	repos[3] = &scm.Repository{ID: 3}
	id := scm.GenerateID(repos)
	if id != 2 {
		t.Errorf("expected id = 2, got %d", id)
	}
	repos[id] = &scm.Repository{ID: id}
	id = scm.GenerateID(repos)
	if id != 4 {
		t.Errorf("expected id = 4, got %d", id)
	}

	teams := make(map[uint64]*scm.Team)
	id = scm.GenerateID(teams)
	if id != 1 {
		t.Errorf("expected id = 1, got %d", id)
	}
	teams[id] = &scm.Team{}
	id = scm.GenerateID(teams)
	if id != 2 {
		t.Errorf("expected id = 2, got %d", id)
	}

	organizations := make(map[uint64]*qf.Organization)
	organizations[2] = &qf.Organization{}
	id = scm.GenerateID(organizations)
	if id != 1 {
		t.Errorf("expected id = 1, got %d", id)
	}

	hooks := make(map[uint64]*scm.Hook)
	for i := 1; i <= 10; i++ {
		hooks[uint64(i)] = &scm.Hook{ID: uint64(i)}
	}
	id = scm.GenerateID(hooks)
	if id != 11 {
		t.Errorf("expected id = 11, got %d", id)
	}

}
