package hooks

import (
	"context"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func (wh GitHubWebHook) handleOrgInvite(e *github.OrganizationEvent) {
	login := e.GetInvitation().GetLogin()
	org := e.GetOrganization()
	course := &qf.Course{
		ScmOrganizationID:   uint64(org.GetID()),
		ScmOrganizationName: org.GetLogin(),
	}

	courseOrg := org.GetLogin()
	// check that invitee is enrolled in course
	user, err := wh.db.GetUserByCourse(course, login)
	if err != nil {
		wh.logger.Error("Invite: Failed to find user %s in course %s: %v", login, courseOrg, err)
		return // course does not exist, or user not enrolled
	}

	sc, ok := wh.scmMgr.GetSCM(courseOrg)
	if !ok {
		wh.logger.Debug("Organization", courseOrg, "does not have a SCM")
		return
	}

	// AcceptInvitations consumes the current refresh token and returns a new one
	// If the token exchange fails, the original refresh token is returned
	newRefreshToken, err := sc.AcceptInvitations(context.Background(), &scm.InvitationOptions{
		Login:        user.GetLogin(),
		Owner:        courseOrg,
		RefreshToken: user.GetRefreshToken(),
	})
	if err != nil {
		// Regardless of the error, we need to update the user's refresh token in the database
		wh.logger.Error(err)
	}

	user.RefreshToken = newRefreshToken
	if err := wh.db.UpdateUser(user); err != nil {
		wh.logger.Errorf("Failed to update refresh token for user (id=%d, name=%s): %w", user.ID, user.Name, err)
	}
}
