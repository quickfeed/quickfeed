package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/autograde/quickfeed/internal/rand"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/term"
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
	Key            []byte
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
			CallbackSecret: rand.String(),
			TokenSecret:    os.Getenv(TokenKeyEnv),
		},
		Paths: &Paths{
			CertPath:    os.Getenv(CertEnv),
			CertKeyPath: os.Getenv(CertKeyEnv),
		},
	}
	return conf
}

// ReadKey asks for a passphrase to decrypt the master key that will be used
// to encrypt access tokens.
func (c *Config) ReadKey() error {
	fmt.Println("Key: ")
	input, err := term.ReadPassword(0)
	if err != nil {
		return err
	}
	passphrase, err := base64.RawStdEncoding.DecodeString(string(input))
	if err != nil {
		return err
	}
	if len(passphrase) != 32 {
		return fmt.Errorf("wrong key length, expected 32, got %d", len(passphrase))
	}
	path := os.Getenv(KeyEnv)
	if path == "" {
		return fmt.Errorf("missing file path")
	}
	keyfile, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	passkey := new([32]byte)
	nonce := new([24]byte)
	copy(passkey[:], passphrase[:32])
	copy(nonce[:], keyfile[:24])

	key, ok := secretbox.Open(nil, keyfile[24:], nonce, passkey)
	if !ok {
		return fmt.Errorf("error decrypting the key")
	}
	c.Secrets.Key = key
	return nil
}
