package scm

import (
	"context"
	"testing"

	"github.com/beatlabs/github-auth/app"
	"github.com/quickfeed/quickfeed/internal/env"
)

// MockSCMManager sets the current provider to "fake", creates a "test" organization
// and a fake scm client for this organization.
func MockSCMManager(t *testing.T) (SCM, *Manager) {
	t.Helper()
	env.SetFakeProvider(t)
	conf := &Config{
		"test",
		"test",
		&app.Config{},
	}
	sc := NewFakeSCMClient()
	_, _ = sc.CreateOrganization(context.Background(), &OrganizationOptions{
		Name: "test",
		Path: "test",
	})
	return sc, &Manager{
		scms: map[string]SCM{
			"test": sc,
		},
		Config: conf,
	}
}
