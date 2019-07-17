package auth

import (
	"sync"

	"github.com/autograde/aguis/scm"
	"go.uber.org/zap"
)

// Scms stores information about active scm clients.
type Scms struct {
	scms map[string]scm.SCM
	mu   sync.RWMutex
}

// NewScms returns reference to new thread-safe map
func NewScms() *Scms {
	return &Scms{scms: make(map[string]scm.SCM)}
}

// GetSCM returns an scm client for the given access token, if such token exists;
// otherwise, nil and false is returned.
func (s *Scms) GetSCM(accessToken string) (sc scm.SCM, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok = s.scms[accessToken]
	return
}

// GetOrCreateSCMEntry returns an scm client for the given remote identity
// (provider, access token) pair. If no scm client exists for the given
// remote identity, one will be created and stored for later retrival.
func (s *Scms) GetOrCreateSCMEntry(logger *zap.Logger, provider, accessToken string) (scm.SCM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	client, ok := s.scms[accessToken]
	if ok {
		return client, nil
	}
	client, err := scm.NewSCMClient(logger, provider, accessToken)
	if err != nil {
		return nil, err
	}
	s.scms[accessToken] = client
	return client, nil
}
