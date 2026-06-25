package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
)

// CreateEnrollment creates a new pending enrollment.
func (db *BunDB) CreateEnrollment(enrollment *qf.Enrollment) error {
	ctx := context.Background()
	var user qf.User
	if err := db.conn.NewSelect().Model(&user).Where("id = ?", enrollment.GetUserID()).Scan(ctx); err != nil {
		return err
	}
	if err := user.ValidateProfile(); err != nil {
		return errors.Join(ErrIncompleteProfile, err)
	}

	exists, err := db.conn.NewSelect().Model((*qf.Course)(nil)).Where("id = ?", enrollment.GetCourseID()).Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	enrollment.Status = qf.Enrollment_PENDING
	enrollment.State = qf.Enrollment_VISIBLE
	_, err = db.conn.NewInsert().Model(enrollment).Exec(ctx)
	return err
}

// RejectEnrollment removes the user enrollment from the database.
func (db *BunDB) RejectEnrollment(userID, courseID uint64) error {
	ctx := context.Background()
	enrol, err := db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		return err
	}
	_, err = db.conn.NewDelete().Model(enrol).WherePK().Exec(ctx)
	return err
}

// UpdateEnrollment changes status and display state of the given enrollment.
func (db *BunDB) UpdateEnrollment(enrol *qf.Enrollment) error {
	if enrol.GetID() == 0 {
		return errors.New("enrollment query missing primary key: ID")
	}
	ctx := context.Background()
	_, err := db.conn.NewUpdate().Model(enrol).
		Where("course_id = ? AND user_id = ?", enrol.GetCourseID(), enrol.GetUserID()).
		Exec(ctx)
	return err
}

// GetEnrollmentByCourseAndUser returns a user enrollment for the given course and user IDs.
func (db *BunDB) GetEnrollmentByCourseAndUser(courseID, userID uint64) (*qf.Enrollment, error) {
	ctx := context.Background()
	var enrollment qf.Enrollment
	if err := db.conn.NewSelect().
		Model(&enrollment).
		Relation("Course").
		Relation("User").
		Relation("UsedSlipDays").
		Where("enrollment.course_id = ? AND enrollment.user_id = ?", courseID, userID).
		Scan(ctx); err != nil {
		return nil, err
	}
	enrollment.SetSlipDays(enrollment.GetCourse())
	return &enrollment, nil
}

// GetEnrollmentsByCourse fetches all course enrollments with given statuses.
func (db *BunDB) GetEnrollmentsByCourse(courseID uint64, statuses ...qf.Enrollment_UserStatus) ([]*qf.Enrollment, error) {
	return db.getEnrollments("course", courseID, statuses...)
}

// GetEnrollmentsByUser returns all existing enrollments for the given user.
func (db *BunDB) GetEnrollmentsByUser(userID uint64, statuses ...qf.Enrollment_UserStatus) ([]*qf.Enrollment, error) {
	return db.getEnrollments("user", userID, statuses...)
}

// getEnrollments is a generic helper that returns enrollments filtered by course or user.
func (db *BunDB) getEnrollments(by string, id uint64, statuses ...qf.Enrollment_UserStatus) ([]*qf.Enrollment, error) {
	ctx := context.Background()
	if len(statuses) == 0 {
		statuses = []qf.Enrollment_UserStatus{
			qf.Enrollment_PENDING,
			qf.Enrollment_STUDENT,
			qf.Enrollment_TEACHER,
		}
	}
	var enrollments []*qf.Enrollment
	q := db.conn.NewSelect().
		Model(&enrollments).
		Relation("User").
		Relation("Course").
		Relation("Group").
		Relation("Group.Users").
		Relation("UsedSlipDays").
		Where("enrollment.status IN (?)", bun.In(statuses))
	switch by {
	case "user":
		q = q.Where("enrollment.user_id = ?", id)
	case "course":
		q = q.Where("enrollment.course_id = ?", id)
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	for _, enrollment := range enrollments {
		enrollment.SetSlipDays(enrollment.GetCourse())
	}
	return enrollments, nil
}

// UpdateSlipDays updates used slip days for the given course enrollment.
func (db *BunDB) UpdateSlipDays(usedSlipDays []*qf.UsedSlipDays) error {
	ctx := context.Background()
	for _, slipDays := range usedSlipDays {
		if slipDays.GetID() == 0 {
			if _, err := db.conn.NewInsert().Model(slipDays).Exec(ctx); err != nil {
				return err
			}
		} else {
			if _, err := db.conn.NewUpdate().Model(slipDays).WherePK().Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
