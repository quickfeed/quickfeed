package env

import (
	"fmt"
	"os"
	"testing"
)

const (
	defaultProvider = "github"
)

var (
	provider     string
	clientID     string
	clientSecret string
)

func init() {
	provider = os.Getenv("QUICKFEED_SCM_PROVIDER")
	if provider == "" {
		provider = defaultProvider
	}
	clientID = os.Getenv("QUICKFEED_CLIENT_ID")
	clientSecret = os.Getenv("QUICKFEED_CLIENT_SECRET")
}

// ScmProvider returns the current SCM provider supported by this backend.
func ScmProvider() string {
	return provider
}

// ClientID returns the client ID for the current SCM provider.
func ClientID() (string, error) {
	if clientID == "" {
		return "", fmt.Errorf("missing client ID for %s", provider)
	}
	return clientID, nil
}

// ClientSecret returns the client secret for the current SCM provider.
func ClientSecret() (string, error) {
	if clientSecret == "" {
		return "", fmt.Errorf("missing client secret for %s", provider)
	}
	return clientSecret, nil
}

// SetFakeProvider sets the provider to fake. This is only for testing.
// The t argument is added as a reminder that this is only for testing.
func SetFakeProvider(t *testing.T) {
	t.Helper()
	provider = "fake"
}
