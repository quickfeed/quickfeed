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
	AppEnv           = "APP_ID"
	KeyEnv           = "APP_KEY"
	SecretEnv        = "APP_SECRET"
	KeyPath          = "APP_KEYPATH"
	InstallationsAPI = "https://api.github.com/app/installations"
	GitHubUserAPI    = "https://api.github.com/user"
)

// GithubConfig keeps parameters of the GitHub app
type GithubConfig struct {
	appID     string
	clientID  string
	secret    string
	keyPath   string
	appConfig *app.Config
}

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

// Valid ensures that all configuration fields are not empty
func (conf *GithubConfig) Valid() bool {
	return conf.appID != "" && conf.keyPath != "" &&
		conf.clientID != "" && conf.secret != ""
}

// AppClient creates client for the Quickfeed GitHub Application
// This client can only access the metadata of the Application itself
// To access organizations via GitHub API we need to derive an installation client
// from this Application client for each course organization
func NewApp() (*SCMMaker, error) {
	config := newAppConfig()
	if !config.Valid() {
		return nil, fmt.Errorf("error configuring GitHub App: %+v", config)
	}
	appKey, err := key.FromFile(config.keyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading key from file: %s", err)
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
	resp, err := sm.githubConfig.appConfig.Client().Get(InstallationsAPI)
	if err != nil {
		return nil, fmt.Errorf("error fetching installations for GitHub app %s: %s", sm.githubConfig.appID, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading installation response: %s", err)
	}
	var installations []*github.Installation
	if err := json.Unmarshal(body, &installations); err != nil {
		return nil, fmt.Errorf("error unmarshalling installation response: %s", err)
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
		return nil, fmt.Errorf("error configuring github client for installation: %s", err)
	}
	return github.NewClient(install.Client(ctx)), nil
}

// GetIDs returns app client ID and secret to be used in auth flow
func (sm *SCMMaker) GetID() (string, string) {
	return sm.githubConfig.clientID, sm.githubConfig.secret
}

func (sm *SCMMaker) GetUserURL() string {
	return GitHubUserAPI
}

// TODO(vera): update and move to a file with test helpers
func NewTestApp() *SCMMaker {
	return &SCMMaker{
		scms: NewScms(),
	}
}
