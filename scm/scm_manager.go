package scm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
	"github.com/google/go-github/v45/github"
	"go.uber.org/zap"
)

const installationAPI = "https://api.github.com/app/installations"

// GithubConfig keeps parameters of the GitHub app.
type GithubConfig struct {
	appID     string
	clientID  string
	secret    string
	keyPath   string
	appConfig *app.Config
}

// SCMManager keeps provider-specific configs (currently only for GitHub)
// and a map of scm clients for each course.
type SCMManager struct {
	Scms         *Scms
	githubConfig *GithubConfig
}

// newGitHubConfig creates a new configuration for GitHub app
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

// NewSCMManager creates base client for the QuickFeed GitHub Application. This client
// cannto access organization-related GitHub API and is used to derive installation
// API clients for each course organization.
func NewSCMManager(appID, appKey, appSecret, appKeyFile string) (*SCMManager, error) {
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
	return &SCMManager{
		githubConfig: config,
		Scms:         NewScms(),
	}, nil
}

func (sm *SCMManager) NewInstallationClient(ctx context.Context, organization string) (*github.Client, error) {
	resp, err := sm.githubConfig.appConfig.Client().Get(installationAPI)
	if err != nil {
		return nil, fmt.Errorf("error fetching installations for GitHub app %s: %w", sm.githubConfig.appID, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading installation response: %w", err)
	}
	var installations []*github.Installation
	if err := json.Unmarshal(body, &installations); err != nil {
		return nil, fmt.Errorf("error unmarshalling installation response: %w", err)
	}
	var installationID int64
	for _, inst := range installations {
		if *inst.Account.Login == organization {
			installationID = *inst.ID
			break
		}
	}
	if installationID == 0 {
		return nil, fmt.Errorf("cannot find GitHub app installation for organization %s", organization)
	}
	install, err := sm.githubConfig.appConfig.InstallationConfig(strconv.Itoa(int(installationID)))
	if err != nil {
		return nil, fmt.Errorf("error configuring github client for installation: %w", err)
	}
	return github.NewClient(install.Client(ctx)), nil
}

func (sm *SCMManager) NewSCMClient(ctx context.Context, logger *zap.SugaredLogger, organization string) (*GithubSCM, error) {
	client, err := sm.NewInstallationClient(ctx, organization)
	if err != nil {
		return nil, err
	}
	return &GithubSCM{
		logger: logger,
		client: client,
	}, nil
}
