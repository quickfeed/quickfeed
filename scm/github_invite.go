package scm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// GitHub creates organization invitations asynchronously. The first attempt to
// accept one often returns github.AcceptedError ("job scheduled on GitHub side").
// These defaults match waitForRepository: a few seconds is typical; the cap
// covers slower GitHub jobs without hanging the RPC indefinitely.
// Tests may shorten these.
var (
	waitForInviteMaxAttempts  = 10
	waitForInviteInitialDelay = 1 * time.Second
	waitForInviteMaxDelay     = 5 * time.Second
)

// acceptOrgInvitation accepts an organization membership invitation on behalf of the user.
// GitHub may return github.AcceptedError until the invitation job has finished;
// this retries with exponential backoff until the invitation is ready.
func (s *GithubSCM) acceptOrgInvitation(ctx context.Context, opt *InvitationOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid options: %+v", opt)
	}
	userSCM := s.createUserClientFn(opt.AccessToken)
	state := "active"
	logger := qlog.FromContext(ctx).With(label.TargetUser, opt.Login, label.Organization, opt.Owner)
	delay := waitForInviteInitialDelay
	var err error
	for attempt := range waitForInviteMaxAttempts {
		_, _, err = userSCM.Organizations.EditOrgMembership(ctx, "", opt.Owner, &github.Membership{State: &state})
		if err == nil {
			if attempt > 0 {
				logger.Debug("organization invitation accepted", "attempts", attempt+1)
			}
			return nil
		}
		if !isGitHubAccepted(err) {
			return fmt.Errorf("accepting invitation for %s to organization %s: %w", opt.Login, opt.Owner, err)
		}
		logger.Debug("organization invitation not ready", "attempt", attempt+1, "max_attempts", waitForInviteMaxAttempts, "delay", delay)
		select {
		case <-ctx.Done():
			return fmt.Errorf("accepting invitation for %s to organization %s: %w", opt.Login, opt.Owner, ctx.Err())
		case <-time.After(delay):
		}
		delay = min(delay*2, waitForInviteMaxDelay)
	}
	return fmt.Errorf("accepting invitation for %s to organization %s: invitation not ready after %d attempts: %w", opt.Login, opt.Owner, waitForInviteMaxAttempts, err)
}

func isGitHubAccepted(err error) bool {
	var accepted *github.AcceptedError
	return errors.As(err, &accepted)
}
