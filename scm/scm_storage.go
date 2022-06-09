package scm

import (
	"context"
	"sync"

	pb "github.com/autograde/quickfeed/ag/types"
	"go.uber.org/zap"
)

// Scms stores information about active scm clients.
type Scms struct {
	scms map[uint64]SCM
	mu   sync.RWMutex
}

// NewScms returns reference to new thread-safe map
func NewScms() *Scms {
	return &Scms{scms: make(map[uint64]SCM)}
}

// GetSCM returns an scm client for the given access token, if such token exists;
// otherwise, nil and false is returned.
func (sm *SCMMaker) GetSCM(courseID uint64) (sc SCM, ok bool) {
	sm.scms.mu.RLock()
	defer sm.scms.mu.RUnlock()
	sc, ok = sm.scms.scms[courseID]
	return
}

// AddSCM adds a new scm client to scm storage.
func (sm *SCMMaker) AddSCM(scm SCM, courseID uint64) {
	sm.scms.mu.Lock()
	defer sm.scms.mu.Unlock()
	sm.scms.scms[courseID] = scm
}

// GetOrCreateSCMEntry returns an existing scm client for the given course or creates a new one.
func (sm *SCMMaker) GetOrCreateSCMEntry(logger *zap.SugaredLogger, course *pb.Course) (SCM, error) {
	sm.scms.mu.Lock()
	defer sm.scms.mu.Unlock()
	client, ok := sm.scms.scms[course.GetID()]
	if ok {
		return client, nil
	}
	logger.Debug("COURSE ORGANIZATION: ", course.OrganizationPath)
	sc, err := sm.NewSCM(context.Background(), logger, course.OrganizationPath, course.GetAccessToken())
	if err != nil {
		return nil, err
	}
	sm.AddSCM(sc, course.ID)
	return sc, nil
}
