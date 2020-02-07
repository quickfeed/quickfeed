package database

import (
	"fmt"
	"strings"

	pb "github.com/autograde/aguis/ag"
	"github.com/jinzhu/gorm"
)

// CreateGroup creates a new group and assign users to newly created group.
func (db *GormDB) CreateGroup(group *pb.Group) error {
	if len(group.Users) == 0 {
		return ErrEmptyGroup
	}
	if group.CourseID == 0 {
		return gorm.ErrRecordNotFound
	}
	var course uint64
	if err := db.conn.Model(&pb.Course{}).
		Where(&pb.Course{ID: group.CourseID}).
		Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	tx := db.conn.Begin()
	if err := tx.Model(&pb.Group{}).Create(group).Error; err != nil {
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
	query := tx.Model(&pb.Enrollment{}).
		Where(&pb.Enrollment{CourseID: group.CourseID}).
		Where("user_id IN (?) AND status IN (?)", userids,
			[]pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_TEACHER}).
		Updates(&pb.Enrollment{GroupID: group.ID})
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
func (db *GormDB) UpdateGroup(group *pb.Group) error {
	if group.CourseID == 0 {
		return gorm.ErrRecordNotFound
	}
	var course uint64
	if err := db.conn.Model(&pb.Course{}).
		Where(&pb.Course{ID: group.CourseID}).
		Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}

	tx := db.conn.Begin()
	if err := tx.Model(&pb.Group{}).Updates(group).Error; err != nil {
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
	query := tx.Model(&pb.Enrollment{}).
		Where(&pb.Enrollment{CourseID: group.CourseID}).
		Where("user_id IN (?) AND status IN (?)", userids,
			[]pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_TEACHER}).
		Updates(&pb.Enrollment{GroupID: group.ID})
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
func (db *GormDB) UpdateGroupStatus(group *pb.Group) error {
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
func (db *GormDB) GetGroup(groupID uint64) (*pb.Group, error) {
	var group pb.Group
	if err := db.conn.Preload("Enrollments").First(&group, groupID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("error fetching group record for group with ID %d: %w", groupID, err)
	}
	var userIds []uint64
	for _, enrollment := range group.Enrollments {
		userIds = append(userIds, enrollment.UserID)
		u, err := db.GetUser(enrollment.UserID)
		if err != nil {
			return nil, err
		}
		enrollment.User = u
	}
	if len(userIds) > 0 {
		users, err := db.GetUsers(userIds...)
		if err != nil {
			return nil, err
		}
		group.Users = users
	}
	return &group, nil
}

// GetGroupsByCourse returns the groups for the given course.
func (db *GormDB) GetGroupsByCourse(courseID uint64) ([]*pb.Group, error) {
	var groups []*pb.Group
	if err := db.conn.
		Preload("Enrollments").
		Where(&pb.Group{CourseID: courseID}).
		Find(&groups).Error; err != nil {
		return nil, err
	}

	for _, group := range groups {
		var userIds []uint64
		for _, enrollment := range group.Enrollments {
			userIds = append(userIds, enrollment.UserID)
			u, err := db.GetUser(enrollment.UserID)
			if err != nil {
				return nil, err
			}
			enrollment.User = u
		}
		if len(userIds) > 0 {
			users, err := db.GetUsers(userIds...)
			if err != nil {
				return nil, err
			}
			group.Users = users
		}
	}
	return groups, nil
}
