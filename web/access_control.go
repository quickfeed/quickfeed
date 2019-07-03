package web

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/google/go-cmp/cmp"
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

// isTeacher returns true only if the current user is teacher for the given course.
func (s *AutograderService) isTeacher(userID uint64, courseID uint64) bool {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	return enrollment.Status == pb.Enrollment_TEACHER
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

// hasGroupAccess returns true if the current user is administrator,
// teacher of the given course, or one of the provided users.
func (s *AutograderService) hasGroupAccess(ctx context.Context, courseID, userID, groupID uint64) bool {
	currentUser, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	if currentUser.IsAdmin {
		return true
	}

	enrollment, err := s.db.GetEnrollmentByCourseAndUser(courseID, currentUser.ID)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	if enrollment.Status == pb.Enrollment_TEACHER || enrollment.Status == pb.Enrollment_STUDENT {
		return true
	}

	group, err := s.db.GetGroup(groupID)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	for _, u := range group.Users {
		if currentUser.ID == u.ID {
			return true
		}
	}
	return false
}

// hasAccessG returns true if the current user is administrator or one of the provided users.
func (s *AutograderService) hasAccessG(ctx context.Context, users []*pb.User) bool {
	currentUser, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return false
	}
	if currentUser.IsAdmin {
		return true
	}
	for _, u := range users {
		if currentUser.ID == u.ID {
			return true
		}
	}
	return false
}

// getUserAndSCM returns the current user and scm for the given provider.
// All errors are logged, but only a single error is returned to the client.
// This is a helper method to facilitate consistent treatment of errors and logging.
func (s *AutograderService) getUserAndSCM(ctx context.Context, provider string, mustBeAdmin bool) (*pb.User, scm.SCM, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, nil, status.Errorf(codes.NotFound, "failed to get current user")
	}
	if mustBeAdmin && !usr.IsAdmin {
		s.logger.Error("user must be admin to create or update")
		return nil, nil, status.Error(codes.PermissionDenied, "user must be admin to create or update")
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
func (s *AutograderService) getUserAndSCM2(ctx context.Context, courseID uint64, mustBeAdmin bool) (*pb.User, scm.SCM, error) {
	crs, err := s.getCourse(courseID)
	if err != nil {
		s.logger.Error(err)
		return nil, nil, status.Errorf(codes.NotFound, "failed to get course")
	}
	return s.getUserAndSCM(ctx, crs.GetProvider(), mustBeAdmin)
}

// teacherScopes defines scopes that must be enabled for a teacher token to be valid.
var teacherScopes = []string{"admin:org", "delete_repo", "repo", "user"}

// hasTeacherScopes checks whether current user has upgraded scopes on provided scm client.
func (s *AutograderService) hasTeacherScopes(ctx context.Context, sc scm.SCM) bool {
	auth := sc.GetUserScopes(ctx)
	sort.Strings(auth.Scopes)
	sort.Strings(teacherScopes)
	return cmp.Equal(auth.Scopes, teacherScopes)
}
