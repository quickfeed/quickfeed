package scm

import (
	"context"
	"fmt"
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
func (app *GithubApp) GetSCM(courseID uint64) (sc SCM, ok bool) {
	app.scms.mu.RLock()
	defer app.scms.mu.RUnlock()
	sc, ok = app.scms.scms[courseID]
	return
}

// AddSCM adds a new scm client to scm storage.
func (app *GithubApp) AddSCM(scm SCM, courseID uint64) {
	app.scms.mu.Lock()
	defer app.scms.mu.Unlock()
	app.scms.scms[courseID] = scm
}

// GetOrCreateSCMEntry returns an existing scm client for the given course or creates a new one.
func (app *GithubApp) GetOrCreateSCMEntry(logger *zap.SugaredLogger, course *pb.Course, accessToken string) (SCM, error) {
	app.scms.mu.Lock()
	defer app.scms.mu.Unlock()
	client, ok := app.scms.scms[course.GetID()]
	if ok {
		return client, nil
	}
	if accessToken == "" {
		return nil, fmt.Errorf("failed to create an scm client for course %s: missing access token", course.Code)
	}
	ghClient, err := app.NewInstallationClient(context.Background(), course.GetOrganizationPath())
	if err != nil {
		return nil, err
	}
	client, err = NewSCMClient(logger, ghClient, course.GetProvider(), accessToken)
	if err != nil {
		return nil, err
	}
	app.scms.scms[course.GetID()] = client
	return client, nil
}
