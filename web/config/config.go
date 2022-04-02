package config

import (
	"os"
	"path/filepath"

	"github.com/autograde/quickfeed/internal/rand"
)

const (
	// Environmental variables
	WebhookEnv = "WEBHOOK_SECRET"
	JWTKeyEnv  = "JWT_KEY" // TODO: where to store? Or reuse another secret?

	// Endpoints
	GitHubUser       = "https://api.github.com/user"
	Install          = "https://github.com/apps/appth-gh" // TODO: change to the real URL (or better yet read from a config file)
	InstallationsAPI = "https://api.github.com/app/installations"
	Login            = "/login"
	Logout           = "/logout"
	Callback         = "/callback"
	Webhook          = "/hook/github/events"

	// Paths //TODO: read from env
	appKeyPath = "./appth.private-key.pem"
	pemPath    = "cert/server.crt"
	keyPath    = "cert/server.key"
	indexFile  = "index.html"
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
	Public        string
	HttpAddress   string
}

// Secrets keeps secrets that have been generated
// or read from the environment
type Secrets struct {
	WebhookSecret  string
	CallbackSecret string
	TokenSecret    string
}

type Paths struct {
	PemPath    string
	KeyPath    string
	AppKeyPath string
}

// Config keeps all configuration information in one place
type Config struct {
	Endpoints      *Endpoints
	Secrets        *Secrets
	Paths          *Paths
	TokensToUpdate *TokenManager // TODO: not sure if this and app belongs here
	App            *GithubApp
}

// TokenManager keeps track of UserIDs for token updates
type TokenManager []uint64

func NewConfig(baseURL, public, httpAddr string) *Config {
	conf := &Config{
		Endpoints: &Endpoints{
			BaseURL:       baseURL,
			Public:        filepath.Join(public, indexFile),
			HttpAddress:   httpAddr,
			LoginURL:      Login,
			LogoutURL:     Logout,
			CallbackURL:   Callback,
			GithubUserURL: GitHubUser,
			InstallAppURL: Install,
		},
		Secrets: &Secrets{
			WebhookSecret:  os.Getenv(WebhookEnv),
			CallbackSecret: rand.String(),
			TokenSecret:    os.Getenv(JWTKeyEnv),
		},
		Paths: &Paths{
			AppKeyPath: appKeyPath,
			PemPath:    pemPath,
			KeyPath:    keyPath,
		},
	}
	return conf
}
