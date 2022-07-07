package database

import "github.com/quickfeed/quickfeed/qf/types"

// GetUser fetches a user by ID with remote identities.
func (db *GormDB) GetUser(userID uint64) (*types.User, error) {
	var user types.User
	if err := db.conn.Preload("RemoteIdentities").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByRemoteIdentity fetches user by remote identity.
func (db *GormDB) GetUserByRemoteIdentity(remote *types.RemoteIdentity) (*types.User, error) {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity types.RemoteIdentity
	if err := tx.
		Where(&types.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get the user.
	var user types.User
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
func (db *GormDB) GetUserByCourse(query *types.Course, login string) (*types.User, *types.Course, error) {
	var user types.User
	var course types.Course
	enrollmentStatuses := []types.Enrollment_UserStatus{
		types.Enrollment_STUDENT,
		types.Enrollment_TEACHER,
	}

	if err := db.conn.First(&course, query).Error; err != nil {
		return nil, nil, err
	}

	if err := db.conn.Preload("Enrollments", "status in (?)", enrollmentStatuses).First(&user, &types.User{Login: login}).Error; err != nil {
		return nil, nil, err
	}
	for _, e := range user.Enrollments {
		if e.CourseID == course.ID {
			user.Enrollments = make([]*types.Enrollment, 0)
			return &user, &course, nil
		}
	}
	return nil, nil, ErrNotEnrolled
}

// GetUserWithEnrollments returns user with the given ID with all enrollments.
func (db *GormDB) GetUserWithEnrollments(userID uint64) (*types.User, error) {
	var user types.User
	if err := db.conn.Preload("Enrollments").Preload("Enrollments.UsedSlipDays").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers fetches all users by provided IDs.
func (db *GormDB) GetUsers(userIDs ...uint64) ([]*types.User, error) {
	m := db.conn
	if len(userIDs) > 0 {
		m = m.Where(userIDs)
	}
	m = m.Preload("RemoteIdentities")

	var users []*types.User
	if err := m.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates user information.
func (db *GormDB) UpdateUser(user *types.User) error {
	if err := db.conn.First(&types.User{ID: user.GetID()}).Error; err != nil {
		return err
	}
	return db.conn.Save(&user).Error
}
