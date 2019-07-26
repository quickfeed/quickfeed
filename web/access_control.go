package web

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"google.golang.org/grpc/metadata"
)

func (s *AutograderService) getCurrentUser(ctx context.Context) (*pb.User, error) {
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

func (s *AutograderService) getSCM(ctx context.Context, user *pb.User, provider string) (scm.SCM, error) {
	providers, err := s.GetProviders(ctx, &pb.Void{})
	if err != nil {
		return nil, err
	}
	if !providers.IsValidProvider(provider) {
		return nil, fmt.Errorf("invalid provider(%s)", provider)
	}

	for _, remoteID := range user.RemoteIdentities {
		if remoteID.Provider == provider {
			scm, ok := s.scms.GetSCM(remoteID.GetAccessToken())
			if !ok {
				return nil, fmt.Errorf("invalid token for user(%d) provider(%s)", user.ID, provider)
			}
			return scm, nil
		}
	}
	return nil, errors.New("no SCM found")
}

// isAdmin returns true only if the current user is administrator.
// Any non-admin user in the context returns false.
func (s *AutograderService) isAdmin(ctx context.Context) bool {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	return usr.IsAdmin
}

// isTeacher returns true only if the given user is teacher for the given course.
func (s *AutograderService) isTeacher(userID, courseID uint64) bool {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	return enrollment.Status == pb.Enrollment_TEACHER
}

// isOwner returns true only if the given user IDs match and is enrolled as student for the given course.
func (s *AutograderService) isOwner(userID, userID2, courseID uint64) bool {
	return s.isStudent(userID, courseID) && userID == userID2
}

func (s *AutograderService) isEnrolled(userID, courseID uint64) bool {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	return enrollment.Status == pb.Enrollment_STUDENT || enrollment.Status == pb.Enrollment_TEACHER
}

// isEnrolled returns true if the given user is enrolled as student for the given course.
func (s *AutograderService) isStudent(userID, courseID uint64) bool {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	return enrollment.Status == pb.Enrollment_STUDENT
}

// isInGroup returns true if the current user is in the provided group.
func isInGroup(currentUser *pb.User, group *pb.Group) bool {
	for _, u := range group.GetUsers() {
		if currentUser.ID == u.ID {
			return true
		}
	}
	return false
}

// hasAccess returns true if the current user is administrator or the user with userID.
func (s *AutograderService) hasAccess(ctx context.Context, userID uint64) bool {
	currentUser, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	return currentUser.IsAdmin || currentUser.ID == userID
}

// getUserAndSCM returns the current user and scm for the given provider.
// All errors are logged, but only a single error is returned to the client.
// This is a helper method to facilitate consistent treatment of errors and logging.
func (s *AutograderService) getUserAndSCM(ctx context.Context, provider string) (*pb.User, scm.SCM, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	scm, err := s.getSCM(ctx, usr, provider)
	if err != nil {
		s.logger.Error(err)
		return nil, nil, status.Errorf(codes.NotFound, "failed to get SCM for user")
	}
	return usr, scm, nil
}

// getUserAndSCM2 returns the current user and scm for the given course.
// All errors are logged, but only a single error is returned to the client.
// This is a helper method to facilitate consistent treatment of errors and logging.
func (s *AutograderService) getUserAndSCM2(ctx context.Context, courseID uint64) (*pb.User, scm.SCM, error) {
	crs, err := s.getCourse(courseID)
	if err != nil {
		s.logger.Error(err)
		return nil, nil, status.Errorf(codes.NotFound, "failed to get course")
	}
	return s.getUserAndSCM(ctx, crs.GetProvider())
}

// teacherScopes defines scopes that must be enabled for a teacher token to be valid.
var teacherScopes = map[string]bool{
	"admin:org":   true,
	"delete_repo": true,
	"repo":        true,
	"user":        true,
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
