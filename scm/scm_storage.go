package scm

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/autograde/quickfeed/ag"
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

// GetOrCreateSCMEntry returns an scm client for the given course. If no scm client exists for the course,
// creates a new client for the course.
func (app *GithubApp) GetOrCreateSCMEntry(logger *zap.Logger, course *pb.Course, provider, accessToken string) (SCM, error) {
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
	client, err = NewSCMClient(logger.Sugar(), ghClient, provider, accessToken)
	if err != nil {
		return nil, err
	}
	app.scms.scms[course.GetID()] = client
	return client, nil
}
