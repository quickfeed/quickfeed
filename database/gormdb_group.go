package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

// CreateGroup creates a new group and assign users to newly created group.
func (db *GormDB) CreateGroup(group *qf.Group) error {
	if len(group.Users) == 0 {
		return ErrEmptyGroup
	}
	if group.CourseID == 0 {
		return gorm.ErrRecordNotFound
	}
	var course int64
	if err := db.conn.Model(&qf.Course{}).
		Where(&qf.Course{ID: group.CourseID}).
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
	for _, u := range group.Users {
		userids = append(userids, u.ID)
	}
	query := tx.Model(&qf.Enrollment{}).
		Where(&qf.Enrollment{CourseID: group.CourseID}).
		Where("user_id IN (?) AND status IN (?)", userids,
			[]qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_TEACHER}).
		Updates(&qf.Enrollment{GroupID: group.ID})
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
	if group.CourseID == 0 {
		return gorm.ErrRecordNotFound
	}
	var course int64
	if err := db.conn.Model(&qf.Course{}).
		Where(&qf.Course{ID: group.CourseID}).
		Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	tx := db.conn.Begin()
	if err := tx.Model(group).Updates(group).Error; err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateGroup
		}
		return err
	}
	if err := tx.Exec("UPDATE enrollments SET group_id= ? WHERE group_id= ?", 0, group.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	var userids []uint64
	for _, u := range group.Users {
		userids = append(userids, u.ID)
	}
	query := tx.Model(&qf.Enrollment{}).
		Where(&qf.Enrollment{CourseID: group.CourseID}).
		Where("user_id IN (?) AND status IN (?)", userids,
			[]qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_TEACHER}).
		Updates(&qf.Enrollment{GroupID: group.ID})
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

// UpdateGroupStatus updates status field of a group.
func (db *GormDB) UpdateGroupStatus(group *qf.Group) error {
	return db.conn.Model(group).Update("status", group.Status).Error
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
		return nil, fmt.Errorf("error fetching group record for group with ID %d: %w", groupID, err)
	}
	if len(userIDs) == 0 {
		return nil, errors.New("failed to get next student reviewer: no users in group")
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
