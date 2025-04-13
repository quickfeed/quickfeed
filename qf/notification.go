package qf

import "google.golang.org/protobuf/types/known/timestamppb"

// IsSelfNotification checks if the notification is configured to be sent to the creator.
func (n *Notification) IsSelfNotification() bool {
	for _, recipient := range n.Recipients {
		if recipient.UserID == n.Sender {
			return true
		}
	}
	return false
}

// GetReceivers returns a slice of user IDs of the recipients of the notification.
func (n *Notification) GetReceivers() []uint64 {
	var userIDs []uint64
	for _, recipient := range n.Recipients {
		userIDs = append(userIDs, recipient.UserID)
	}
	return userIDs
}

func (n *Notification) AddRecipients(users []*User) {
	for _, user := range users {
		n.Recipients = append(n.Recipients, &NotificationRecipient{
			UserID: user.GetID(),
		})
	}
}

// ReadBy marks the notification as read by the given user IDs.
// Should be used when users click or interact with the notification.
func (n *Notification) ReadBy(userIDs []uint64) {
	for _, userID := range userIDs {
		for _, recipient := range n.Recipients {
			if recipient.UserID == userID {
				recipient.Read = timestamppb.Now()
				recipient.IsRead = true
				break
			}
		}
	}
}
