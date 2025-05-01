package web

import "github.com/quickfeed/quickfeed/qf"

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
