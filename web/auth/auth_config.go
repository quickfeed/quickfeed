package auth

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AuthConfig struct {
	providers map[string]*oauth2.Config
}

// AuthConfig creates a new OAuth 2.0 config for every enabled OAuth 2.0 provider.
// Currently only works for GitHub.
func NewAuthConfig(baseURL string) (*AuthConfig, error) {
	providers := make(map[string]*oauth2.Config)

	// Enable GitHub.
	clientID := os.Getenv("GITHUB_KEY")
	clientSecret := os.Getenv("GITHUB_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing GitHub client variables")
	}
	githubConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  GetCallbackURL(baseURL, "github"),
		Scopes:       []string{"repo:invite"},
	}
	providers["github"] = githubConfig
	return &AuthConfig{
		providers: providers,
	}, nil
}
