package auth

import (
	"github.com/quickfeed/quickfeed/scm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// NewGitHubConfig creates a new OAuth config for GitHub.
func NewGitHubConfig(c *scm.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  GetCallbackURL(),
		Scopes:       []string{"repo:invite", "write:org"},
	}
}
