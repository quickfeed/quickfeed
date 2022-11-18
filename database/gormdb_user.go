package database

import "github.com/quickfeed/quickfeed/qf"

// CreateUser creates new user record. The first user is set as admin.
func (db *GormDB) CreateUser(user *qf.User) error {
	if err := db.conn.Create(&user).Error; err != nil {
		return err
	}
	// The first user defaults to admin user.
	if user.ID == 1 {
		user.IsAdmin = true
		if err := db.UpdateUser(user); err != nil {
			return err
		}
	}
	return nil
}

// GetUser returns the given user.
func (db *GormDB) GetUser(userID uint64) (*qf.User, error) {
	var user qf.User
	if err := db.conn.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByRemoteIdentity returns the user for the given remote identity.
func (db *GormDB) GetUserByRemoteIdentity(scmRemoteID uint64) (*qf.User, error) {
	var user qf.User
	if err := db.conn.
		Where(&qf.User{ScmRemoteID: scmRemoteID}).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByCourse returns the given user with enrollments matching the given course query.
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

	if err := db.conn.
		Preload("Enrollments", "status in (?)", enrollmentStatuses).
		First(&user, &qf.User{Login: login}).Error; err != nil {
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

// GetUserWithEnrollments returns the given user with enrollments.
func (db *GormDB) GetUserWithEnrollments(userID uint64) (*qf.User, error) {
	var user qf.User
	if err := db.conn.
		Preload("Enrollments").
		Preload("Enrollments.Course").
		Preload("Enrollments.UsedSlipDays").
		First(&user, userID).Error; err != nil {
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
