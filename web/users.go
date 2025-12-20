package web

import (
	"context"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// editUserProfile updates the user profile according to the user data in
// the request object. If curUser is admin, and the request may also
// promote the user to admin.
func (s *QuickFeedService) editUserProfile(curUser, request *qf.User) error {
	updateUser, err := s.db.GetUser(request.GetID())
	if err != nil {
		return err
	}

	if request.GetName() != "" {
		updateUser.Name = request.GetName()
	}
	if request.GetStudentID() != "" {
		updateUser.StudentID = request.GetStudentID()
	}
	if request.GetEmail() != "" {
		updateUser.Email = request.GetEmail()
	}
	if request.GetAvatarURL() != "" {
		updateUser.AvatarURL = request.GetAvatarURL()
	}

	// log every change to admin state
	if updateUser.GetIsAdmin() != request.GetIsAdmin() {
		s.logger.Debugf("User %s attempting to change admin status of user %s to %v", curUser.GetLogin(), updateUser.GetLogin(), request.GetIsAdmin())
	}
	// current user must be admin to change admin status of another user
	// admin status of super admin (user with ID 1) cannot be changed
	if curUser.GetIsAdmin() && request.GetID() > 1 {
		updateUser.IsAdmin = request.GetIsAdmin()
	}
	return s.db.UpdateUser(updateUser)
}

// updateUserFromSCM fetches the latest user info from the SCM and updates the local user
// record in the database. This should be used ahead of operations that require valid SCM
// user info, such as adding users to organizations or teams.
func (s *QuickFeedService) updateUserFromSCM(ctx context.Context, sc scm.SCM, user *qf.User) error {
	ghUser, err := sc.GetUserByID(ctx, user.GetScmRemoteID())
	if err != nil {
		return err
	}
	ghLogin := ghUser.GetLogin()
	if ghLogin != "" && ghLogin != user.GetLogin() {
		s.logger.Infof("Updating SCM login for user ID %d from %q to %q", user.GetID(), user.GetLogin(), ghLogin)
		user.Login = ghLogin
	}
	ghAvatar := ghUser.GetAvatarURL()
	if ghAvatar != "" && ghAvatar != user.GetAvatarURL() {
		user.AvatarURL = ghAvatar
	}
	return s.db.UpdateUser(user)
}
