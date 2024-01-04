package web

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"gorm.io/gorm"
)

const maxGroupNameLength = 20

var (
	ErrGroupNameDuplicate = connect.NewError(connect.CodeAlreadyExists, errors.New("group name already in use"))
	ErrGroupNameTooLong   = connect.NewError(connect.CodeInvalidArgument, errors.New("group name is too long"))
	ErrGroupNameInvalid   = connect.NewError(connect.CodeInvalidArgument, errors.New("group name contains invalid characters"))
	ErrUserNotInGroup     = connect.NewError(connect.CodeNotFound, errors.New("user is not in group"))
)

// getGroupByUserAndCourse returns the group of the given user and course.
func (s *QuickFeedService) getGroupByUserAndCourse(request *qf.GroupRequest) (*qf.Group, error) {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.CourseID, request.UserID)
	if err != nil {
		return nil, err
	}
	grp, err := s.db.GetGroup(enrollment.GroupID)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = ErrUserNotInGroup
	}
	return grp, err
}

// DeleteGroup deletes group with the provided ID.
func (s *QuickFeedService) deleteGroup(ctx context.Context, sc scm.SCM, request *qf.GroupRequest) error {
	course, group, err := s.getCourseGroup(request)
	if err != nil {
		return err
	}
	if err := s.db.DeleteGroup(request.GetGroupID()); err != nil {
		s.logger.Debugf("Failed to delete %s group %q from database: %v", course.Code, group.Name, err)
		// continue with other delete operations
	}
	repo, err := s.getRepo(course, group.GetID(), qf.Repository_GROUP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for group %q: %w", course.Code, group.Name, err)
	}
	if repo == nil {
		s.logger.Debugf("No %s repository found for group %q: %v", course.Code, group.Name, err)
		// cannot continue without repository information
		return nil
	}

	// when deleting an approved group, remove github repository and team as well
	if err = s.db.DeleteRepository(repo.GetScmRepositoryID()); err != nil {
		s.logger.Debugf("Failed to delete %s repository for %q from database: %v", course.Code, group.Name, err)
		// continue with other delete operations
	}
	opt := &scm.GroupOptions{
		OrganizationID: repo.GetScmOrganizationID(),
		RepositoryID:   repo.GetScmRepositoryID(),
		TeamID:         group.GetScmTeamID(),
	}
	return sc.DeleteGroup(ctx, opt)
}

// createGroup creates a new group for the given course and users.
// This function is typically called by a student when creating
// a group, which will later be (optionally) edited and approved
// by a teacher of the course using the updateGroup function below.
func (s *QuickFeedService) createGroup(request *qf.Group) (*qf.Group, error) {
	if err := s.checkGroupName(request.GetCourseID(), request.GetName()); err != nil {
		return nil, err
	}
	// get users of group, check consistency of group request
	if _, err := s.getGroupUsers(request); err != nil {
		s.logger.Errorf("CreateGroup: failed to retrieve users for group %s: %v", request.GetName(), err)
		return nil, err
	}
	// create new group and update groupID in enrollment table
	if err := s.db.CreateGroup(request); err != nil {
		return nil, err
	}
	return s.db.GetGroup(request.ID)
}

// updateGroup updates the group for the given group request.
// Only teachers can invoke this, and allows the teacher to add or remove
// members from a group, before a repository is created on the SCM and
// the member details are updated in the database.
func (s *QuickFeedService) updateGroup(ctx context.Context, sc scm.SCM, request *qf.Group) error {
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
	if group.Name != request.Name && group.Status == qf.Group_PENDING {
		// return error to user if group name is invalid
		if err := s.checkGroupName(request.GetCourseID(), request.GetName()); err != nil {
			return err
		}
		group.Name = request.Name
	}

	newGroup := &qf.Group{
		ID:          group.ID,
		Name:        group.Name,
		CourseID:    group.CourseID,
		ScmTeamID:   group.ScmTeamID,
		Status:      group.Status,
		Users:       users,
		Enrollments: group.Enrollments,
	}

	repo, err := s.getRepo(course, group.GetID(), qf.Repository_GROUP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for group %q: %w", course.Code, group.Name, err)
	}
	if repo == nil {
		// no group repository exists; create new repository for group
		if request.Name != "" && newGroup.ScmTeamID < 1 {
			// update group name only if team not already created on SCM
			newGroup.Name = request.Name
		}
		repo, team, err := createRepoAndTeam(ctx, sc, course, newGroup)
		if err != nil {
			return err
		}
		s.logger.Debugf("Creating group repo in the database: %+v", repo)
		if err := s.db.CreateRepository(repo); err != nil {
			return err
		}
		newGroup.ScmTeamID = team.ID
		// when updating a group for an existing team, name changes are not allowed.
		// this to avoid a mismatch between database group name and SCM team name
		s.logger.Debugf("updateGroup: SCM team name: %s, requested group name: %s", team.Name, request.Name)
		if team.Name != request.Name {
			newGroup.Name = team.Name
		}
	}

	// if there are changes in group membership, update SCM team
	if !group.ContainsAll(newGroup) {
		if err := updateGroupTeam(ctx, sc, newGroup, course.GetScmOrganizationID()); err != nil {
			return err
		}
	}

	// approve and update the group in the database
	newGroup.Status = qf.Group_APPROVED
	return s.db.UpdateGroup(newGroup)
}

// getGroupUsers returns the users of the specified group request, and checks
// that the group's users are enrolled in the course,
// that the enrollment has been accepted, and
// that the group's users are not already enrolled in another group.
func (s *QuickFeedService) getGroupUsers(request *qf.Group) ([]*qf.User, error) {
	if len(request.Users) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("no users in group"))
	}
	var userIds []uint64
	for _, user := range request.Users {
		enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.CourseID, user.ID)
		switch {
		case err == gorm.ErrRecordNotFound:
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not enrolled in this course"))
		case err != nil:
			return nil, err
		case enrollment.GroupID > 0 && request.ID == 0:
			// new group check (request group ID should be 0)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user already enrolled in another group"))
		case enrollment.GroupID > 0 && enrollment.GroupID != request.ID:
			// update group check (request group ID should be non-0)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user already enrolled in another group"))
		case enrollment.Status < qf.Enrollment_STUDENT:
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user not yet accepted for this course"))
		}
		userIds = append(userIds, user.ID)
	}

	users, err := s.db.GetUsers(userIds...)
	if err != nil {
		return nil, err
	}
	if len(request.Users) != len(users) || len(users) != len(userIds) {
		return nil, fmt.Errorf("invariant violation (request.Users=%d, users=%d, userIds=%d)",
			len(request.Users), len(users), len(userIds))
	}
	return users, nil
}

// only allow letters, numbers, dash and underscore.
var regexpNonAuthorizedChars = regexp.MustCompile("[^a-zA-Z0-9-_]")

// checkGroupName returns an error if the group name is invalid; otherwise nil is returned.
func (s *QuickFeedService) checkGroupName(courseID uint64, groupName string) error {
	if len(groupName) > maxGroupNameLength {
		return ErrGroupNameTooLong
	}
	if regexpNonAuthorizedChars.MatchString(groupName) {
		return ErrGroupNameInvalid
	}
	courseGroups, err := s.db.GetGroupsByCourse(courseID)
	if err != nil {
		return err
	}
	for _, group := range courseGroups {
		if group.GetName() == groupName {
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
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return nil, nil, err
	}
	return course, group, nil
}
