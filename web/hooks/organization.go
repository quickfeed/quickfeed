package hooks

import (
	"context"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func (wh GitHubWebHook) handleOrgInvite(e *github.OrganizationEvent) {
	// check that invitee is enrolled in course
	invitee := e.GetInvitation().GetLogin()
	user, err := wh.db.GetUserByCourse(&qf.Course{ScmOrganizationID: uint64(e.GetOrganization().GetID())}, invitee)
	if err != nil {
		wh.logger.Error("Invite: Failed to find user:", err)
		return // course does not exist, or user not enrolled
	}

	invite := e.GetInvitation()
	if invite == nil {
		wh.logger.Error("No invite")
		return
	}

	orgName := e.GetOrganization().GetLogin()
	sc, ok := wh.scmMgr.GetSCM(orgName)
	if !ok {
		wh.logger.Debug("Organization", orgName, "does not have a SCM")
		return
	}

	// AcceptInvitations consumes the current refresh token and returns a new one
	newRefreshToken, err := sc.AcceptInvitations(context.Background(), &scm.InvitationOptions{
		Login:        user.GetLogin(),
		Owner:        orgName,
		RefreshToken: user.GetRefreshToken(),
	})
	if err != nil {
		wh.logger.Error(err)
	}
	// The new refresh token needs to be stored in the DB
	user.RefreshToken = newRefreshToken
	if err := wh.db.UpdateUser(user); err != nil {
		wh.logger.Errorf("failed to update refresh token for user (id=%d, name=%s): %w", user.ID, user.Name, err)
	}
}
