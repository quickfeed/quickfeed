package scm

import (
	"fmt"
	"os"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
)

// GithubConfig keeps parameters of the GitHub app.
type GithubConfig struct {
	appID     string
	clientID  string
	secret    string
	keyPath   string
	appConfig *app.Config
}

// SCMMaker keeps provider-specific configs
// and a map of scm clients for each course.
type SCMMaker struct {
	scms         *Scms
	githubConfig *GithubConfig
}

func newGitHubConfig(appID, appKey, appSecret, appKeyFile string) *GithubConfig {
	return &GithubConfig{
		appID:    appID,
		clientID: appKey,
		secret:   appSecret,
		keyPath:  appKeyFile,
	}
}

// valid ensures that all configuration fields are not empty
func (conf *GithubConfig) valid() bool {
	return conf.appID != "" && conf.keyPath != "" &&
		conf.clientID != "" && conf.secret != ""
}

// NewSCMMaker creates base client for the Quickfeed GitHub Application.
// This client can only access the metadata of the Application itself
// like ID, settings or a list of installations.
// To access organizations via GitHub API we need to derive an installation client
// from this application client for each course organization.
func NewSCMMaker(appID, appKey, appSecret, appKeyFile string) (*SCMMaker, error) {
	config := newGitHubConfig(appID, appKey, appSecret, appKeyFile)
	if !config.valid() {
		return nil, fmt.Errorf("error configuring GitHub App: %+v", config)
	}
	createAppKey, err := key.FromFile(appKeyFile)
	if err != nil {
		wd, _ := os.Getwd()
		return nil, fmt.Errorf("wd %s, error reading key from file: %s", wd, err)
	}
	appClientConfig, err := app.NewConfig(config.appID, createAppKey)
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub application client: %s", err)
	}
	config.appConfig = appClientConfig
	return &SCMMaker{
		githubConfig: config,
		scms:         NewScms(),
	}, nil
}
