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
	return &Manager{
		scms: map[string]SCM{
			"testorg": sc,
		},
		Config: conf,
	}
}
