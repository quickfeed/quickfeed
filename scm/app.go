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
	gh "github.com/google/go-github/v43/github"
)

const (
	AppEnv           = "APP_ID"
	KeyEnv           = "APP_KEY"
	SecretEnv        = "APP_SECRET"
	KeyPath          = "./appth.private-key.pem" // TODO(vera): read path from env
	InstallationsAPI = "https://api.github.com/app/installations"
)

// GithubAppConfig keeps parameters of the GitHub app
type GithubAppConfig struct {
	appID    string
	clientID string
	secret   string
	keyPath  string
}

type GithubApp struct {
	app    *app.Config
	config *GithubAppConfig
}

func newAppConfig() *GithubAppConfig {
	return &GithubAppConfig{
		appID:    os.Getenv(AppEnv),
		clientID: os.Getenv(KeyEnv),
		secret:   os.Getenv(SecretEnv),
	}
}

func (conf *GithubAppConfig) Valid() bool {
	return conf.appID != "" &&
		conf.clientID != "" && conf.secret != ""
}

// AppClient creates client for the Quickfeed GitHub Application
// This client can only access the metadata of the Application itself
// To access organizations via GitHub API we need to derive an installation client
// from this Application client for each course organization
func NewApp() (*GithubApp, error) {
	config := newAppConfig()
	if !config.Valid() {
		return nil, fmt.Errorf("error configuring GitHub App: %+v", config)
	}
	appKey, err := key.FromFile(KeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading key from file: %s", err)
	}
	appClientConfig, err := app.NewConfig(config.appID, appKey)
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub application client: %s", err)
	}
	return &GithubApp{
		config: config,
		app:    appClientConfig,
	}, nil
}

// Creates a new scm client with access to the course organization
func (ghApp *GithubApp) NewInstallationClient(ctx context.Context, courseOrg string) (*gh.Client, error) {
	resp, err := ghApp.app.Client().Get(InstallationsAPI)
	if err != nil {
		return nil, fmt.Errorf("error fetching installations for GitHub app %s: %s", ghApp.config.appID, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading installation response: %s", err)
	}
	var installations []*gh.Installation
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
	install, err := ghApp.app.InstallationConfig(strconv.Itoa(int(installationID)))
	if err != nil {
		return nil, fmt.Errorf("error configuring github client for installation: %s", err)
	}
	return gh.NewClient(install.Client(ctx)), nil
}
