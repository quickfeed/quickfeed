package scm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"go.uber.org/zap"
)

const installationAPI = "https://api.github.com/app/installations"

// SCMManager keeps provider-specific configs (currently only for GitHub)
// and a map of scm clients for each course.
type SCMManager struct {
	Scms      *Scms
	appConfig *app.Config
}

// SCMConfig stores SCM varibles.
type SCMConfig struct {
	AppID        string
	AppKey       string
	ClientID     string
	ClientSecret string
}

// NewSCMConfig creates a new SCMConfig.
func NewSCMConfig() (*SCMConfig, error) {
	appID, err := env.AppID()
	if err != nil {
		return nil, err
	}
	clientID, err := env.ClientID()
	if err != nil {
		return nil, err
	}
	clientSecret, err := env.ClientSecret()
	if err != nil {
		return nil, err
	}
	appKey := env.AppKey()
	return &SCMConfig{
		AppID:        appID,
		AppKey:       appKey,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

// NewSCMManager creates base client for the QuickFeed GitHub Application.
// This client can be used to install API clients for each course organization.
func NewSCMManager(c *SCMConfig) (*SCMManager, error) {
	createAppKey, err := key.FromFile(c.AppKey)
	if err != nil {
		return nil, fmt.Errorf("error reading key from file: %w", err)
	}
	appConfig, err := app.NewConfig(c.AppID, createAppKey)
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub application client: %w", err)
	}
	return &SCMManager{
		appConfig: appConfig,
		Scms:      NewScms(),
	}, nil
}

// newInstallationClient creates a new client for a course organization.
func (sm *SCMManager) newInstallationClient(ctx context.Context, organization string) (*github.Client, error) {
	resp, err := sm.appConfig.Client().Get(installationAPI)
	if err != nil {
		return nil, fmt.Errorf("error fetching app installation for course organization %s: %w", organization, err)
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
	install, err := sm.appConfig.InstallationConfig(strconv.Itoa(int(installationID)))
	if err != nil {
		return nil, fmt.Errorf("error configuring github client for installation: %w", err)
	}
	return github.NewClient(install.Client(ctx)), nil
}

// SCMClient gets an existing SCM client by organization name or creates a new client for course organization.
func (s *SCMManager) GetOrCreateSCM(ctx context.Context, logger *zap.SugaredLogger, organization string) (SCM, error) {
	client, ok := s.Scms.scms[organization]
	if !ok {
		cli, err := s.newInstallationClient(ctx, organization)
		if err != nil {
			return nil, err
		}
		client = &GithubSCM{
			logger: logger,
			client: cli,
		}
	}
	s.Scms.scms[organization] = client
	return client, nil
}
