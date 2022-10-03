package web

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
)

var ErrMissingInstallation = connect.NewError(connect.CodePermissionDenied, errors.New("github application is not installed on the course organization"))

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

// isValidSubmission returns true if submitting student has active course enrollment or
// if submitting group belongs to the given course.
func (s *QuickFeedService) isValidSubmissionRequest(submission *qf.SubmissionRequest) bool {
	if !submission.IsValid() {
		return false
	}
	// ensure that group belongs to course
	if submission.GetGroupID() > 0 {
		group, err := s.db.GetGroup(submission.GetGroupID())
		if err != nil || group.GetCourseID() != submission.GetCourseID() {
			return false
		}
		return true
	}
	// ensure that student has active enrollment
	return s.hasCourseAccess(submission.GetUserID(), submission.GetCourseID(), func(e *qf.Enrollment) bool {
		return e.Status >= qf.Enrollment_STUDENT
	})
}

// isValidSubmission returns true if submission belongs to active lab of the given course
// and submitted by valid student or group.
func (s *QuickFeedService) isValidSubmission(submissionID uint64) bool {
	submission, err := s.db.GetSubmission(&qf.Submission{ID: submissionID})
	if err != nil {
		return false
	}
	assignment, err := s.db.GetAssignment(&qf.Assignment{ID: submission.GetAssignmentID()})
	if err != nil {
		return false
	}

	request := &qf.SubmissionRequest{
		CourseID: assignment.GetCourseID(),
		UserID:   submission.GetUserID(),
		GroupID:  submission.GetGroupID(),
	}
	return s.isValidSubmissionRequest(request)
}

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
