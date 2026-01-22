package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	defaultProvider       = "github"
	defaultAppName        = "QuickFeed"
	defaultAppKeyFile     = "quickfeed.pem"
	defaultGitHubDir      = "github"
	defaultRepositoryPath = "$HOME/courses"
)

var provider string

func init() {
	provider = os.Getenv("QUICKFEED_SCM_PROVIDER")
	if provider == "" {
		provider = defaultProvider
	}
}

// ScmProvider returns the current SCM provider supported by this backend.
func ScmProvider() string {
	return provider
}

func RepositoryPath() string {
	repositoryPath := os.Getenv("QUICKFEED_REPOSITORY_PATH")
	if repositoryPath == "" {
		repositoryPath = defaultRepositoryPath
	}
	return os.ExpandEnv(repositoryPath)
}

// ClientID returns the client ID for the current SCM provider.
func ClientID() (string, error) {
	clientID := os.Getenv("QUICKFEED_CLIENT_ID")
	if clientID == "" {
		return "", fmt.Errorf("missing client ID for %s", provider)
	}
	return clientID, nil
}

// ClientSecret returns the client secret for the current SCM provider.
func ClientSecret() (string, error) {
	clientSecret := os.Getenv("QUICKFEED_CLIENT_SECRET")
	if clientSecret == "" {
		return "", fmt.Errorf("missing client secret for %s", provider)
	}
	return clientSecret, nil
}

// AppID returns the application ID for the current SCM provider.
func AppID() (string, error) {
	appID := os.Getenv("QUICKFEED_APP_ID")
	if appID == "" {
		return "", fmt.Errorf("missing application ID for provider %s", provider)
	}
	return appID, nil
}

// AppKey returns path to the file with .pem private key.
// For GitHub apps a key must be generated on the App's
// settings page and saved into a file.
// If QUICKFEED_APP_KEY is not set, the default path $HOME/.config/quickfeed/github/quickfeed.pem is used.
func AppKey() string {
	appKey := os.Getenv("QUICKFEED_APP_KEY")
	if appKey == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to source tree if home directory is not available
			return filepath.Join(Root(), defaultConfigDir, defaultGitHubDir, defaultAppKeyFile)
		}
		appKey = filepath.Join(home, defaultConfigDir, defaultGitHubDir, defaultAppKeyFile)
	}
	return appKey
}

// AppName returns the name of the QuickFeed app on GitHub.
func AppName() string {
	appName := os.Getenv("QUICKFEED_APP_NAME")
	if appName == "" {
		return defaultAppName
	}
	return appName
}

func GetAccessToken() (string, error) {
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) == 0 {
		return "", errors.New("required 'GITHUB_ACCESS_TOKEN' is not set")
	}
	return accessToken, nil
}

// SetFakeProvider sets the provider to fake. This is only for testing.
// The t argument is added as a reminder that this is only for testing.
func SetFakeProvider(t *testing.T) {
	t.Helper()
	provider = "fake"
}

// HasAppID returns true if the environment specifies an APP_ID.
func HasAppID() bool {
	_, err := AppID()
	return err == nil
}

func GetAppURL() string {
	return os.Getenv("QUICKFEED_APP_URL")
}
