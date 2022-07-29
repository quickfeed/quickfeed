package web

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"google.golang.org/grpc/metadata"
)

// ErrInvalidUserInfo is returned to user if user information in context is invalid.
var ErrInvalidUserInfo = status.Errorf(codes.PermissionDenied, "authorization failed. please try to logout and sign in again")

func (s *QuickFeedService) getCurrentUser(ctx context.Context) (*qf.User, error) {
	// process user id from context
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("malformed request")
	}
	userValues := meta.Get("user")
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

func (s *QuickFeedService) getSCM(user *qf.User, provider string) (scm.SCM, error) {
	currentProvider := env.ScmProvider()
	if provider != currentProvider {
		return nil, fmt.Errorf("invalid provider (%s), active scm provider is %s", provider, currentProvider)
	}
	for _, remoteID := range user.RemoteIdentities {
		if remoteID.Provider == provider {
			// TODO(vera): must add methods on the SCMManger to get/create scms
			scm, ok := s.scms.Scms.GetSCM(remoteID.GetAccessToken())
			if !ok {
				return nil, fmt.Errorf("invalid token for user(%d) provider(%s)", user.ID, provider)
			}
			return scm, nil
		}
	}
	return nil, errors.New("no SCM found")
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

// getUserAndSCM returns the current user and scm for the given provider.
// All errors are logged, but only a single error is returned to the client.
// This is a helper method to facilitate consistent treatment of errors and logging.
func (s *QuickFeedService) getUserAndSCM(ctx context.Context, provider string) (*qf.User, scm.SCM, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		return nil, nil, err
	}
	scm, err := s.getSCM(usr, provider)
	if err != nil {
		return nil, nil, err
	}
	return usr, scm, nil
}

// getUserAndSCMForCourse returns the current user and scm for the given course.
// All errors are logged, but only a single error is returned to the client.
// This is a helper method to facilitate consistent treatment of errors and logging.
func (s *QuickFeedService) getUserAndSCMForCourse(ctx context.Context, courseID uint64) (*qf.User, scm.SCM, error) {
	crs, err := s.getCourse(courseID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get course with ID %d: %w", courseID, err)
	}
	return s.getUserAndSCM(ctx, crs.GetProvider())
}
