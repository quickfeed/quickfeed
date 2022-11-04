package scm

import (
	"testing"

	"github.com/beatlabs/github-auth/app"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

// MockSCMManager sets the current provider to "fake", creates a "test" organization
// and a fake scm client for this organization.
func MockSCMManager(t *testing.T, withCourse bool) (SCM, *Manager) {
	t.Helper()
	env.SetFakeProvider(t)
	conf := &Config{
		"qfClientID",
		"qfClientSecret",
		&app.Config{},
	}
	sc := NewMockSCMClient(withCourse)
	return sc, &Manager{
		scms: map[string]SCM{
			qtest.MockOrg: sc,
		},
		Config: conf,
	}
}
