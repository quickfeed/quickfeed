package web

import (
	pb "github.com/autograde/quickfeed/ag"
)

// getUsers returns all the users in the database.
func (s *AutograderService) getUsers() (*pb.Users, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	return &pb.Users{Users: users}, nil
}

// getUserByCourse returns the user matching the given GitHub login if
// the user is enrolled in the given course.
func (s *AutograderService) getUserByCourse(request *pb.CourseUserRequest, currentUser *pb.User) (*pb.User, error) {
	courseQuery := &pb.Course{Code: request.CourseCode, Year: request.CourseYear}
	user, course, err := s.db.GetUserByCourse(courseQuery, request.UserLogin)
	if err != nil {
		return nil, err
	}
	if !(currentUser.IsAdmin || s.isTeacher(currentUser.ID, course.ID)) {
		return nil, ErrInvalidUserInfo
	}
	return user, nil
}

// updateUser updates the user profile according to the user data in
// the request object. If curUser is admin, and the request may also
// promote the user to admin.
func (s *AutograderService) updateUser(curUser *pb.User, request *pb.User) (*pb.User, error) {
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
