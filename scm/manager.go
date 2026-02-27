package scm

import (
	"context"
	"fmt"
	"sync"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"go.uber.org/zap"
)

// Manager keeps provider-specific configs (currently only for GitHub)
// and a map of scm clients for each course.
type Manager struct {
	// map: [course organization] -> scm client
	scms map[string]SCM
	mu   sync.Mutex
	*Config
	exchangeTokenFn func(refreshToken string) (*ExchangeToken, error)
}

// ExchangeAndUpdateToken returns a new access token for the given user.
// It also updates the user's refresh token in the user object.
func (m *Manager) ExchangeAndUpdateToken(user *qf.User) (string, error) {
	exchangeToken, err := m.exchangeTokenFn(user.GetRefreshToken())
	if err != nil {
		return "", fmt.Errorf("failed to exchange token for user %s: %w", user.GetLogin(), err)
	}
	user.UpdateRefreshToken(exchangeToken.RefreshToken)
	return exchangeToken.AccessToken, nil
}

// Config stores SCM variables.
type Config struct {
	ClientID     string
	ClientSecret string
	*app.Config
}

// NewSCMConfig creates a new SCMConfig.
func NewSCMConfig() (*Config, error) {
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
	createAppKey, err := key.FromFile(env.AppPrivKeyFile())
	if err != nil {
		return nil, fmt.Errorf("error reading key from file: %w", err)
	}
	appConfig, err := app.NewConfig(appID, createAppKey)
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub application client: %w", err)
	}
	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Config:       appConfig,
	}, nil
}

// NewSCMManager creates base client for the QuickFeed GitHub Application.
// This client can be used to install API clients for each course organization.
func NewSCMManager() (*Manager, error) {
	c, err := NewSCMConfig()
	if err != nil {
		return nil, err
	}
	return &Manager{
		scms:            make(map[string]SCM),
		Config:          c,
		exchangeTokenFn: c.ExchangeToken,
	}, nil
}

// GetOrCreateSCM returns an SCM client for the given organization, or creates a new SCM client if non exists.
func (s *Manager) GetOrCreateSCM(ctx context.Context, logger *zap.SugaredLogger, organization string) (SCM, error) {
	s.mu.Lock()
	client, ok := s.scms[organization]
	s.mu.Unlock()
	if ok {
		return client, nil
	}
	client, err := newSCMAppClient(ctx, logger, s.Config, organization)
	if err != nil {
		return nil, fmt.Errorf("failed to create github application for %s: %w", organization, err)
	}
	s.mu.Lock()
	s.scms[organization] = client
	s.mu.Unlock()
	return client, nil
}

// GetSCM returns an SCM client for the given organization if exists;
// otherwise, nil and false is returned.
func (s *Manager) GetSCM(organization string) (sc SCM, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc, ok = s.scms[organization]
	return
}
