package database

import (
	"strings"

	pb "github.com/autograde/aguis/ag"
	"github.com/jinzhu/gorm"
)

// CreateGroup creates a new group and assign users to newly created group
func (db *GormDB) CreateGroup(group *pb.Group) error {
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

// UpdateGroup updates a group
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

// UpdateGroupStatus updates status field of a group
func (db *GormDB) UpdateGroupStatus(group *pb.Group) error {
	return db.conn.Model(group).Update("status", group.Status).Error
}

// DeleteGroup delete a group
func (db *GormDB) DeleteGroup(gid uint64) error {
	group, err := db.GetGroup(gid)
	if err != nil {
		return err
	}

	tx := db.conn.Begin()
	if err := tx.Delete(group).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("UPDATE enrollments SET group_id= ? WHERE group_id= ?", 0, gid).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetGroup returns a group specified by id return error if does not exits
func (db *GormDB) GetGroup(gid uint64) (*pb.Group, error) {
	var group pb.Group
	if err := db.conn.Preload("Enrollments").First(&group, gid).Error; err != nil {
		return nil, err
	}
	var userIds []uint64
	for _, enrollment := range group.Enrollments {
		userIds = append(userIds, enrollment.UserID)
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

// GetGroupsByCourse returns a list of groups
//TODO(meling) add test for this method
//TODO(meling) can this also Preload("Users") to avoid the GetUsers below.
func (db *GormDB) GetGroupsByCourse(cid uint64) ([]*pb.Group, error) {
	var groups []*pb.Group
	if err := db.conn.
		Preload("Enrollments").
		Where(&pb.Group{
			CourseID: cid,
		}).
		Find(&groups).Error; err != nil {
		return nil, err
	}

	for _, group := range groups {
		var userIds []uint64
		for _, enrollment := range group.Enrollments {
			userIds = append(userIds, enrollment.UserID)
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
