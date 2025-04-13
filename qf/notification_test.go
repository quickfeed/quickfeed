package qf_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

func TestIsSelfNotification(t *testing.T) {
	tests := []struct {
		name       string
		sender     uint64
		recipients []*qf.NotificationRecipient
		expected   bool
	}{
		{
			name:   "Self notification",
			sender: 1,
			recipients: []*qf.NotificationRecipient{
				{UserID: 1},
			},
			expected: true,
		},
		{
			name:   "Not self notification",
			sender: 1,
			recipients: []*qf.NotificationRecipient{
				{UserID: 2},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notification := &qf.Notification{
				Sender:     test.sender,
				Recipients: test.recipients,
			}
			result := notification.IsSelfNotification()
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestGetReceivers(t *testing.T) {
	notification := &qf.Notification{
		Recipients: []*qf.NotificationRecipient{
			{UserID: 1},
			{UserID: 2},
			{UserID: 3},
		},
	}

	receivers := notification.GetReceivers()

	if len(receivers) != 3 {
		t.Errorf("Expected 3 receivers, got %d", len(receivers))
	}
}

func TestReadBy(t *testing.T) {
	tests := []struct {
		name       string
		recipients []*qf.NotificationRecipient
		userIDs    []uint64
		expected   map[uint64]bool
	}{
		{
			name: "Mark single recipient as read",
			recipients: []*qf.NotificationRecipient{
				{UserID: 1},
				{UserID: 2},
			},
			userIDs: []uint64{1},
			expected: map[uint64]bool{
				1: true,
				2: false,
			},
		},
		{
			name: "Mark multiple recipients as read",
			recipients: []*qf.NotificationRecipient{
				{UserID: 1},
				{UserID: 2},
				{UserID: 3},
			},
			userIDs: []uint64{1, 3},
			expected: map[uint64]bool{
				1: true,
				2: false,
				3: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notification := &qf.Notification{
				Recipients: test.recipients,
			}
			notification.ReadBy(test.userIDs)
			for _, recipient := range notification.Recipients {
				if recipient.IsRead != test.expected[recipient.UserID] {
					t.Errorf("Expected recipient with UserID %d to be marked as %v, got %v", recipient.UserID, test.expected[recipient.UserID], recipient.IsRead)
				}
			}
		})
	}
}
