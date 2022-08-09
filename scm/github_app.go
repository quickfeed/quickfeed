package scm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

func newGithubAppClient(ctx context.Context, logger *zap.SugaredLogger, cfg *Config, organization string) (*GithubSCM, error) {
	inst, err := cfg.fetchInstallation(ctx, organization)
	if err != nil {
		return nil, err
	}
	installCfg, err := cfg.InstallationConfig(strconv.Itoa(int(inst.GetID())))
	if err != nil {
		return nil, fmt.Errorf("error configuring github client for installation: %w", err)
	}
	httpClient := installCfg.Client(ctx)
	return &GithubSCM{
		logger:      logger,
		client:      github.NewClient(httpClient),
		clientV4:    githubv4.NewClient(httpClient),
		providerURL: "github.com",
		tokenURL:    fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", inst.GetID()),
	}, nil
}

func (cfg *Config) fetchInstallation(ctx context.Context, organization string) (*github.Installation, error) {
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
		return nil, fmt.Errorf("error unmarshalling installation response: %w", err)
	}
	for _, inst := range installations {
		if *inst.Account.Login == organization {
			return inst, nil
		}
	}
	return nil, fmt.Errorf("could not find GitHub app installation for organization %s", organization)
}

func (s *GithubSCM) refreshToken(ctx context.Context, cfg *Config, organization string) error {
	resp, err := cfg.Client().Post(s.tokenURL, "application/vnd.github.v3+json", nil)
	if err != nil {
		// Note: If the installation was deleted on GitHub, the installation ID will be invalid.
		return err
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
		return fmt.Errorf("failed to fetch installation access token for %s (response status %s): %s", organization, resp.Status, string(body))
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return err
	}
	s.token = tokenResponse.Token
	return nil
}
