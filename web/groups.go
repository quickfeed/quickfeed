package web

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"gorm.io/gorm"
)

const maxGroupNameLength = 20

var (
	ErrGroupNameEmpty     = connect.NewError(connect.CodeInvalidArgument, errors.New("group name is empty"))
	ErrGroupNameDuplicate = connect.NewError(connect.CodeAlreadyExists, errors.New("group name already in use"))
	ErrGroupNameTooLong   = connect.NewError(connect.CodeInvalidArgument, errors.New("group name is too long"))
	ErrGroupNameInvalid   = connect.NewError(connect.CodeInvalidArgument, errors.New("group name contains invalid characters"))
	ErrUserNotInGroup     = connect.NewError(connect.CodeNotFound, errors.New("user is not in group"))
)

// getGroupByUserAndCourse returns the group of the given user and course.
func (s *QuickFeedService) getGroupByUserAndCourse(request *qf.GroupRequest) (*qf.Group, error) {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.GetCourseID(), request.GetUserID())
	if err != nil {
		return nil, err
	}
	grp, err := s.db.GetGroup(enrollment.GetGroupID())
	if err != nil && err == gorm.ErrRecordNotFound {
		err = ErrUserNotInGroup
	}
	return grp, err
}

// DeleteGroup deletes group with the provided ID.
func (s *QuickFeedService) internalDeleteGroup(ctx context.Context, sc scm.SCM, request *qf.GroupRequest) error {
	course, group, err := s.getCourseGroup(request)
	if err != nil {
		return err
	}
	if err := s.db.DeleteGroup(request.GetGroupID()); err != nil {
		s.logger.Debugf("Failed to delete %s group %q from database: %v", course.GetCode(), group.GetName(), err)
		// continue with other delete operations
	}
	repo, err := s.getRepo(course, group.GetID(), qf.Repository_GROUP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for group %q: %w", course.GetCode(), group.GetName(), err)
	}
	if repo == nil {
		s.logger.Debugf("No %s repository found for group %q: %v", course.GetCode(), group.GetName(), err)
		// cannot continue without repository information
		return nil
	}

	// when deleting an approved group, remove github repository as well
	if err = s.db.DeleteRepository(repo.GetScmRepositoryID()); err != nil {
		s.logger.Debugf("Failed to delete %s repository for %q from database: %v", course.GetCode(), group.GetName(), err)
		// continue with other delete operations
	}
	opt := &scm.RepositoryOptions{
		ID: repo.GetScmRepositoryID(),
	}
	return sc.DeleteGroup(ctx, opt.ID)
}

// updateGroup updates the group for the given group request.
// Only teachers can invoke this, and allows the teacher to add or remove
// members from a group, before a repository is created on the SCM and
// the member details are updated in the database.
func (s *QuickFeedService) internalUpdateGroup(ctx context.Context, sc scm.SCM, request *qf.Group) error {
	course, group, err := s.getCourseGroup(&qf.GroupRequest{
		CourseID: request.GetCourseID(),
		GroupID:  request.GetID(),
	})
	if err != nil {
		return err
	}

	// get users of group, check consistency of group request
	users, err := s.getGroupUsers(request)
	if err != nil {
		return err
	}

	// allow changing the name of the group only if the group
	// is not already approved and the new name is valid
	newGroup, err := s.newGroup(group, request, users)
	if err != nil {
		return err
	}

	repo, err := s.getRepo(course, group.GetID(), qf.Repository_GROUP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for group %q: %w", course.GetCode(), group.GetName(), err)
	}
	if repo == nil {
		repo, err := createRepo(ctx, sc, course, newGroup)
		if err != nil {
			return err
		}
		s.logger.Debugf("Created group repo on SCM: %+v", repo)
		if err := s.db.CreateRepository(repo); err != nil {
			return err
		}
		s.logger.Debugf("Created group repo in database: %+v", repo)
	}

	// if there are changes in group membership, update group repository
	if !group.ContainsAll(newGroup) {
		if err := updateGroupMembers(ctx, sc, newGroup, course.GetScmOrganizationName()); err != nil {
			return err
		}
	}

	// approve and update the group in the database
	newGroup.Status = qf.Group_APPROVED
	return s.db.UpdateGroup(newGroup)
}

// newGroup returns a new group based on the request and the existing group.
func (s *QuickFeedService) newGroup(group, request *qf.Group, users []*qf.User) (*qf.Group, error) {
	if group.GetName() != request.GetName() && group.GetStatus() == qf.Group_PENDING {
		if err := s.checkGroupName(request.GetCourseID(), request.GetName()); err != nil {
			return nil, err // group name is invalid
		}
		group.Name = request.GetName()
	}
	return &qf.Group{
		ID:          group.GetID(),
		Name:        group.GetName(),
		CourseID:    group.GetCourseID(),
		Status:      group.GetStatus(),
		Users:       users,
		Enrollments: group.GetEnrollments(),
	}, nil
}

// getGroupUsers returns the users of the specified group request, and checks
// that the group's users are enrolled in the course,
// that the enrollment has been accepted, and
// that the group's users are not already enrolled in another group.
func (s *QuickFeedService) getGroupUsers(request *qf.Group) ([]*qf.User, error) {
	if len(request.GetUsers()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("no users in group"))
	}
	var userIds []uint64
	for _, user := range request.GetUsers() {
		enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.GetCourseID(), user.GetID())
		switch {
		case err == gorm.ErrRecordNotFound:
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not enrolled in this course"))
		case err != nil:
			return nil, err
		case enrollment.GetGroupID() > 0 && request.GetID() == 0:
			// new group check (request group ID should be 0)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user already enrolled in another group"))
		case enrollment.GetGroupID() > 0 && enrollment.GetGroupID() != request.GetID():
			// update group check (request group ID should be non-0)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user already enrolled in another group"))
		case enrollment.Status < qf.Enrollment_STUDENT:
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user not yet accepted for this course"))
		}
		userIds = append(userIds, user.GetID())
	}

	users, err := s.db.GetUsers(userIds...)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get users"))
	}
	if len(request.GetUsers()) != len(users) || len(users) != len(userIds) {
		return nil, fmt.Errorf("invariant violation (request.Users=%d, users=%d, userIds=%d)",
			len(request.GetUsers()), len(users), len(userIds))
	}
	return users, nil
}

// only allow letters, numbers, dash and underscore.
var regexpNonAuthorizedChars = regexp.MustCompile("[^a-zA-Z0-9-_]")

// checkGroupName returns an error if the group name is invalid; otherwise nil is returned.
func (s *QuickFeedService) checkGroupName(courseID uint64, groupName string) error {
	if groupName == "" {
		return ErrGroupNameEmpty
	}
	if len(groupName) > maxGroupNameLength {
		return ErrGroupNameTooLong
	}
	if regexpNonAuthorizedChars.MatchString(groupName) {
		return ErrGroupNameInvalid
	}
	courseGroups, err := s.db.GetGroupsByCourse(courseID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to get groups"))
	}
	for _, group := range courseGroups {
		if strings.EqualFold(group.GetName(), groupName) {
			return ErrGroupNameDuplicate
		}
	}
	return nil
}

// getCourseGroup returns the course and group specified in the GroupRequest.
func (s *QuickFeedService) getCourseGroup(request *qf.GroupRequest) (*qf.Course, *qf.Group, error) {
	group, err := s.db.GetGroup(request.GetGroupID())
	if err != nil {
		return nil, nil, err
	}
	course, err := s.db.GetCourse(request.GetCourseID())
	if err != nil {
		return nil, nil, err
	}
	return course, group, nil
}
