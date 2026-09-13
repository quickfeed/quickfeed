package web

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

var (
	errDuplicateStudentID = connect.NewError(connect.CodeAlreadyExists, errors.New("a QuickFeed account with this student ID already exists"))
	errDuplicateEmail     = connect.NewError(connect.CodeAlreadyExists, errors.New("a QuickFeed account with this email already exists"))
)

// editUserProfile updates the user profile according to the user data in
// the request object. If curUser is admin, and the request may also
// promote the user to admin.
func (s *QuickFeedService) editUserProfile(ctx context.Context, curUser, request *qf.User) error {
	targetUser, err := s.db.GetUser(request.GetID())
	if err != nil {
		return err
	}

	prevStudentID, prevEmail := targetUser.GetStudentID(), targetUser.GetEmail()
	if name := strings.TrimSpace(request.GetName()); name != "" {
		targetUser.Name = name
	}
	if studentID := strings.TrimSpace(request.GetStudentID()); studentID != "" {
		targetUser.StudentID = studentID
	}
	if email := strings.TrimSpace(request.GetEmail()); email != "" {
		targetUser.Email = email
	}
	if request.GetAvatarURL() != "" {
		targetUser.AvatarURL = request.GetAvatarURL()
	}
	// Only the changed fields are checked for duplicates; an unrelated update, such as
	// a name change, must not be rejected because of accounts that already collided.
	studentID, email := targetUser.GetStudentID(), targetUser.GetEmail()
	if studentID == prevStudentID {
		studentID = ""
	}
	if strings.EqualFold(email, prevEmail) {
		email = ""
	}
	if err := s.checkDuplicateProfile(ctx, targetUser, studentID, email); err != nil {
		return err
	}

	// log every change to admin state
	if targetUser.GetIsAdmin() != request.GetIsAdmin() {
		qlog.FromContext(ctx).Debug("changing administrator status", label.User, curUser.GetLogin(), label.TargetUser, targetUser.GetLogin(), "is_admin", request.GetIsAdmin())
	}
	// current user must be admin to change admin status of another user
	// admin status of super admin (user with ID 1) cannot be changed
	if curUser.GetIsAdmin() && request.GetID() > 1 {
		targetUser.IsAdmin = request.GetIsAdmin()
	}
	return s.db.UpdateUser(targetUser)
}

// checkDuplicateProfile returns an error if a user other than the given user already
// has the given student ID or email. Without this check a student that signs in with
// two different SCM accounts ends up with two QuickFeed accounts for the same person,
// which breaks enrollment and grading for the course.
func (s *QuickFeedService) checkDuplicateProfile(ctx context.Context, user *qf.User, studentID, email string) error {
	users, err := s.db.GetUsersByStudentIDOrEmail(studentID, email)
	if err != nil {
		return err
	}
	for _, other := range users {
		if other.GetID() == user.GetID() {
			continue
		}
		duplicateErr := errDuplicateEmail
		if studentID != "" && strings.TrimSpace(other.GetStudentID()) == studentID {
			duplicateErr = errDuplicateStudentID
		}
		qlog.FromContext(ctx).Info("rejecting duplicate user profile",
			label.TargetUser, user.GetLogin(), "duplicate_user", other.GetLogin(), label.Error, duplicateErr)
		return duplicateErr
	}
	return nil
}

// updateUserFromSCM fetches the latest user info from the SCM and updates the local user
// record in the database. This should be used ahead of operations that require valid SCM
// user info, such as adding users to organizations or teams.
func (s *QuickFeedService) updateUserFromSCM(ctx context.Context, sc scm.SCM, user *qf.User) error {
	ghUser, err := sc.GetUserByID(ctx, user.GetScmRemoteID())
	if err != nil {
		return err
	}
	ghLogin := ghUser.GetLogin()
	if ghLogin != "" && ghLogin != user.GetLogin() {
		qlog.FromContext(ctx).Info("updating SCM login", label.TargetUserID, user.GetID(), "old_login", user.GetLogin(), "new_login", ghLogin)
		user.Login = ghLogin
	}
	ghAvatar := ghUser.GetAvatarURL()
	if ghAvatar != "" && ghAvatar != user.GetAvatarURL() {
		user.AvatarURL = ghAvatar
	}
	return s.db.UpdateUser(user)
}
