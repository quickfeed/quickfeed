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
func (db *GormDB) RejectEnrollment(uid, cid uint64) error {
	enrol, err := db.GetEnrollmentByCourseAndUser(cid, uid)
	if err != nil {
		return err
	}
	return db.conn.Delete(enrol).Error
}

// EnrollStudent enrolls user as course student.
func (db *GormDB) EnrollStudent(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, pb.Enrollment_STUDENT)
}

// EnrollTeacher enrolls user as course teacher.
func (db *GormDB) EnrollTeacher(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, pb.Enrollment_TEACHER)
}

// SetPendingEnrollment sets enrollment status to pending.
func (db *GormDB) SetPendingEnrollment(uid, cid uint64) error {
	return db.setEnrollment(uid, cid, pb.Enrollment_PENDING)
}

// UpdateGroupEnrollment sets GroupID of a student enrollment to 0.
func (db *GormDB) UpdateGroupEnrollment(uid, cid uint64) error {
	return db.conn.
		Model(&pb.Enrollment{}).
		Where(&pb.Enrollment{CourseID: cid, UserID: uid}).
		Update("group_id", uint64(0)).Error
}

// GetEnrollmentByCourseAndUser returns a user enrollment for the given course ID.
func (db *GormDB) GetEnrollmentByCourseAndUser(cid uint64, uid uint64) (*pb.Enrollment, error) {
	var enrollment pb.Enrollment
	m := db.conn.Preload("Course").Preload("User")
	if err := m.
		Where(&pb.Enrollment{
			CourseID: cid,
			UserID:   uid,
		}).
		First(&enrollment).Error; err != nil {
		return nil, err
	}
	return &enrollment, nil
}

// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
func (db *GormDB) GetEnrollmentsByCourse(cid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Enrollment, error) {
	return db.getEnrollments(&pb.Course{ID: cid}, statuses...)
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
	if err := db.conn.Preload("User").Preload("Course").Model(model).
		Where("status in (?)", statuses).
		Association("Enrollments").
		Find(&enrollments).Error; err != nil {
		return nil, err
	}
	return enrollments, nil
}

// setEnrollment updates enrollment status.
func (db *GormDB) setEnrollment(uid, cid uint64, status pb.Enrollment_UserStatus) error {
	return db.conn.
		Model(&pb.Enrollment{}).
		Where(&pb.Enrollment{CourseID: cid, UserID: uid}).
		Update(&pb.Enrollment{Status: status}).Error
}
