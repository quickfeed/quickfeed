package scm

import (
	"testing"

	"github.com/beatlabs/github-auth/app"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

// MockManager sets the current provider to "fake", and initializes the
// MockedGithubSCMClient based on the provided mock options, which can be
// used to mock different scenarios (course organizations and repositories).
// Two options are available: WithMockOrgs() and WithMockCourses().
func MockManager(t *testing.T, opts ...MockOption) *Manager {
	t.Helper()
	env.SetFakeProvider(t)
	sc := NewMockedGithubSCMClient(qtest.Logger(t), opts...)
	// We reuse the same github mock client for all organizations.
	scms := make(map[string]SCM)
	for _, org := range sc.orgs {
		scms[*org.Login] = sc
	}
	return &Manager{
		scms:   scms,
		Config: &Config{"qfClientID", "qfClientSecret", &app.Config{}},
	}
}
