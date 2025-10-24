package database

import (
	"fmt"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

// CreateGroup creates a new group and assign users to newly created group.
func (db *GormDB) CreateGroup(group *qf.Group) error {
	if len(group.GetUsers()) == 0 {
		return ErrEmptyGroup
	}
	if group.GetCourseID() == 0 {
		return gorm.ErrRecordNotFound
	}
	var course int64
	if err := db.conn.Model(&qf.Course{}).
		Where(&qf.Course{ID: group.GetCourseID()}).
		Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	tx := db.conn.Begin()
	if err := tx.Model(&qf.Group{}).Create(group).Error; err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateGroup
		}
		return err
	}

	var userids []uint64
	for _, u := range group.GetUsers() {
		userids = append(userids, u.GetID())
	}
	query := tx.Model(&qf.Enrollment{}).
		Where(&qf.Enrollment{CourseID: group.GetCourseID()}).
		Where("user_id IN (?) AND status IN (?)", userids,
			[]qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_TEACHER}).
		Updates(&qf.Enrollment{GroupID: group.GetID()})
	if query.Error != nil {
		tx.Rollback()
		return query.Error
	}
	if query.RowsAffected != int64(len(userids)) {
		tx.Rollback()
		return ErrUpdateGroup
	}

	tx.Commit()
	return nil
}

// UpdateGroup updates a group with the specified users and enrollments.
func (db *GormDB) UpdateGroup(group *qf.Group) error {
	if group.GetCourseID() == 0 {
		return gorm.ErrRecordNotFound
	}
	var course int64
	if err := db.conn.Model(&qf.Course{}).
		Where(&qf.Course{ID: group.GetCourseID()}).
		Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	tx := db.conn.Begin()
	if err := tx.Model(group).Select("*").Updates(group).Error; err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateGroup
		}
		return err
	}
	// Set group ID to zero to remove all enrollments from the given group to safely add all members of the incoming group request.
	if err := tx.Exec("UPDATE enrollments SET group_id= ? WHERE group_id= ?", 0, group.GetID()).Error; err != nil {
		tx.Rollback()
		return err
	}

	var userids []uint64
	for _, u := range group.GetUsers() {
		userids = append(userids, u.GetID())
	}
	query := tx.Model(&qf.Enrollment{}).
		Where(&qf.Enrollment{CourseID: group.GetCourseID()}).
		Where("user_id IN (?)", userids).
		Where("status IN (?)", []qf.Enrollment_UserStatus{
			qf.Enrollment_STUDENT,
			qf.Enrollment_TEACHER,
		}).Updates(&qf.Enrollment{GroupID: group.GetID()})
	if query.Error != nil {
		tx.Rollback()
		return query.Error
	}
	if query.RowsAffected != int64(len(userids)) {
		tx.Rollback()
		return ErrUpdateGroup
	}

	// Synchronize grades for all group submissions to match current group membership.
	if err := syncGroupGrades(tx, group.GetID(), userids); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// UpdateGroupStatus updates status field of a group.
func (db *GormDB) UpdateGroupStatus(group *qf.Group) error {
	return db.conn.Model(group).Update("status", group.GetStatus()).Error
}

// DeleteGroup deletes a group and its corresponding enrollments.
func (db *GormDB) DeleteGroup(groupID uint64) error {
	group, err := db.GetGroup(groupID)
	if err != nil {
		return err
	}

	tx := db.conn.Begin()
	if err := tx.Delete(group).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("UPDATE enrollments SET group_id= ? WHERE group_id= ?", 0, groupID).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetGroup returns the group with the specified group id.
func (db *GormDB) GetGroup(groupID uint64) (*qf.Group, error) {
	var group qf.Group
	userIDs := make([]uint64, 0)
	if err := db.conn.
		Model(&qf.Enrollment{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	if err := db.conn.
		Preload("Enrollments").
		Preload("Enrollments.UsedSlipDays").
		Preload("Enrollments.User").
		Preload("Users", "id IN ?", userIDs).
		First(&group, groupID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get group with ID %d: %w", groupID, err)
	}
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("no users found for group with ID %d", groupID)
	}
	return &group, nil
}

// GetGroupsByCourse returns the groups for the given course.
func (db *GormDB) GetGroupsByCourse(courseID uint64, statuses ...qf.Group_GroupStatus) ([]*qf.Group, error) {
	if len(statuses) == 0 {
		statuses = []qf.Group_GroupStatus{
			qf.Group_PENDING,
			qf.Group_APPROVED,
		}
	}
	var groups []*qf.Group
	if err := db.conn.
		Preload("Enrollments").
		Preload("Enrollments.UsedSlipDays").
		Preload("Enrollments.User").
		Preload("Users").
		Where(&qf.Group{CourseID: courseID}).
		Where("status in (?)", statuses).
		Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// syncGroupGrades synchronizes grade records for all group submissions to match current group membership.
// Creates new grades for newly added members and removes grades for users no longer in the group.
// Existing grades are preserved with their current status (e.g., APPROVED).
func syncGroupGrades(tx *gorm.DB, groupID uint64, userIDs []uint64) error {
	// Get all submissions for this group
	var submissions []*qf.Submission
	if err := tx.Model(&qf.Submission{}).
		Preload("Grades").
		Where(&qf.Submission{GroupID: groupID}).
		Find(&submissions).Error; err != nil {
		return err
	}

	for _, submission := range submissions {
		// Find which users already have grades
		existingGrades := make(map[uint64]*qf.Grade)
		for _, grade := range submission.GetGrades() {
			existingGrades[grade.GetUserID()] = grade
		}

		// Create grades for all current users, preserving existing ones
		// This will also remove grades for users no longer in the group
		submission.Grades = make([]*qf.Grade, len(userIDs))
		for i, userID := range userIDs {
			if existing, found := existingGrades[userID]; found {
				submission.Grades[i] = existing // Preserve existing grade
			} else {
				submission.Grades[i] = &qf.Grade{UserID: userID, SubmissionID: submission.GetID()} // New grade
			}
		}

		// Get assignment to set grade status
		var assignment qf.Assignment
		if err := tx.First(&assignment, submission.GetAssignmentID()).Error; err != nil {
			return err
		}
		// Set grades based on submission score
		submission.SetGradesIfApproved(&assignment, submission.GetScore())

		if err := tx.Model(&qf.Submission{
			ID: submission.GetID(),
		}).Association("Grades").Clear(); err != nil {
			return err
		}
		if err := tx.Model(submission).Association("Grades").Append(submission.Grades); err != nil {
			return err
		}
	}

	return nil
}
