package web

import (
	"context"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
)

func userID(ctx context.Context) uint64 {
	return ctx.Value(auth.ContextKeyUserID).(uint64)
}

// hasCourseAccess returns true if the given user has access to the given course,
// as defined by the check function.
func (s *QuickFeedService) hasCourseAccess(userID, courseID uint64, check func(*qf.Enrollment) bool) bool {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	s.logger.Debugf("(user=%d, course=%d) has enrollment status %+v", userID, courseID, enrollment.GetStatus())
	return check(enrollment)
}

// // IsValidSubmission returns true if submitting student has active course enrollment or
// // if submitting group belongs to the given course.
// func IsValidSubmission(db database.Database, courseID, submissionID uint64) bool {
// 	sbm, err := db.GetSubmission(&qf.Submission{ID: submissionID})
// 	if err != nil {
// 		return false
// 	}

// 	if sbm.GroupID > 0 {
// 		grp, err := db.GetGroup(sbm.GroupID)
// 		if err != nil || grp.GetCourseID() != courseID {
// 			return false
// 		}
// 		return true
// 	}

// 	enrol, err := db.GetEnrollmentByCourseAndUser(courseID, sbm.UserID)
// 	if err != nil || enrol.IsNone() || enrol.IsPending() {
// 		return false
// 	}
// 	return true
// }

// isTeacher returns true if the given user is teacher for the given course.
func (s *QuickFeedService) isTeacher(userID, courseID uint64) bool {
	return s.hasCourseAccess(userID, courseID, func(e *qf.Enrollment) bool {
		return e.Status == qf.Enrollment_TEACHER
	})
}

// isCourseCreator returns true if the given user is course creator for the given course.
func (s *QuickFeedService) isCourseCreator(courseID, userID uint64) bool {
	course, _ := s.db.GetCourse(courseID, false)
	return course.GetCourseCreatorID() == userID
}
