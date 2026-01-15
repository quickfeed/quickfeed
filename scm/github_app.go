package scm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		tokenManager:       newAppTokenManager(cfg, inst.GetID()),
		providerURL:        "https://github.com",
		createUserClientFn: newGithubUserClient,
	}, nil
}

type appTokenManager struct {
	config   *Config
	tokenURL string

	mu        sync.Mutex // protects token and expiresAt
	token     string
	expiresAt time.Time
}

func newAppTokenManager(cfg *Config, installationID int64) *appTokenManager {
	return &appTokenManager{
		config:   cfg,
		tokenURL: fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID),
	}
}

// Token returns a valid installation access token, refreshing it if necessary.
func (m *appTokenManager) Token(ctx context.Context) (string, error) {
	// Return valid token if not expired
	if token := m.validToken(); token != "" {
		return token, nil
	}

	if m.config == nil {
		return "", errors.New("cannot refresh token without config")
	}

	resp, err := m.config.Client().Post(m.tokenURL, "application/vnd.github.v3+json", nil)
	if err != nil {
		// Note: If the installation was deleted on GitHub, the installation ID will be invalid.
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return "", fmt.Errorf("failed to fetch installation access token (response status %s): %s", resp.Status, string(body))
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

	return m.updateToken(tokenResponse.Token, tokenResponse.ExpiresAt), nil
}

// validToken returns the current token if it is still valid, or an empty string otherwise.
func (m *appTokenManager) validToken() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token != "" && time.Now().Before(m.expiresAt) {
		return m.token
	}
	return ""
}

// updateToken updates the token and its expiration time, returning the new token.
func (m *appTokenManager) updateToken(token string, expiresAt time.Time) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = token
	m.expiresAt = expiresAt
	return m.token
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
