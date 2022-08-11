package web

import "github.com/quickfeed/quickfeed/qf"

// getUsers returns all the users in the database.
func (s *QuickFeedService) getUsers() (*qf.Users, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	return &qf.Users{Users: users}, nil
}

// getUserByCourse returns the user matching the given GitHub login if
// the user is enrolled in the given course.
func (s *QuickFeedService) getUserByCourse(request *qf.CourseUserRequest, currentUser *qf.User) (*qf.User, error) {
	courseQuery := &qf.Course{Code: request.CourseCode, Year: request.CourseYear}
	user, _, err := s.db.GetUserByCourse(courseQuery, request.UserLogin)
	if err != nil {
		return nil, err
	}
	return user, nil
}

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
