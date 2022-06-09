package config

import (
	"log"
	"os"

	qfrand "github.com/autograde/quickfeed/internal/rand"
)

// Endpoints keeps all URL endpoints used by the server for user authentication,
// authorization and GitHub API interactions.
type Endpoints struct {
	BaseURL       string
	LoginURL      string
	CallbackURL   string
	LogoutURL     string
	WebhookURL    string
	InstallAppURL string
	Public        string
	PortNumber    string
}

// Secrets keeps secrets that have been generated.
// or read from the environment
type Secrets struct {
	WebhookSecret  string
	CallbackSecret string
	TokenSecret    string
	key            *[32]byte
}

type Paths struct {
	CertPath    string
	CertKeyPath string
	AppKeyPath  string
}

// Config keeps all configuration information in one place.
type Config struct {
	Endpoints *Endpoints
	Secrets   *Secrets
	Paths     *Paths
}

func NewConfig(baseURL, public, portNumber string) *Config {
	log.Printf("Making new config: base URL (%s), public (%s), httpAddr (%s)", baseURL, public, portNumber) // tmp
	conf := &Config{
		Endpoints: &Endpoints{
			BaseURL:       baseURL,
			Public:        public, // filepath.Join(public, indexFile),
			PortNumber:    portNumber,
			LoginURL:      Login,
			LogoutURL:     Logout,
			WebhookURL:    Webhook,
			CallbackURL:   Callback,
			InstallAppURL: Install,
		},
		Secrets: &Secrets{
			WebhookSecret:  os.Getenv(WebhookEnv),
			CallbackSecret: qfrand.String(),
			TokenSecret:    os.Getenv(TokenKeyEnv),
		},
		Paths: &Paths{
			CertPath:    os.Getenv(CertEnv),
			CertKeyPath: os.Getenv(CertKeyEnv),
		},
	}
	return conf
}
