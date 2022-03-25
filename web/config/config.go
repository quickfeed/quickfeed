package config

import (
	"os"

	"github.com/autograde/quickfeed/internal/rand"
)

const (
	// Environmental variables
	WebhookEnv = "WEBHOOK_SECRET"
	AppEnv     = "APP_ID"
	KeyEnv     = "APP_KEY"
	SecretEnv  = "APP_SECRET"

	// Endpoints
	GitHubUser       = "https://api.github.com/user"
	Install          = "https://github.com/apps/appth-gh" // TODO: change to the real URL (or better yet read from a config file)
	InstallationsAPI = "https://api.github.com/app/installations"
	Login            = "/login"
	Logout           = "/logout"
	Callback         = "/callback"
	Webhook          = "/hook/github/events"
)

// Endpoints keeps all URL endpoints used by the server for user authentication,
// authorization and GitHub API interactions
type Endpoints struct {
	BaseURL       string
	LoginURL      string
	CallbackURL   string
	LogoutURL     string
	GithubUserURL string
	WebhookURL    string
	InstallAppURL string
}

// Secrets keeps locally generated secrets
type Secrets struct {
	WebhookSecret string
	OAuth2Code    string
}

// Config keeps all configuration information in one place
type Config struct {
	Endpoints      *Endpoints
	Secrets        *Secrets
	TokensToUpdate *TokenManager // TODO: not sure if this and app belongs here
	App            *GithubApp
}

// TokenManager keeps track of UserIDs for token updates
type TokenManager []uint64

func NewConfig(baseURL string) *Config {
	conf := &Config{
		Endpoints: &Endpoints{
			BaseURL:       baseURL,
			LoginURL:      Login,
			LogoutURL:     Logout,
			CallbackURL:   Callback,
			GithubUserURL: GitHubUser,
			InstallAppURL: Install,
		},
		Secrets: &Secrets{
			WebhookSecret: os.Getenv(WebhookEnv),
			OAuth2Code:    rand.String(),
		},
	}
	return conf
}
