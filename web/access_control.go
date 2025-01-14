package web

import (
	"context"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
)

func userID(ctx context.Context) uint64 {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return 0
	}
	return claims.UserID
}

// isTeacher returns true if the user is a teacher in the given course.
func isTeacher(ctx context.Context, courseID uint64) bool {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.Courses[courseID] == qf.Enrollment_TEACHER
}

// courseStatus returns the user status in the given course.
func courseStatus(ctx context.Context, courseID uint64) qf.Enrollment_UserStatus {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return qf.Enrollment_NONE
	}
	return claims.Courses[courseID]
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

// isTeacher returns true if the given user is teacher for the given course.
func (s *QuickFeedService) isTeacher(userID, courseID uint64) bool {
	return s.hasCourseAccess(userID, courseID, func(e *qf.Enrollment) bool {
		return e.Status == qf.Enrollment_TEACHER
	})
}

// isCourseCreator returns true if the given user is course creator for the given course.
func (s *QuickFeedService) isCourseCreator(courseID, userID uint64) bool {
	course, _ := s.db.GetCourse(courseID)
	return course.GetCourseCreatorID() == userID
}
