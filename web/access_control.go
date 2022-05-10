package web

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web/auth/interceptors"
)

// ErrInvalidUserInfo is returned to user if user information in context is invalid.
var ErrInvalidUserInfo = status.Errorf(codes.PermissionDenied, "authorization failed. please try to logout and sign in again")

func (s *AutograderService) getCurrentUser(ctx context.Context) (*pb.User, error) {
	// process user id from context
	token, err := interceptors.GetFromMetadata(ctx, "token", "")
	if err != nil {
		s.logger.Errorf("Failed to get current user: %s", err)
		return nil, errors.New("no user metadata in context")
	}
	claims, err := s.tokenManager.GetClaims(token)
	if err != nil {
		return nil, err
	}
	// return the user corresponding to userID, or an error.
	return s.db.GetUser(claims.UserID)
}

func (s *AutograderService) getSCM(courseID uint64) (scm.SCM, error) {
	sc, ok := s.scmMaker.GetSCM(courseID)
	if ok {
		return sc, nil
	}
	return nil, errors.New(fmt.Sprintf("no SCM found for course %d", courseID))
}

// hasCourseAccess returns true if the given user has access to the given course,
// as defined by the check function.
func (s *AutograderService) hasCourseAccess(userID, courseID uint64, check func(*pb.Enrollment) bool) bool {
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
func (s *AutograderService) isValidSubmissionRequest(submission *pb.SubmissionRequest) bool {
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
	return s.hasCourseAccess(submission.GetUserID(), submission.GetCourseID(), func(e *pb.Enrollment) bool {
		return e.Status >= pb.Enrollment_STUDENT
	})
}

// isValidSubmission returns true if submission belongs to active lab of the given course
// and submitted by valid student or group.
func (s *AutograderService) isValidSubmission(submissionID uint64) bool {
	submission, err := s.db.GetSubmission(&pb.Submission{ID: submissionID})
	if err != nil {
		return false
	}
	assignment, err := s.db.GetAssignment(&pb.Assignment{ID: submission.GetAssignmentID()})
	if err != nil {
		return false
	}
	request := &pb.SubmissionRequest{
		CourseID: assignment.GetCourseID(),
		UserID:   submission.GetUserID(),
		GroupID:  submission.GetGroupID(),
	}
	return s.isValidSubmissionRequest(request)
}

// isCourseCreator returns true if the given user is course creator for the given course.
func (s *AutograderService) isCourseCreator(courseID, userID uint64) bool {
	course, _ := s.db.GetCourse(&pb.Course{ID: courseID}, false)
	return course.GetCourseCreatorID() == userID
}

// TODO(vera): these two methods can be repurposed to return course-related scm client for most methods
// and personal access token based scm for user when accepting invitations
// getUserAndSCM returns the current user and scm for the given provider.
// All errors are logged, but only a single error is returned to the client.
// This is a helper method to facilitate consistent treatment of errors and logging.
func (s *AutograderService) getUserAndSCM(ctx context.Context, courseID uint64) (*pb.User, scm.SCM, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		return nil, nil, err
	}
	scm, err := s.getSCM(courseID)
	if err != nil {
		return nil, nil, err
	}
	return usr, scm, nil
}

// getUserAndSCMForCourse returns the current user and scm for the given course.
// All errors are logged, but only a single error is returned to the client.
// This is a helper method to facilitate consistent treatment of errors and logging.
func (s *AutograderService) getUserAndSCMForCourse(ctx context.Context, courseID uint64) (*pb.User, scm.SCM, error) {
	return s.getUserAndSCM(ctx, courseID)
}

// teacherScopes defines scopes that must be enabled for a teacher token to be valid.
var teacherScopes = map[string]bool{
	"admin:org":      true,
	"delete_repo":    true,
	"repo":           true,
	"user":           true,
	"admin:org_hook": true,
}

// hasTeacherScopes checks whether current user has upgraded scopes on provided scm client.
func hasTeacherScopes(ctx context.Context, sc scm.SCM) bool {
	authorization := sc.GetUserScopes(ctx)
	scopesFound := 0
	for _, scope := range authorization.Scopes {
		if teacherScopes[scope] {
			scopesFound++
		}
	}
	return scopesFound == len(teacherScopes)
}
