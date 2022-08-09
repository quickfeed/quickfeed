package scm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"go.uber.org/zap"
)

// Manager keeps provider-specific configs (currently only for GitHub)
// and a map of scm clients for each course.
type Manager struct {
	Scms      *Scms
	appConfig *app.Config
}

// SCMConfig stores SCM variables.
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
func NewSCMManager(c *SCMConfig) (*Manager, error) {
	createAppKey, err := key.FromFile(c.AppKey)
	if err != nil {
		return nil, fmt.Errorf("error reading key from file: %w", err)
	}
	appConfig, err := app.NewConfig(c.AppID, createAppKey)
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub application client: %w", err)
	}
	return &Manager{
		appConfig: appConfig,
		Scms:      NewScms(),
	}, nil
}

func (s *Manager) SCMWithToken(ctx context.Context, logger *zap.SugaredLogger, organization string) (SCM, error) {
	scmClient, err := s.GetOrCreateSCM(ctx, logger, organization)
	if err != nil {
		return nil, err
	}
	token, err := s.getInstallationToken(ctx, organization)
	if err != nil {
		return nil, err
	}
	scmClient.SetToken(token)
	return scmClient, err
}

// SCMClient gets an existing SCM client by organization name or creates a new client for course organization.
func (s *Manager) GetOrCreateSCM(ctx context.Context, logger *zap.SugaredLogger, organization string) (SCM, error) {
	client, ok := s.Scms.scms[organization]
	if !ok {
		cli, err := s.newInstallationClient(ctx, organization)
		if err != nil {
			return nil, err
		}
		client = &GithubSCM{
			logger:      logger,
			client:      cli,
			providerURL: "github.com",
		}
	}
	s.Scms.scms[organization] = client
	return client, nil
}

func (s *Manager) getInstallationToken(ctx context.Context, organization string) (string, error) {
	inst, err := s.getInstallation(ctx, organization)
	if err != nil {
		return "", err
	}
	tokenURL := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", inst.GetID())
	resp, err := s.appConfig.Client().Post(tokenURL, "application/vnd.github.v3+json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var tokenResponse struct {
		Token       string    `json:"token"`
		ExpiresAt   time.Time `json:"expires_at"`
		Permissions struct {
			Contents     string `json:"contents"`
			Metadata     string `json:"metadata"`
			PullRequests string `json:"pull_requests"`
		} `json:"permissions"`
		RepositorySelection string `json:"repository_selection"`
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return "", fmt.Errorf("failed to fetch installation access token for %s (response status %s): %s", organization, resp.Status, string(body))
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}
	return tokenResponse.Token, nil
}

func (s *Manager) getInstallation(ctx context.Context, organization string) (*github.Installation, error) {
	installationURL := "https://api.github.com/app/installations"
	var installations []*github.Installation
	resp, err := s.appConfig.Client().Get(installationURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching app installation for course organization %s: %w", organization, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading installation response: %w", err)
	}
	if err := json.Unmarshal(body, &installations); err != nil {
		return nil, fmt.Errorf("error unmarshalling installation response: %w", err)
	}
	var installation *github.Installation
	for _, inst := range installations {
		if *inst.Account.Login == organization {
			installation = inst
			break
		}
	}
	if installation == nil {
		return nil, fmt.Errorf("cannot find GitHub app installation for organization %s", organization)
	}
	return installation, err
}

// newInstallationClient creates a new client for a course organization.
func (s *Manager) newInstallationClient(ctx context.Context, organization string) (*github.Client, error) {
	inst, err := s.getInstallation(ctx, organization)
	if err != nil {
		return nil, err
	}
	install, err := s.appConfig.InstallationConfig(fmt.Sprintf("%d", inst.GetID()))
	if err != nil {
		return nil, fmt.Errorf("error configuring github client for installation: %w", err)
	}
	return github.NewClient(install.Client(ctx)), nil
}
