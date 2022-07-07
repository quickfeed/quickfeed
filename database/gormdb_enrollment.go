package database

import (
	"github.com/quickfeed/quickfeed/qf/types"
	"gorm.io/gorm"
)

// CreateEnrollment creates a new pending enrollment.
func (db *GormDB) CreateEnrollment(enrollment *types.Enrollment) error {
	var user, course int64
	if err := db.conn.Model(&types.User{}).Where(&types.User{
		ID: enrollment.UserID,
	}).Count(&user).Error; err != nil {
		return err
	}
	if err := db.conn.Model(&types.Course{}).Where(&types.Course{
		ID: enrollment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if user+course != 2 {
		return gorm.ErrRecordNotFound
	}

	enrollment.Status = types.Enrollment_PENDING
	enrollment.State = types.Enrollment_VISIBLE
	return db.conn.Create(&enrollment).Error
}

// RejectEnrollment removes the user enrollment from the database.
func (db *GormDB) RejectEnrollment(userID, courseID uint64) error {
	enrol, err := db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		return err
	}
	return db.conn.Delete(enrol).Error
}

// UpdateEnrollment changes status and display state of the given enrollment.
func (db *GormDB) UpdateEnrollment(enrol *types.Enrollment) error {
	return db.conn.Model(&types.Enrollment{}).
		Where(&types.Enrollment{CourseID: enrol.CourseID, UserID: enrol.UserID}).
		Updates(&types.Enrollment{State: enrol.State, Status: enrol.Status, LastActivityDate: enrol.LastActivityDate}).Error
}

// GetEnrollmentByCourseAndUser returns a user enrollment for the given course ID.
func (db *GormDB) GetEnrollmentByCourseAndUser(courseID uint64, userID uint64) (*types.Enrollment, error) {
	var enrollment types.Enrollment
	m := db.conn.Preload("Course").Preload("User").Preload("UsedSlipDays")
	if err := m.
		Where(&types.Enrollment{
			CourseID: courseID,
			UserID:   userID,
		}).
		First(&enrollment).Error; err != nil {
		return nil, err
	}
	return &enrollment, nil
}

// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
func (db *GormDB) GetEnrollmentsByCourse(courseID uint64, statuses ...types.Enrollment_UserStatus) ([]*types.Enrollment, error) {
	return db.getEnrollments(&types.Course{ID: courseID}, statuses...)
}

// GetEnrollmentsByUser returns all existing enrollments for the given user
func (db *GormDB) GetEnrollmentsByUser(userID uint64, statuses ...types.Enrollment_UserStatus) ([]*types.Enrollment, error) {
	return db.getEnrollments(&types.User{ID: userID}, statuses...)
}

// getEnrollments is generic helper function that return enrollments for either course and user.
func (db *GormDB) getEnrollments(model interface{}, statuses ...types.Enrollment_UserStatus) ([]*types.Enrollment, error) {
	if len(statuses) == 0 {
		statuses = []types.Enrollment_UserStatus{
			types.Enrollment_PENDING,
			types.Enrollment_STUDENT,
			types.Enrollment_TEACHER,
		}
	}
	var enrollments []*types.Enrollment
	if err := db.conn.Preload("User").
		Preload("Course").
		Preload("Group").
		Preload("UsedSlipDays").
		Model(model).
		Where("status in (?)", statuses).
		Association("Enrollments").
		Find(&enrollments); err != nil {
		return nil, err
	}
	return enrollments, nil
}

// UpdateSlipDays updates used slip days for the given course enrollment
func (db *GormDB) UpdateSlipDays(usedSlipDays []*types.UsedSlipDays) error {
	for _, slipDaysForAssignment := range usedSlipDays {
		if err := db.internalUpdateSlipDays(slipDaysForAssignment); err != nil {
			return err
		}
	}
	return nil
}

// internalUpdateSlipdays updates or creates UsedSlipDays record
func (db *GormDB) internalUpdateSlipDays(query *types.UsedSlipDays) error {
	return db.conn.Save(query).Error
}
