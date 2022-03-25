package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
	gh "github.com/google/go-github/v35/github"
)

const (
	keyPath = "./appth.private-key.pem"
)

// GithubAppConfig keeps parameters of the GitHub app
type GithubAppConfig struct {
	AppID    string
	ClientID string
	Secret   string
}

// TODO: rename because confusing
type GithubApp struct {
	Config *GithubAppConfig
	App    *app.Config
}

func NewAppConfig() *GithubAppConfig {
	return &GithubAppConfig{
		AppID:    os.Getenv(AppEnv),
		ClientID: os.Getenv(KeyEnv),
		Secret:   os.Getenv(SecretEnv),
	}
}

func (conf *GithubAppConfig) Valid() bool {
	return conf.AppID != "" &&
		conf.ClientID != "" && conf.Secret != ""
}

// AppClient creates client for the Quickfeed GitHub Application
// This client can only access the metadata of the Application itself
// To access organizations via GitHub API we need to derive an installation client
// from this Application client for each course organization
func NewApp() (*GithubApp, error) {
	config := NewAppConfig()
	if !config.Valid() {
		return nil, fmt.Errorf("Error configuring GitHub App: %+v", config)
	}
	appKey, err := key.FromFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading key from file: %s", err)
	}
	appClientConfig, err := app.NewConfig(config.AppID, appKey)
	if err != nil {
		return nil, fmt.Errorf("Error creating GitHub application client: %s", err)
	}
	return &GithubApp{
		Config: config,
		App:    appClientConfig,
	}, nil
}

// Creates a new scm client with access to the course organization
func (ghApp *GithubApp) NewInstallationClient(ctx context.Context, courseOrg string) (*gh.Client, error) {
	resp, err := ghApp.App.Client().Get(InstallationsAPI)
	if err != nil {
		return nil, fmt.Errorf("Cannot get installations for App %s: %s", ghApp.Config.AppID, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte
	if err != nil {
		return nil, fmt.Errorf("Error reading installation response: %s", err)
	}
	var installations []*gh.Installation
	if err := json.Unmarshal(body, &installations); err != nil {
		return nil, fmt.Errorf("Error unmarshalling installation response: %s", err)
	}
	var installationID int64
	for _, inst := range installations {
		log.Println("Checking installation ", *inst.Account.Login)
		if *inst.Account.Login == courseOrg {
			log.Println("Found installation for ", courseOrg)
			log.Println("Installation ID is ", *inst.ID)
			installationID = *inst.ID
		}
	}
	if installationID == 0 {
		return nil, fmt.Errorf("Installation not found for organization %s", courseOrg)
	}
	install, err := ghApp.App.InstallationConfig(strconv.Itoa(int(installationID)))
	if err != nil {
		return nil, fmt.Errorf("Error configuring github client for installation: %s", err)
	}
	return gh.NewClient(install.Client(ctx)), err
}
