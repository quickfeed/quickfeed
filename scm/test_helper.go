package scm

import (
	"testing"

	"github.com/beatlabs/github-auth/app"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

// MockSCMManager sets the current provider to "fake", creates a "test" organization
// and a fake scm client for this organization.
func MockSCMManager(t *testing.T) (SCM, *Manager) {
	t.Helper()
	env.SetFakeProvider(t)
	conf := &Config{
		"qfClientID",
		"qfClientSecret",
		&app.Config{},
	}
	sc := NewMockSCMClient()
	return sc, &Manager{
		scms: map[string]SCM{
			qtest.MockOrg: sc,
		},
		Config: conf,
	}
}

// MockSCMManagerWithCourse sets provider to "fake", creates a mock organization
// and sets up default course repositories and teams for this organization.
func MockSCMManagerWithCourse(t *testing.T) (SCM, *Manager) {
	t.Helper()
	env.SetFakeProvider(t)
	conf := &Config{
		"qfClientID",
		"qfClientSecret",
		&app.Config{},
	}
	sc := NewMockSCMClientWithCourse()
	return sc, &Manager{
		scms: map[string]SCM{
			qtest.MockOrg: sc,
		},
		Config: conf,
	}
}
