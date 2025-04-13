package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetNotifications(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	user := qtest.CreateFakeUser(t, db)
	notification := qtest.CreateNotification(t, db, admin.GetID(), []uint64{user.GetID()})

	want := []*qf.Notification{notification}
	got, err := db.GetNotifications(user.GetID())
	if err != nil {
		t.Fatalf("GetNotifications(%d) = %v", user.GetID(), err)
	}

	qtest.Diff(t, "GetNotifications() mismatch", got, want, protocmp.Transform())
}

// TestCreateNotification tests updating the notification as read
func TestUpdateNotification(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	user := qtest.CreateFakeUser(t, db)
	notification := qtest.CreateNotification(t, db, admin.GetID(), []uint64{user.GetID()})
	want := []*qf.Notification{notification}

	notification.ReadBy([]uint64{user.GetID()})

	if err := db.UpdateNotification(notification); err != nil {
		t.Fatalf("UpdateNotification() error = %v", err)
	}

	got, err := db.GetNotifications(user.GetID())
	if err != nil {
		t.Fatalf("GetNotifications(%d) = %v", user.GetID(), err)
	}

	qtest.Diff(t, "UpdateNotification() mismatch", got, want, protocmp.Transform())
}
