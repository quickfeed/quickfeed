package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Config struct {
	providers map[string]*oauth2.Config
}

// Config creates a new OAuth 2.0 config for every enabled OAuth 2.0 provider.
// Currently only works for GitHub.
func NewConfig(baseURL string) (*Config, error) {
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
	return &Config{
		providers: providers,
	}, nil
}

func (c *Config) get(provider string) (*oauth2.Config, error) {
	conf, ok := c.providers[provider]
	if !ok {
		return nil, fmt.Errorf("no configuration for provider %s", provider)
	}
	return conf, nil
}

func (c *Config) getRedirectURL(provider, secret string) string {
	conf, ok := c.providers[provider]
	if !ok {
		return ""
	}
	return conf.AuthCodeURL(secret)
}

func (c *Config) extractAccessToken(r *http.Request, provider, secret string) (string, error) {
	if err := r.ParseForm(); err != nil {
		return "", err
	}
	callbackSecret := r.FormValue("state")
	if callbackSecret != secret {
		return "", errors.New("incorrect callback secret")
	}
	code := r.FormValue("code")
	if code == "" {
		return "", errors.New("got empty code on callback")
	}
	providerConfig, err := c.get(provider)
	if err != nil {
		return "", err
	}
	authToken, err := providerConfig.Exchange(r.Context(), code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange access token: %s", err)
	}
	return authToken.AccessToken, nil
}
