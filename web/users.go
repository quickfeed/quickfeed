package web

import "github.com/quickfeed/quickfeed/qf"

// updateUser updates the user profile according to the user data in
// the request object. If curUser is admin, and the request may also
// promote the user to admin.
func (s *QuickFeedService) updateUser(curUser *qf.User, request *qf.User) (*qf.User, error) {
	updateUser, err := s.db.GetUser(request.ID)
	if err != nil {
		return nil, err
	}

	if request.Name != "" {
		updateUser.Name = request.Name
	}
	if request.StudentID != "" {
		updateUser.StudentID = request.StudentID
	}
	if request.Email != "" {
		updateUser.Email = request.Email
	}
	if request.AvatarURL != "" {
		updateUser.AvatarURL = request.AvatarURL
	}

	// log every change to admin state
	if updateUser.IsAdmin != request.IsAdmin {
		s.logger.Debugf("User %s attempting to change admin status of user %s to %v", curUser.Login, updateUser.Login, request.IsAdmin)
	}
	// current user must be admin to change admin status of another user
	// admin status of super admin (user with ID 1) cannot be changed
	if curUser.IsAdmin && request.ID > 1 {
		updateUser.IsAdmin = request.IsAdmin
	}

	err = s.db.UpdateUser(updateUser)
	return updateUser, err
}
