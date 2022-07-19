package env

import (
	"errors"
	"os"
	"testing"
)

const (
	defaultProvider = "github"
)

var (
	provider     string
	clientKey    string
	clientSecret string
)

func init() {
	provider = os.Getenv("QUICKFEED_SCM_PROVIDER")
	if provider == "" {
		provider = defaultProvider
	}
	clientKey = os.Getenv("QUICKFEED_CLIENT_KEY")
	clientSecret = os.Getenv("QUICKFEED_CLIENT_SECRET")
}

// ScmProvider returns the current SCM provider supported by this backend.
func ScmProvider() string {
	return provider
}

// ClientKey returns client ID for the current SCM provider.
func ClientKey() (string, error) {
	if clientKey == "" {
		return "", errors.New("missing client ID for SCM provider")
	}
	return clientKey, nil
}

// ClientSecret returns secret for the current SCM provider.
func ClientSecret() (string, error) {
	if clientSecret == "" {
		return "", errors.New("missing client secret for SCM provider")
	}
	return clientSecret, nil
}

// SetFakeProvider sets the provider to fake. This is only for testing.
// The t argument is added as a reminder that this is only for testing.
func SetFakeProvider(t *testing.T) {
	t.Helper()
	provider = "fake"
}
