package hooks

import (
	"net/http"

	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/log"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
)

// GitHubWebHook holds references and data for handling webhook events.
type GitHubWebHook struct {
	logger *zap.SugaredLogger
	db     database.Database
	runner ci.Runner
	secret string
}

// NewGitHubWebHook creates a new webhook to handle POST requests from GitHub to the QuickFeed server.
func NewGitHubWebHook(logger *zap.SugaredLogger, db database.Database, runner ci.Runner, secret string) *GitHubWebHook {
	return &GitHubWebHook{logger: logger, db: db, runner: runner, secret: secret}
}

// Handle take POST requests from GitHub, representing Push events
// associated with course repositories, which then triggers various
// actions on the QuickFeed backend.
func (wh GitHubWebHook) Handle(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(wh.secret))
	if err != nil {
		wh.logger.Errorf("Error in request body: %v", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		wh.logger.Errorf("Could not parse github webhook: %v", err)
		return
	}
	switch e := event.(type) {
	case *github.PushEvent:
		wh.logger.Debug(log.IndentJson(e))
		wh.handlePush(e)
	case *github.PullRequestEvent:
		switch e.GetAction() {
		case "opened":
			wh.logger.Debug(log.IndentJson(e))
			wh.handlePullRequestOpened(e)
		case "closed":
			wh.logger.Debug(log.IndentJson(e))
			wh.handlePullRequestClosed(e)
		}
	case *github.PullRequestReviewEvent:
		wh.logger.Debug(log.IndentJson(e))
		wh.handlePullRequestReview(e)
	default:
		wh.logger.Debugf("Ignored event type %s", github.WebHookType(r))
	}
}
