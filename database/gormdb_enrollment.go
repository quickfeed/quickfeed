package database

import (
	pb "github.com/autograde/aguis/ag"
	"github.com/jinzhu/gorm"
)

// CreateEnrollment creates a new pending enrollment.
func (db *GormDB) CreateEnrollment(enrollment *pb.Enrollment) error {
	var user, course uint64
	if err := db.conn.Model(&pb.User{}).Where(&pb.User{
		ID: enrollment.UserID,
	}).Count(&user).Error; err != nil {
		return err
	}
	if err := db.conn.Model(&pb.Course{}).Where(&pb.Course{
		ID: enrollment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if user+course != 2 {
		return gorm.ErrRecordNotFound
	}

	enrollment.Status = pb.Enrollment_PENDING
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

// UpdateEnrollmentStatus changes status of an enrollment of the given user ID in the given course ID.
func (db *GormDB) UpdateEnrollmentStatus(userID, courseID uint64, status pb.Enrollment_UserStatus) error {
	return db.setEnrollment(userID, courseID, status)
}

// GetEnrollmentByCourseAndUser returns a user enrollment for the given course ID.
func (db *GormDB) GetEnrollmentByCourseAndUser(courseID uint64, userID uint64) (*pb.Enrollment, error) {
	var enrollment pb.Enrollment
	m := db.conn.Preload("Course").Preload("User")
	if err := m.
		Where(&pb.Enrollment{
			CourseID: courseID,
			UserID:   userID,
		}).
		First(&enrollment).Error; err != nil {
		return nil, err
	}
	return &enrollment, nil
}

// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
func (db *GormDB) GetEnrollmentsByCourse(courseID uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error) {
	return db.getEnrollments(&pb.Course{ID: courseID}, statuses...)
}

// GetEnrollmentsByUser returns all existing enrollments for the given user
func (db *GormDB) GetEnrollmentsByUser(userID uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error) {
	return db.getEnrollments(&pb.User{ID: userID}, statuses...)
}

// getEnrollments is generic helper function that return enrollments for either course and user.
func (db *GormDB) getEnrollments(model interface{}, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error) {
	if len(statuses) == 0 {
		statuses = []pb.Enrollment_UserStatus{
			pb.Enrollment_PENDING,
			pb.Enrollment_STUDENT,
			pb.Enrollment_TEACHER,
		}
	}
	var enrollments []*pb.Enrollment
	if err := db.conn.Preload("User").Preload("Course").Preload("Group").Model(model).
		Where("status in (?)", statuses).
		Association("Enrollments").
		Find(&enrollments).Error; err != nil {
		return nil, err
	}
	return enrollments, nil
}

// setEnrollment updates enrollment status.
func (db *GormDB) setEnrollment(userID, courseID uint64, status pb.Enrollment_UserStatus) error {
	return db.conn.
		Model(&pb.Enrollment{}).
		Where(&pb.Enrollment{CourseID: courseID, UserID: userID}).
		Update(&pb.Enrollment{Status: status}).Error
}
