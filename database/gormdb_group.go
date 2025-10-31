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

	return db.conn.Transaction(func(tx *gorm.DB) error {
		// Update Users association
		if err := tx.Model(group).Association("Users").Replace(group.GetUsers()); err != nil {
			return err
		}

		// Clear group_id for previous group members
		if err := tx.Model(group).Association("Enrollments").Clear(); err != nil {
			return err
		}

		var userIDs []uint64
		for _, u := range group.GetUsers() {
			userIDs = append(userIDs, u.GetID())
		}
		// Set group_id for current group members
		query := tx.Model(&qf.Enrollment{}).
			Where(&qf.Enrollment{CourseID: group.GetCourseID()}).
			Where("user_id IN (?)", userIDs).
			Updates(&qf.Enrollment{GroupID: group.GetID()})
		if query.Error != nil {
			return query.Error
		}

		if query.RowsAffected != int64(len(userIDs)) {
			return ErrUpdateGroup
		}
		return nil
	})
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

	return db.conn.Transaction(func(tx *gorm.DB) error {
		// Clear all associations before deleting the group
		if err := tx.Model(group).Association("Users").Clear(); err != nil {
			return err
		}

		if err := tx.Model(group).Association("Enrollments").Clear(); err != nil {
			return err
		}

		// Delete the group
		return tx.Delete(group).Error
	})
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
