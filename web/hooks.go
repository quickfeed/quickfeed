package web

import (
	"context"
	"net/http"
	"strings"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/sirupsen/logrus"

	webhooks "gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/go-playground/webhooks.v3/gitlab"
)

// GithubHook handles events from GitHub.
func GithubHook(logger logrus.FieldLogger, db database.Database, runner ci.Runner) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := github.Event(h.Get("X-GitHub-Event"))

		switch event {
		case github.PushEvent:
			p := payload.(github.PushPayload)
			logger.WithField("payload", p).Println("Push event")

			remoteIdentity, err := db.GetRemoteIdentity("github", uint64(p.Sender.ID))
			if err != nil {
				logger.WithError(err).Warn("Failed to get sender's remote identity")
				return
			}
			logger.WithField("identity", remoteIdentity).Warn("Found sender's remote identity")

			getURL := p.Repository.CloneURL
			getURL = strings.TrimPrefix(getURL, "https://")
			getURL = strings.TrimSuffix(getURL, ".git")
			logger.WithField("url", getURL).Warn("Repository's go get URL")

			out, err := runner.Run(context.Background(), &ci.Job{
				Image: "golang:1.8.3",
				Commands: []string{
					`echo "\n\n==START_CI==\n"`,
					`git config --global url."https://` + remoteIdentity.AccessToken + `:x-oauth-basic@github.com/".insteadOf "https://github.com/"`,
					`go get "` + getURL + `"`,
					`cd "$GOPATH/src/` + getURL + `"`,
					`go test -v`,
					`echo "\n==DONE_CI==\n"`,
				},
			})

			if err != nil {
				logger.WithError(err).Warn("Docker failed")
				return
			}

			logger.WithField("out", out).Warn("Docker success")
		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}

// GitlabHook handles events from Gitlab.
func GitlabHook(logger logrus.FieldLogger) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := gitlab.Event(h.Get("X-Gitlab-Event"))

		switch event {
		case gitlab.PushEvents:
			p := payload.(gitlab.PushEventPayload)
			logger.WithField("payload", p).Println("Push event")
		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}
