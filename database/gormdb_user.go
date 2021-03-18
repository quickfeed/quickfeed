package database

import (
	pb "github.com/autograde/quickfeed/ag"
)

// GetUser fetches a user by ID with remote identities.
func (db *GormDB) GetUser(userID uint64) (*pb.User, error) {
	var user pb.User
	if err := db.conn.Preload("RemoteIdentities").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByAccessToken fetches user ID by access token
func (db *GormDB) GetUserByAccessToken(token string) (uint64, error) {
	tx := db.conn.Begin()

	var remoteIdentity pb.RemoteIdentity
	if err := tx.
		Where(
			&pb.RemoteIdentity{
				AccessToken: token,
			}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return remoteIdentity.UserID, nil
}

// GetUserByRemoteIdentity fetches user by remote identity.
func (db *GormDB) GetUserByRemoteIdentity(remote *pb.RemoteIdentity) (*pb.User, error) {
	tx := db.conn.Begin()
	// Get the remote identity.
	var remoteIdentity pb.RemoteIdentity
	if err := tx.
		Where(&pb.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get the user.
	var user pb.User
	if err := tx.Preload("RemoteIdentities").First(&user, remoteIdentity.UserID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByCourse returns user and course matching the provided course query
// and the provided user login name.
func (db *GormDB) GetUserByCourse(query *pb.Course, login string) (*pb.User, *pb.Course, error) {
	var user pb.User
	var course pb.Course
	enrollmentStatuses := []pb.Enrollment_UserStatus{
		pb.Enrollment_STUDENT,
		pb.Enrollment_TEACHER,
	}

	if err := db.conn.First(&course, query).Error; err != nil {
		return nil, nil, err
	}

	if err := db.conn.Preload("Enrollments", "status in (?)", enrollmentStatuses).First(&user, &pb.User{Login: login}).Error; err != nil {
		return nil, nil, err
	}
	for _, e := range user.Enrollments {

		if e.CourseID == course.ID {
			user.Enrollments = make([]*pb.Enrollment, 0)
			return &user, &course, nil
		}
	}
	return nil, nil, ErrNotEnrolled
}

// GetUserWithEnrollments returns user with the given ID with all enrollments.
func (db *GormDB) GetUserWithEnrollments(userID uint64) (*pb.User, error) {
	var user pb.User
	if err := db.conn.Preload("Enrollments").Preload("Enrollments.UsedSlipDays").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers fetches all users by provided IDs.
func (db *GormDB) GetUsers(userIDs ...uint64) ([]*pb.User, error) {
	m := db.conn
	if len(userIDs) > 0 {
		m = m.Where(userIDs)
	}
	m = m.Preload("RemoteIdentities")

	var users []*pb.User
	if err := m.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates user information.
func (db *GormDB) UpdateUser(user *pb.User) error {
	if err := db.conn.First(&pb.User{ID: user.GetID()}).Error; err != nil {
		return err
	}
	return db.conn.Save(&user).Error
}
