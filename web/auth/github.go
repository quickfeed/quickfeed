package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// newGitHubConfig creates a new OAuth config for GitHub.
func NewGitHubConfig(baseURL, clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  GetCallbackURL(baseURL),
		Scopes:       []string{"repo:invite"},
	}
}
