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
	scms map[string]SCM
	mu   sync.RWMutex
	*Config
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
func NewSCMManager(c *Config) *Manager {
	return &Manager{
		scms:   make(map[string]SCM),
		Config: c,
	}
}

// GetOrCreateSCM returns an SCM client for the given organization, or creates a new SCM client if non exists.
func (s *Manager) GetOrCreateSCM(ctx context.Context, logger *zap.SugaredLogger, organization string) (SCM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	client, ok := s.scms[organization]
	if !ok {
		var err error
		client, err = newSCMAppClient(ctx, logger, s.Config, organization)
		if err != nil {
			return nil, err
		}
	}
	s.scms[organization] = client
	return client, nil
}

// GetSCM returns an SCM client for the given organization if exists;
// otherwise, nil and false is returned.
func (s *Manager) GetSCM(organization string) (sc SCM, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok = s.scms[organization]
	return
}
