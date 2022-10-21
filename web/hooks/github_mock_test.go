package hooks_test

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/google/go-github/v45/github"
	"go.uber.org/zap"
)

// maxConcurrentTestRuns is the maximum number of concurrent test runs.
const maxConcurrentTestRuns = 5

// GitHubWebHook holds references and data for handling webhook events.
type MockWebHook struct {
	logger                *zap.SugaredLogger
	secret                string
	sem                   chan struct{}
	totalCnt              int32
	currentConcurrencyCnt int32
	highestConcurrencyCnt int32
	lowestConcurrencyCnt  int32
	wg                    *sync.WaitGroup
}

// NewMockWebHook creates a new webhook to handle POST requests to the QuickFeed server.
func NewMockWebHook(logger *zap.SugaredLogger, secret string) *MockWebHook {
	// counting semaphore: limit concurrent test runs to maxConcurrentTestRuns
	sem := make(chan struct{}, maxConcurrentTestRuns)
	wg := &sync.WaitGroup{}
	return &MockWebHook{logger: logger, secret: secret, sem: sem, wg: wg}
}

// Handle take POST requests from GitHub, representing Push events
// associated with course repositories, which then triggers various
// actions on the QuickFeed backend.
func (wh *MockWebHook) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, []byte(wh.secret))
		if err != nil {
			wh.logger.Errorf("Error in request body: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		defer r.Body.Close()

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			wh.logger.Errorf("Could not parse github webhook: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch e := event.(type) {
		case *github.PushEvent:
			// The counting semaphore limits concurrency to maxConcurrentTestRuns.
			// This should also allow webhook events to return quickly to GitHub, avoiding timeouts.
			wh.wg.Add(1)
			go func() {
				wh.sem <- struct{}{} // acquire semaphore
				current := atomic.AddInt32(&wh.currentConcurrencyCnt, 1)
				if atomic.LoadInt32(&wh.highestConcurrencyCnt) < current {
					atomic.StoreInt32(&wh.highestConcurrencyCnt, current)
				}
				if atomic.LoadInt32(&wh.lowestConcurrencyCnt) > current {
					atomic.StoreInt32(&wh.lowestConcurrencyCnt, current)
				}
				wh.handlePush(e)
				<-wh.sem // release semaphore
				atomic.AddInt32(&wh.currentConcurrencyCnt, -1)
				wh.wg.Done()
			}()
		default:
			wh.logger.Debugf("Ignored event type %s", github.WebHookType(r))
		}
	}
}

func (wh *MockWebHook) handlePush(payload *github.PushEvent) {
	curCnt := atomic.AddInt32(&wh.totalCnt, 1)
	wh.logger.Debugf("Received push event on %s / %d", payload.GetRepo().GetName(), curCnt)
}
