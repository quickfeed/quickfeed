package hooks

import (
	"net/http"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/stream"
	"go.uber.org/zap"
)

// maxConcurrentTestRuns is the maximum number of concurrent test runs.
const maxConcurrentTestRuns = 5

// GitHubWebHook holds references and data for handling webhook events.
type GitHubWebHook struct {
	logger  *zap.SugaredLogger
	db      database.Database
	scmMgr  *scm.Manager
	runner  ci.Runner
	secret  string
	streams *stream.StreamServices
	sem     chan struct{} // counting semaphore: limit concurrent test runs to maxConcurrentTestRuns
	dup     *Duplicates
	tm      *auth.TokenManager
}

// NewGitHubWebHook creates a new webhook to handle POST requests from GitHub to the QuickFeed server.
func NewGitHubWebHook(logger *zap.SugaredLogger, db database.Database, mgr *scm.Manager, runner ci.Runner, secret string, streams *stream.StreamServices, tm *auth.TokenManager) *GitHubWebHook {
	return &GitHubWebHook{
		logger:  logger,
		db:      db,
		scmMgr:  mgr,
		runner:  runner,
		secret:  secret,
		streams: streams,
		sem:     make(chan struct{}, maxConcurrentTestRuns),
		dup:     NewDuplicateMap(),
		tm:      tm,
	}
}

// Handle take POST requests from GitHub, representing Push events
// associated with course repositories, which then triggers various
// actions on the QuickFeed backend.
func (wh GitHubWebHook) Handle() http.HandlerFunc {
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
		wh.logger.Debug(qlog.IndentJson(event))
		switch e := event.(type) {
		case *github.PushEvent:
			commitID := e.GetHeadCommit().GetID()
			wh.logger.Debugf("Received push event: %s", commitID)
			if wh.dup.Duplicate(commitID) {
				wh.logger.Debugf("Ignoring duplicate push event: %s", commitID)
				return
			}

			// The counting semaphore limits concurrency to maxConcurrentTestRuns.
			// This should also allow webhook events to return quickly to GitHub, avoiding timeouts.
			// Note however, if we receive a large number of push events, we may be creating
			// a large number of goroutines. If this becomes a problem, we can add rate limiting
			// on the number of goroutines created, by returning a http.StatusTooManyRequests.
			go func() {
				wh.sem <- struct{}{} // acquire semaphore
				wh.handlePush(e)
				<-wh.sem // release semaphore
				// remove commitID from duplicate map (to avoid memory leak)
				wh.dup.Remove(commitID)
			}()

		case *github.PullRequestEvent:
			switch e.GetAction() {
			case "opened":
				wh.handlePullRequestOpened(e)
			case "closed":
				wh.handlePullRequestClosed(e)
			}

		case *github.PullRequestReviewEvent:
			wh.handlePullRequestReview(e)

		case *github.InstallationEvent:
			switch e.GetAction() {
			case "created":
				wh.handleInstallationCreated(e)
			default:
				// either "deleted", "suspend", "unsuspend", "new_permissions_accepted"
			}

		default:
			wh.logger.Debugf("Ignored event type %s", github.WebHookType(r))
		}
	}
}
