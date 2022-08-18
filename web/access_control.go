package web

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"google.golang.org/grpc/metadata"
)

// ErrInvalidUserInfo is returned to user if user information in context is invalid.
var (
	ErrInvalidUserInfo     = status.Error(codes.PermissionDenied, "authorization failed. please try to logout and sign in again")
	ErrMissingInstallation = status.Error(codes.PermissionDenied, "github application is not installed on the course organization")
)

func (s *QuickFeedService) getCurrentUser(ctx context.Context) (*qf.User, error) {
	// process user id from context
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("malformed request")
	}
	userValues := meta.Get(auth.UserKey)
	if len(userValues) == 0 {
		return nil, errors.New("no user metadata in context")
	}
	if len(userValues) != 1 || userValues[0] == "" {
		return nil, errors.New("invalid user payload in context")
	}
	userID, err := strconv.ParseUint(userValues[0], 10, 64)
	if err != nil {
		return nil, err
	}
	// return the user corresponding to userID, or an error.
	return s.db.GetUser(userID)
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

// isEnrolled returns true if the given user is enrolled in the given course.
func (s *QuickFeedService) isEnrolled(userID, courseID uint64) bool {
	return s.hasCourseAccess(userID, courseID, func(e *qf.Enrollment) bool {
		return e.Status == qf.Enrollment_STUDENT || e.Status == qf.Enrollment_TEACHER
	})
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
