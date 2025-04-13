package database

import (
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func (db *GormDB) CreateNotification(query *qf.Notification) error {
	query.CreatedAt = timestamppb.Now()
	return db.conn.Create(query).Error
}

func (db *GormDB) GetNotifications(userID uint64) ([]*qf.Notification, error) {
	var readNotificationsIDs []uint64
	if err := db.conn.Model(&qf.NotificationRecipient{}).
		Where("user_id = ? AND is_read", userID).
		Pluck("notification_id", &readNotificationsIDs).Error; err != nil {
		return nil, err
	}

	var notificationIDs []uint64
	if err := db.conn.Model(&qf.NotificationRecipient{}).
		Where("user_id = ?", userID).
		Pluck("notification_id", &notificationIDs).Error; err != nil {
		return nil, err
	}

	var notifications []*qf.Notification
	if err := db.conn.Model(&qf.Notification{}).Preload("Recipients").
		Where("id in (?)", notificationIDs).Find(&notifications).Error; err != nil {
		return nil, err
	}

	for _, notification := range notifications {
		for _, id := range readNotificationsIDs {
			if notification.GetID() == id {
				notification.IsRead = true
			}
		}
	}

	return notifications, nil
}

func (db *GormDB) UpdateNotification(query *qf.Notification) error {
	// full save associations is required to update the nested recipients
	return db.conn.Session(&gorm.Session{FullSaveAssociations: true}).Updates(query).Error
}
