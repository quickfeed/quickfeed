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
	"github.com/google/go-github/v43/github"
	"go.uber.org/zap"
)

const (
	AppEnv          = "APP_ID"
	KeyEnv          = "APP_KEY"
	SecretEnv       = "APP_SECRET"
	KeyPath         = "APP_KEYPATH"
	installationAPI = "https://api.github.com/app/installations"
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

func newAppConfig() *GithubConfig {
	return &GithubConfig{
		appID:    os.Getenv(AppEnv),
		clientID: os.Getenv(KeyEnv),
		secret:   os.Getenv(SecretEnv),
		keyPath:  os.Getenv(KeyPath),
	}
}

// valid ensures that all configuration fields are not empty
func (conf *GithubConfig) valid() bool {
	return conf.appID != "" && conf.keyPath != "" &&
		conf.clientID != "" && conf.secret != ""
}

// NewSCMMaker creates client for the Quickfeed GitHub Application.
// This client can only access the metadata of the Application itself
// like ID, settings or a list of installations.
// To access organizations via GitHub API we need to derive an installation client
// from this Application client for each course organization
func NewSCMMaker() (*SCMMaker, error) {
	config := newAppConfig()
	if !config.valid() {
		return nil, fmt.Errorf("error configuring GitHub App: %+v", config)
	}
	appKey, err := key.FromFile(config.keyPath)
	if err != nil {
		wd, _ := os.Getwd()
		return nil, fmt.Errorf("wd %s error reading key from file: %s", wd, err)
	}
	appClientConfig, err := app.NewConfig(config.appID, appKey)
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub application client: %s", err)
	}
	config.appConfig = appClientConfig
	return &SCMMaker{
		githubConfig: config,
		scms:         NewScms(),
	}, nil
}

func (sm *SCMMaker) NewSCM(ctx context.Context, logger *zap.SugaredLogger, courseOrg, token string) (SCM, error) {
	client, err := sm.NewInstallationClient(ctx, courseOrg)
	if err != nil {
		return nil, err
	}
	return &GithubSCM{
		logger: logger,
		client: client,
		token:  token,
	}, nil
}

// Creates a new scm client with access to the course organization
func (sm *SCMMaker) NewInstallationClient(ctx context.Context, courseOrg string) (*github.Client, error) {
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
		if *inst.Account.Login == courseOrg {
			installationID = *inst.ID
			break
		}
	}
	if installationID == 0 {
		return nil, fmt.Errorf("cannot find GitHub app installation for organization %s", courseOrg)
	}
	install, err := sm.githubConfig.appConfig.InstallationConfig(strconv.Itoa(int(installationID)))
	if err != nil {
		return nil, fmt.Errorf("error configuring github client for installation: %w", err)
	}
	return github.NewClient(install.Client(ctx)), nil
}

// GetIDs returns app client ID and secret to be used in auth flow
func (sm *SCMMaker) GetIDs() (string, string) {
	return sm.githubConfig.clientID, sm.githubConfig.secret
}
