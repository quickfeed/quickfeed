package hooks

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/stream"
)

// maxConcurrentTestRuns is the maximum number of concurrent test runs.
const maxConcurrentTestRuns = 5

// webhookTimeout bounds the handling of a single push event, which clones the
// repository and runs the assignment's tests in a container.
const webhookTimeout = 30 * time.Minute

const (
	// eventTypeLabel is the attribute holding the GitHub webhook event type.
	eventTypeLabel = "event_type"
	// branchRefLabel is the attribute holding the pushed git reference,
	// e.g., refs/heads/main; see label.Branch for the short branch name.
	branchRefLabel = "branch_ref"
)

// GitHubWebHook holds references and data for handling webhook events.
type GitHubWebHook struct {
	logger  *slog.Logger
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
func NewGitHubWebHook(logger *slog.Logger, db database.Database, mgr *scm.Manager, runner ci.Runner, secret string, streams *stream.StreamServices, tm *auth.TokenManager) *GitHubWebHook {
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
			wh.logger.Error("invalid webhook request body", label.Error, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		defer r.Body.Close()

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			wh.logger.Error("failed to parse GitHub webhook", label.Error, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Scope every record for this event to the event type; the handlers below
		// derive their logger from the context and add their own scope.
		// The context is not used for cancellation to avoid that a slow handler
		// cancels the event handler.
		ctx := context.WithoutCancel(r.Context())
		ctx, logger := qlog.WithLogger(qlog.NewContext(ctx, wh.logger), eventTypeLabel, github.WebHookType(r))
		logger.Debug("received webhook event")

		switch e := event.(type) {
		case *github.PushEvent:
			commitID := e.GetHeadCommit().GetID()
			ctx, logger := qlog.WithLogger(ctx, label.Commit, commitID)
			logger.Debug("received push event")
			if wh.dup.Duplicate(commitID) {
				logger.Debug("ignoring duplicate push event")
				return
			}

			// The counting semaphore limits concurrency to maxConcurrentTestRuns.
			// This should also allow webhook events to return quickly to GitHub, avoiding timeouts.
			// Note however, if we receive a large number of push events, we may be creating
			// a large number of goroutines. If this becomes a problem, we can add rate limiting
			// on the number of goroutines created, by returning a http.StatusTooManyRequests.
			go func() {
				wh.sem <- struct{}{} // acquire semaphore
				defer func() {
					<-wh.sem // release semaphore
					// remove commitID from duplicate map (to avoid memory leak)
					wh.dup.Remove(commitID)
				}()
				// Start the timeout after acquiring the semaphore, so that time spent
				// queueing behind other test runs does not count against the handler.
				ctx, cancel := context.WithTimeout(ctx, webhookTimeout)
				defer cancel()
				wh.handlePush(ctx, e)
			}()

		case *github.PullRequestEvent:
			switch e.GetAction() {
			case "opened":
				wh.handlePullRequestOpened(ctx, e)
			case "closed":
				wh.handlePullRequestClosed(ctx, e)
			}

		case *github.PullRequestReviewEvent:
			wh.handlePullRequestReview(ctx, e)

		case *github.InstallationEvent:
			switch e.GetAction() {
			case "created":
				wh.handleInstallationCreated(ctx, e)
			case "deleted":
				wh.handleInstallationDeleted(ctx, e)
			default:
				// either "suspend", "unsuspend", "new_permissions_accepted"
				wh.logger.Debug("Ignored installation event action", "action", e.GetAction())
			}

		default:
			logger.Debug("ignored webhook event")
		}
	}
}
