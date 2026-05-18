package database

import (
	"context"
	"database/sql"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
)

// CreateUser creates new user record. The first user is set as admin.
func (db *BunDB) CreateUser(user *qf.User) error {
	ctx := context.Background()

	if _, err := db.conn.NewInsert().Model(user).Exec(ctx); err != nil {
		return err
	}
	// The first user defaults to admin user.
	if user.GetID() == 1 {
		user.IsAdmin = true
		if err := db.UpdateUser(user); err != nil {
			return err
		}
	}
	return nil
}

// GetUser returns the given user.
func (db *BunDB) GetUser(userID uint64) (*qf.User, error) {
	ctx := context.Background()

	var user qf.User
	if err := db.conn.NewSelect().Model(&user).Where("id = ?", userID).Scan(ctx); err != nil {
		return nil, toDBError(err)
	}
	return &user, nil
}

// GetUserByRemoteIdentity returns the user for the given remote identity.
func (db *BunDB) GetUserByRemoteIdentity(scmRemoteID uint64) (*qf.User, error) {
	ctx := context.Background()

	var user qf.User
	if err := db.conn.NewSelect().Model(&user).Where("scm_remote_id = ?", scmRemoteID).Scan(ctx); err != nil {
		return nil, toDBError(err)
	}
	return &user, nil
}

// GetUserByCourse returns the given user with enrollments matching the given course query.
func (db *BunDB) GetUserByCourse(query *qf.Course, login string) (*qf.User, error) {
	ctx := context.Background()

	var user qf.User
	var course qf.Course
	enrollmentStatuses := []qf.Enrollment_UserStatus{
		qf.Enrollment_STUDENT,
		qf.Enrollment_TEACHER,
	}

	if err := db.conn.NewSelect().Model(&course).Where("id = ?", query.GetID()).Scan(ctx); err != nil {
		return nil, err
	}

	if err := db.conn.NewSelect().Model(&user).
		Relation("Enrollments", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("status IN (?)", bun.In(enrollmentStatuses))
		}).
		Where("login = ?", login).
		Scan(ctx); err != nil {
		return nil, err
	}
	for _, e := range user.GetEnrollments() {
		if e.GetCourseID() == course.GetID() {
			user.Enrollments = make([]*qf.Enrollment, 0)
			return &user, nil
		}
	}
	return nil, ErrNotEnrolled
}

// GetUserWithEnrollments returns the given user with enrollments.
func (db *BunDB) GetUserWithEnrollments(userID uint64) (*qf.User, error) {
	ctx := context.Background()

	var user qf.User
	if err := db.conn.
		NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Relation("Enrollments").
		Relation("Enrollments.Course").
		Relation("Enrollments.UsedSlipDays").
		Relation("FeedbackReceipts").
		Scan(ctx); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers fetches all users by provided IDs.
func (db *BunDB) GetUsers(userIDs ...uint64) ([]*qf.User, error) {
	ctx := context.Background()

	var users []*qf.User

	m := db.conn.NewSelect().Model(&users)
	if len(userIDs) > 0 {
		m = m.Where("id IN (?)", bun.In(userIDs))
	}
	if err := m.Scan(ctx); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates user information.
func (db *BunDB) UpdateUser(user *qf.User) error {
	ctx := context.Background()

	result, err := db.conn.NewUpdate().Model(user).WherePK().Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
