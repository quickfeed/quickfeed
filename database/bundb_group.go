package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
)

// CreateGroup creates a new group and assigns users to the newly created group.
func (db *BunDB) CreateGroup(group *qf.Group) error {
	ctx := context.Background()
	if len(group.GetUsers()) == 0 {
		return ErrEmptyGroup
	}
	if group.GetCourseID() == 0 {
		return sql.ErrNoRows
	}
	exists, err := db.conn.NewSelect().Model((*qf.Course)(nil)).Where("id = ?", group.GetCourseID()).Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Always create a new row; callers may reuse the same struct instance.
		group.ID = 0
		if _, err := tx.NewInsert().Model(group).Exec(ctx); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return ErrDuplicateGroup
			}
			return err
		}

		var userIDs []uint64
		for _, u := range group.GetUsers() {
			userIDs = append(userIDs, u.GetID())
		}

		// Link users to the group via the group_users join table.
		for _, uid := range userIDs {
			if _, err := tx.Exec("INSERT INTO group_users (group_id, user_id) VALUES (?, ?)",
				group.GetID(), uid); err != nil {
				return err
			}
		}

		// Update enrollments to reflect the new group.
		res, err := tx.NewUpdate().Model((*qf.Enrollment)(nil)).
			Set("group_id = ?", group.GetID()).
			Where("course_id = ? AND user_id IN (?) AND status IN (?)",
				group.GetCourseID(),
				bun.In(userIDs),
				bun.In([]qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_TEACHER})).
			Exec(ctx)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected != int64(len(userIDs)) {
			return ErrUpdateGroup
		}
		return nil
	})
}

// UpdateGroup updates a group with the specified users and enrollments.
func (db *BunDB) UpdateGroup(group *qf.Group) error {
	ctx := context.Background()
	if group.GetCourseID() == 0 {
		return sql.ErrNoRows
	}
	exists, err := db.conn.NewSelect().Model((*qf.Course)(nil)).Where("id = ?", group.GetCourseID()).Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewUpdate().Model(group).WherePK().Exec(ctx); err != nil {
			return err
		}

		// Replace Users association: clear join table, then re-insert.
		if _, err := tx.NewDelete().TableExpr("group_users").
			Where("group_id = ?", group.GetID()).Exec(ctx); err != nil {
			return err
		}
		for _, u := range group.GetUsers() {
			if _, err := tx.Exec("INSERT INTO group_users (group_id, user_id) VALUES (?, ?)",
				group.GetID(), u.GetID()); err != nil {
				return err
			}
		}

		// Clear group_id for all previous members.
		if _, err := tx.NewUpdate().Model((*qf.Enrollment)(nil)).
			Set("group_id = 0").
			Where("group_id = ?", group.GetID()).
			Exec(ctx); err != nil {
			return err
		}

		userIDs := group.UserIDs()
		if err := bunSyncGroupGrades(ctx, tx, group.GetID(), userIDs); err != nil {
			return err
		}

		// Set group_id for the current members.
		res, err := tx.NewUpdate().Model((*qf.Enrollment)(nil)).
			Set("group_id = ?", group.GetID()).
			Where("course_id = ? AND user_id IN (?)", group.GetCourseID(), bun.In(userIDs)).
			Exec(ctx)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected != int64(len(userIDs)) {
			return ErrUpdateGroup
		}
		return nil
	})
}

// UpdateGroupStatus updates the status field of a group.
func (db *BunDB) UpdateGroupStatus(group *qf.Group) error {
	ctx := context.Background()
	_, err := db.conn.NewUpdate().Model(group).Column("status").WherePK().Exec(ctx)
	return err
}

// DeleteGroup deletes a group and clears its corresponding enrollments.
func (db *BunDB) DeleteGroup(groupID uint64) error {
	ctx := context.Background()
	group, err := db.GetGroup(groupID)
	if err != nil {
		return err
	}
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Clear the group_users join table.
		if _, err := tx.NewDelete().TableExpr("group_users").
			Where("group_id = ?", groupID).Exec(ctx); err != nil {
			return err
		}
		// Clear group_id on all associated enrollments.
		if _, err := tx.NewUpdate().Model((*qf.Enrollment)(nil)).
			Set("group_id = 0").
			Where("group_id = ?", groupID).
			Exec(ctx); err != nil {
			return err
		}
		_, err := tx.NewDelete().Model(group).WherePK().Exec(ctx)
		return err
	})
}

// GetGroup returns the group with the specified group ID.
func (db *BunDB) GetGroup(groupID uint64) (*qf.Group, error) {
	ctx := context.Background()
	var userIDs []uint64
	if err := db.conn.NewSelect().Model((*qf.Enrollment)(nil)).
		Column("user_id").
		Where("group_id = ?", groupID).
		Scan(ctx, &userIDs); err != nil {
		return nil, err
	}

	var group qf.Group
	q := db.conn.NewSelect().
		Model(&group).
		Relation("Enrollments").
		Relation("Enrollments.UsedSlipDays").
		Relation("Enrollments.User").
		Where("id = ?", groupID)
	if len(userIDs) > 0 {
		q = q.Relation("Users", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("user.id IN (?)", bun.In(userIDs))
		})
	}
	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get group with ID %d: %w", groupID, err)
	}
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("no users found for group with ID %d", groupID)
	}
	return &group, nil
}

// GetGroupsByCourse returns the groups for the given course.
func (db *BunDB) GetGroupsByCourse(courseID uint64, statuses ...qf.Group_GroupStatus) ([]*qf.Group, error) {
	ctx := context.Background()
	if len(statuses) == 0 {
		statuses = []qf.Group_GroupStatus{
			qf.Group_PENDING,
			qf.Group_APPROVED,
		}
	}
	var groups []*qf.Group
	if err := db.conn.NewSelect().
		Model(&groups).
		Relation("Enrollments").
		Relation("Enrollments.UsedSlipDays").
		Relation("Enrollments.User").
		Relation("Users").
		Where("course_id = ? AND status IN (?)", courseID, bun.In(statuses)).
		Scan(ctx); err != nil {
		return nil, err
	}
	return groups, nil
}

// bunSyncGroupGrades synchronizes grade records for all group submissions
// to match current group membership.
func bunSyncGroupGrades(ctx context.Context, tx bun.Tx, groupID uint64, userIDs []uint64) error {
	var submissions []*qf.Submission
	if err := tx.NewSelect().
		Model(&submissions).
		Relation("Grades").
		Where("submission.group_id = ?", groupID).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	for _, submission := range submissions {
		existingGrades := make(map[uint64]*qf.Grade)
		for _, grade := range submission.GetGrades() {
			existingGrades[grade.GetUserID()] = grade
		}

		submission.Grades = make([]*qf.Grade, len(userIDs))
		for i, userID := range userIDs {
			if existing, found := existingGrades[userID]; found {
				submission.Grades[i] = existing
			} else {
				submission.Grades[i] = &qf.Grade{UserID: userID, SubmissionID: submission.GetID()}
			}
		}

		var assignment qf.Assignment
		if err := tx.NewSelect().Model(&assignment).Where("id = ?", submission.GetAssignmentID()).Scan(ctx); err != nil {
			return err
		}
		submission.SetGradesIfApproved(&assignment, submission.GetScore())

		if _, err := tx.NewDelete().Model((*qf.Grade)(nil)).
			Where("submission_id = ?", submission.GetID()).Exec(ctx); err != nil {
			return err
		}
		if len(submission.Grades) > 0 {
			if _, err := tx.NewInsert().Model(&submission.Grades).Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
