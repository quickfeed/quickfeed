package scm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

func newGithubAppClient(ctx context.Context, logger *zap.SugaredLogger, cfg *Config, organization string) (*GithubSCM, error) {
	inst, err := cfg.fetchInstallation(organization)
	if err != nil {
		return nil, err
	}
	installCfg, err := cfg.InstallationConfig(strconv.Itoa(int(inst.GetID())))
	if err != nil {
		return nil, fmt.Errorf("error configuring github client for installation: %w", err)
	}
	httpClient := installCfg.Client(ctx)
	return &GithubSCM{
		logger:             logger,
		client:             github.NewClient(httpClient),
		clientV4:           githubv4.NewClient(httpClient),
		tokenManager:       newAppTokenManager(logger, cfg, inst.GetID()),
		providerURL:        "https://github.com",
		createUserClientFn: newGithubUserClient,
	}, nil
}

type appTokenManager struct {
	logger   *zap.SugaredLogger
	config   *Config
	tokenURL string

	mu        sync.Mutex // protects token and expiresAt
	token     string
	expiresAt time.Time
}

func newAppTokenManager(logger *zap.SugaredLogger, cfg *Config, installationID int64) *appTokenManager {
	return &appTokenManager{
		logger:   logger,
		config:   cfg,
		tokenURL: fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID),
	}
}

// Token returns a valid installation access token, refreshing it if necessary.
func (m *appTokenManager) Token(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return the current token if it is still valid.
	if m.token != "" && time.Now().Before(m.expiresAt) {
		return m.token, nil
	}

	body, err := m.requestNewToken(ctx)
	if err != nil {
		return "", err
	}
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
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	m.logger.Infof("Fetched new GitHub App access token; expires at: %s", tokenResponse.ExpiresAt)

	m.token = tokenResponse.Token
	m.expiresAt = tokenResponse.ExpiresAt.Add(-time.Minute) // Refresh one minute before expiry
	return m.token, nil
}

func (m *appTokenManager) requestNewToken(ctx context.Context) ([]byte, error) {
	if m.config == nil {
		return nil, errors.New("cannot refresh token without config")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.tokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.github.v3+json")
	resp, err := m.config.Client().Do(req)
	if err != nil {
		// Note: If the installation was deleted on GitHub, the installation ID will be invalid.
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch installation access token (response status %s): %s", resp.Status, string(body))
	}
	return body, nil
}

func (cfg *Config) fetchInstallation(organization string) (*github.Installation, error) {
	const installationURL = "https://api.github.com/app/installations"
	resp, err := cfg.Client().Get(installationURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching GitHub app installation for organization %s: %w", organization, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading installation response: %w", err)
	}
	var installations []*github.Installation
	if err := json.Unmarshal(body, &installations); err != nil {
		return nil, fmt.Errorf("error unmarshalling installation response: %s: %w", body, err)
	}
	for _, inst := range installations {
		if inst.GetAccount().GetLogin() == organization {
			return inst, nil
		}
	}
	return nil, fmt.Errorf("could not find GitHub app installation for organization %s", organization)
}

type ExchangeToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ExchangeToken exchanges a refresh token for an access token.
func (cfg *Config) ExchangeToken(refreshToken string) (*ExchangeToken, error) {
	if cfg == nil {
		return nil, errors.New("cannot exchange refresh token without config")
	}
	form := map[string][]string{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}
	resp, err := cfg.Client().PostForm("https://github.com/login/oauth/access_token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get access token: %s", resp.Status)
	}
	var token ExchangeToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	if token.AccessToken == "" || token.RefreshToken == "" {
		return nil, fmt.Errorf("tokens are empty")
	}
	return &token, nil
}
