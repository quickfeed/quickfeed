package scm

import (
	"sync"

	"go.uber.org/zap"
)

// Scms stores information about active scm clients.
type Scms struct {
	scms map[string]SCM
	mu   sync.RWMutex
}

// NewScms returns reference to new thread-safe map
func NewScms() *Scms {
	return &Scms{scms: make(map[string]SCM)}
}

// GetSCM returns an scm client for the given owner (access token owner or course organization),
// if such token exists; otherwise, nil and false is returned.
func (s *Scms) GetSCM(owner string) (sc SCM, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok = s.scms[owner]
	return
}

// GetOrCreateSCMEntry returns an scm client for the given remote identity
// (access token or course organization). If no scm client exists for the given remote identity,
// one will be created and stored for later retrieval.
func (s *Scms) GetOrCreateSCMEntry(logger *zap.Logger, owner string) (SCM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	client, ok := s.scms[owner]
	if ok {
		return client, nil
	}
	client, err := NewSCMClient(logger.Sugar(), owner)
	if err != nil {
		return nil, err
	}
	s.scms[owner] = client
	return client, nil
}
