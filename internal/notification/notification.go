package notification

import (
	"errors"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/stream"
)

var ErrSelfNotification = errors.New("you can't send notification to yourself")

type notificationService struct {
	db      database.Database
	streams *stream.Service[uint64, qf.Notification]
}

func New(db database.Database, streams *stream.Service[uint64, qf.Notification]) Notification {
	return &notificationService{
		db:      db,
		streams: streams,
	}
}

type Notification interface {
	Notify(notification *qf.Notification) error
	// Maitenance alerts user that the system will be down for maintenance.
	// BroadCasts to all users.
	Maintenance() error
	// DeadlineIsApproaching reminds user that the deadline is approaching.
	DeadlineIsApproaching(userIDs []uint64) error
	// AssignmentsRepoIsAhead notifies the user that the assignment repository is ahead of their repository for the given course.
	AssignmentsRepoIsAhead(userIDs []uint64, courseID uint64) error
	// GroupCreated notifies the teacher(s) in a course that a new group has been created.
	GroupCreated(userIDs []uint64, groupID uint64) error

	CustomNotification(userID []uint64, title, body, url string) error
}

// Notify validates, creates and streams the notification to the recipients.
func (n *notificationService) Notify(notification *qf.Notification) error {
	/*
		TODO: Uncomment this when the notification system is ready.

		if notification.IsSelfNotification() {
			return ErrSelfNotification
		}
	*/
	switch notification.RecipientType {
	case qf.Notification_ALL:
		users, err := n.db.GetUsers()
		if err != nil {
			return err
		}
		notification.AddRecipients(users)
		/*case qf.Notification_STUDENTS:
			students, err := n.db.GetEnrollmentsByCourse(notification.CourseID, qf.Enrollment_STUDENT)
			if err != nil {
				return err
			}
			students[0].GetUserID()
			notification.AddRecipients(students)
		case qf.Notification_TEACHERS:
			teachers, err := n.db.GetEnrollmentsByCourse(notification.CourseID, qf.Enrollment_TEACHER)
			if err != nil {
				return err
			}
			notification.AddRecipients(teachers)*/
	}
	if err := n.db.CreateNotification(notification); err != nil {
		return err
	}
	// Streams the notification to the recipients.
	n.streams.SendTo(notification, notification.GetReceivers()...)
	return nil
}

func (n *notificationService) CustomNotification(userIDs []uint64, title, body, url string) error {
	// This method is not implemented yet.
	// It should send a notification to the given user IDs with the given title, body, and URL.
	return nil
}
