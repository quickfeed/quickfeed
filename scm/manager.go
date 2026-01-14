package scm

import (
	"context"
	"fmt"
	"sync"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
	"github.com/quickfeed/quickfeed/internal/env"
	"go.uber.org/zap"
)

// Manager keeps provider-specific configs (currently only for GitHub)
// and a map of scm clients for each course.
type Manager struct {
	// map: [course organization] -> scm client
	scms map[string]SCM
	mu   sync.Mutex
	*Config
	// exchangeTokenFn is an optional function for exchanging refresh tokens.
	// If nil, the default Config.ExchangeToken method is used.
	// This field exists to enable mocking in tests.
	exchangeTokenFn func(refreshToken string) (*ExchangeToken, error)
}

// ExchangeToken exchanges a refresh token for an access token.
// If a custom exchangeTokenFn is set (for testing), it uses that.
// Otherwise, it delegates to the Config's ExchangeToken method.
func (m *Manager) ExchangeToken(refreshToken string) (*ExchangeToken, error) {
	if m.exchangeTokenFn != nil {
		return m.exchangeTokenFn(refreshToken)
	}
	return m.Config.ExchangeToken(refreshToken)
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
	appKey := env.AppKey()
	createAppKey, err := key.FromFile(appKey)
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
		scms:   make(map[string]SCM),
		Config: c,
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
