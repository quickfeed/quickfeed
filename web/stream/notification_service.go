package stream

import (
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
)

type NotificationService struct {
	*Service[uint64, qf.Notification]
	db database.Database
}

func NewNotificationService(db database.Database) *NotificationService {
	return &NotificationService{
		Service: NewService[uint64, qf.Notification](),
		db:      db,
	}
}

// SendToTeachers sends a notification to all teachers of the given course.
func (s *NotificationService) SendToTeachers(courseID uint64, data *qf.Notification) {
	teachers, err := s.db.GetCourseTeachers(&qf.Course{ID: courseID})
	if err != nil {
		return
	}
	for _, teacher := range teachers {
		s.SendTo(data, teacher.ID)
	}
}

// SendToCourseMembers sends a notification to all course members with the given statuses.
// If no status is given, the notification is sent to all course members.
func (s *NotificationService) SendToCourseMembers(courseID uint64, data *qf.Notification, statuses ...qf.Enrollment_UserStatus) {
	members, err := s.db.GetEnrollmentsByCourse(courseID, statuses...)
	if err != nil {
		return
	}
	for _, member := range members {
		s.SendTo(data, member.ID)
	}
}
