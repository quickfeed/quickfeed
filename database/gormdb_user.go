package database

import "github.com/quickfeed/quickfeed/qf"

// GetUser fetches a user by ID with remote identities.
func (db *GormDB) GetUser(userID uint64) (*qf.User, error) {
	var user qf.User
	if err := db.conn.Preload("RemoteIdentities").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByRemoteIdentity fetches user by remote identity.
func (db *GormDB) GetUserByRemoteIdentity(remote *qf.RemoteIdentity) (*qf.User, error) {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity qf.RemoteIdentity
	if err := tx.
		Where(&qf.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get the user.
	var user qf.User
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
func (db *GormDB) GetUserByCourse(query *qf.Course, login string) (*qf.User, error) {
	var user qf.User
	var course qf.Course
	enrollmentStatuses := []qf.Enrollment_UserStatus{
		qf.Enrollment_STUDENT,
		qf.Enrollment_TEACHER,
	}

	if err := db.conn.First(&course, query).Error; err != nil {
		return nil, err
	}

	if err := db.conn.Preload("Enrollments", "status in (?)", enrollmentStatuses).First(&user, &qf.User{Login: login}).Error; err != nil {
		return nil, err
	}
	for _, e := range user.Enrollments {
		if e.CourseID == course.ID {
			user.Enrollments = make([]*qf.Enrollment, 0)
			return &user, nil
		}
	}
	return nil, ErrNotEnrolled
}

// GetUserWithEnrollments returns user with the given ID with all enrollments.
func (db *GormDB) GetUserWithEnrollments(userID uint64) (*qf.User, error) {
	var user qf.User
	if err := db.conn.Preload("Enrollments").Preload("Enrollments.UsedSlipDays").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers fetches all users by provided IDs.
func (db *GormDB) GetUsers(userIDs ...uint64) ([]*qf.User, error) {
	m := db.conn
	if len(userIDs) > 0 {
		m = m.Where(userIDs)
	}
	m = m.Preload("RemoteIdentities")

	var users []*qf.User
	if err := m.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates user information.
func (db *GormDB) UpdateUser(user *qf.User) error {
	if err := db.conn.First(&qf.User{ID: user.GetID()}).Error; err != nil {
		return err
	}
	return db.conn.Save(&user).Error
}
