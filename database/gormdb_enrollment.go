package database

import (
	"errors"

	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

// CreateEnrollment creates a new pending enrollment.
func (db *GormDB) CreateEnrollment(enrollment *qf.Enrollment) error {
	var user, course int64
	if err := db.conn.Model(&qf.User{}).Where(&qf.User{
		ID: enrollment.GetUserID(),
	}).Count(&user).Error; err != nil {
		return err
	}
	if err := db.conn.Model(&qf.Course{}).Where(&qf.Course{
		ID: enrollment.GetCourseID(),
	}).Count(&course).Error; err != nil {
		return err
	}
	if user+course != 2 {
		return gorm.ErrRecordNotFound
	}

	enrollment.Status = qf.Enrollment_PENDING
	enrollment.State = qf.Enrollment_VISIBLE
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
func (db *GormDB) UpdateEnrollment(enrol *qf.Enrollment) error {
	// If enrol.ID is zero, Select("*").Updates() would update ID of the Enrollment record to zero.
	if enrol.GetID() == 0 {
		return errors.New("enrollment query missing primary key: ID")
	}
	return db.conn.Model(&qf.Enrollment{}).
		Select("*").
		Where(&qf.Enrollment{
			CourseID: enrol.GetCourseID(),
			UserID:   enrol.GetUserID(),
		}).Updates(enrol).Error
}

// GetEnrollmentByCourseAndUser returns a user enrollment for the given course ID.
func (db *GormDB) GetEnrollmentByCourseAndUser(courseID uint64, userID uint64) (*qf.Enrollment, error) {
	var enrollment qf.Enrollment
	m := db.conn.Preload("Course").Preload("User").Preload("UsedSlipDays")
	if err := m.
		Where(&qf.Enrollment{
			CourseID: courseID,
			UserID:   userID,
		}).
		First(&enrollment).Error; err != nil {
		return nil, err
	}
	enrollment.SetSlipDays(enrollment.GetCourse())
	return &enrollment, nil
}

// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
func (db *GormDB) GetEnrollmentsByCourse(courseID uint64, statuses ...qf.Enrollment_UserStatus) ([]*qf.Enrollment, error) {
	return db.getEnrollments(&qf.Course{ID: courseID}, statuses...)
}

// GetEnrollmentsByUser returns all existing enrollments for the given user
func (db *GormDB) GetEnrollmentsByUser(userID uint64, statuses ...qf.Enrollment_UserStatus) ([]*qf.Enrollment, error) {
	return db.getEnrollments(&qf.User{ID: userID}, statuses...)
}

// getEnrollments is generic helper function that return enrollments for either course and user.
func (db *GormDB) getEnrollments(model interface{}, statuses ...qf.Enrollment_UserStatus) ([]*qf.Enrollment, error) {
	if len(statuses) == 0 {
		statuses = []qf.Enrollment_UserStatus{
			qf.Enrollment_PENDING,
			qf.Enrollment_STUDENT,
			qf.Enrollment_TEACHER,
		}
	}
	var enrollments []*qf.Enrollment
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
	for _, enrollment := range enrollments {
		enrollment.SetSlipDays(enrollment.GetCourse())
	}
	return enrollments, nil
}

// UpdateSlipDays updates used slip days for the given course enrollment
func (db *GormDB) UpdateSlipDays(usedSlipDays []*qf.UsedSlipDays) error {
	for _, slipDaysForAssignment := range usedSlipDays {
		if err := db.internalUpdateSlipDays(slipDaysForAssignment); err != nil {
			return err
		}
	}
	return nil
}

// internalUpdateSlipDays updates or creates UsedSlipDays record
func (db *GormDB) internalUpdateSlipDays(query *qf.UsedSlipDays) error {
	return db.conn.Save(query).Error
}
