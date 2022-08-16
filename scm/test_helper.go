package scm

import (
	"context"

	"github.com/beatlabs/github-auth/app"
)

func TestSCMManager() *Manager {
	conf := &Config{
		"test",
		"test",
		&app.Config{},
	}
	sc := NewFakeSCMClient()
	sc.CreateOrganization(context.Background(), &OrganizationOptions{
		Name: "testorg",
		Path: "testorg",
	})
	scms := make(map[string]SCM)
	scms["testorg"] = sc
	return &Manager{
		scms:   scms,
		Config: conf,
	}
}
