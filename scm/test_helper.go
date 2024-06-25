package scm

import (
	"testing"

	"github.com/beatlabs/github-auth/app"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

// MockSCMManager sets the current provider to "fake", and initializes the
// MockedGithubSCMClient based on the provided mock options, which can be
// used to mock different scenarios (course organizations and repositories).
// Two options are available: WithMockOrgs() and WithMockCourses().
func MockSCMManager(t *testing.T, opts ...MockOption) (SCM, *Manager) {
	t.Helper()
	env.SetFakeProvider(t)
	sc := NewMockedGithubSCMClient(qtest.Logger(t), opts...)
	return sc, &Manager{
		scms:   map[string]SCM{qtest.MockOrg: sc},
		Config: &Config{"qfClientID", "qfClientSecret", &app.Config{}},
	}
}
