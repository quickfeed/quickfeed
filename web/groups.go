package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"github.com/gosimple/slug"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ErrGroupNameDuplicate indicates that another group with the same name already exists on this course
var (
	ErrGroupNameDuplicate = status.Errorf(codes.AlreadyExists, "group with this name already exists. Please choose another name")
	ErrUserNotInGroup     = status.Errorf(codes.NotFound, "user is not in group")
)

// getGroup returns the group for the given group ID.
func (s *AutograderService) getGroup(request *pb.GetGroupRequest) (*pb.Group, error) {
	group, err := s.db.GetGroup(request.GetGroupID())
	if err != nil {
		return nil, err
	}
	return group, nil
}

// getGroups returns all groups for the given course ID.
func (s *AutograderService) getGroups(request *pb.CourseRequest) (*pb.Groups, error) {
	groups, err := s.db.GetGroupsByCourse(request.GetCourseID())
	if err != nil {
		return nil, err
	}
	return &pb.Groups{Groups: groups}, nil
}

// getGroupByUserAndCourse returns the group of the given user and course.
func (s *AutograderService) getGroupByUserAndCourse(request *pb.GroupRequest) (*pb.Group, error) {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.CourseID, request.UserID)
	if err != nil {
		return nil, err
	}
	enrollment.SetSlipDays(enrollment.Course)
	grp, err := s.db.GetGroup(enrollment.GroupID)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = ErrUserNotInGroup
	}
	return grp, err
}

// DeleteGroup deletes group with the provided ID.
func (s *AutograderService) deleteGroup(ctx context.Context, sc scm.SCM, request *pb.GroupRequest) error {
	group, repos, _, err := s.getCourseGroupRepos(request)
	if err != nil {
		return err
	}

	// when deleting an approved group, remove github repository and team as well
	for _, repo := range repos {
		if err = s.db.DeleteRepositoryByRemoteID(repo.GetRepositoryID()); err != nil {
			// even if database record not found, still attempt to remove related github repo and team
			if err != gorm.ErrRecordNotFound {
				return err
			}
		}
		if err = deleteGroupRepoAndTeam(ctx, sc, repo.GetRepositoryID(), group.GetTeamID(), repo.GetOrganizationID()); err != nil {
			return err
		}
	}

	return s.db.DeleteGroup(request.GetGroupID())
}

// createGroup creates a new group for the given course and users.
// This function is typically called by a student when creating
// a group, which will later be (optionally) edited and approved
// by a teacher of the course using the updateGroup function below.
func (s *AutograderService) createGroup(request *pb.Group) (*pb.Group, error) {
	if !s.isValidGroupName(request.GetCourseID(), request.GetName()) {
		return nil, ErrGroupNameDuplicate
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
func (s *AutograderService) updateGroup(ctx context.Context, sc scm.SCM, request *pb.Group) error {
	group, repos, course, err := s.getCourseGroupRepos(&pb.GroupRequest{
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
	if group.Name != request.Name && group.Status == pb.Group_PENDING {
		// if the new name coincides with one of the existing groups,
		// fail and inform the user
		if !s.isValidGroupName(request.CourseID, request.Name) {
			return ErrGroupNameDuplicate
		}
		group.Name = request.Name
	}

	newGroup := &pb.Group{
		ID:          group.ID,
		Name:        group.Name,
		CourseID:    group.CourseID,
		TeamID:      group.TeamID,
		Status:      group.Status,
		Users:       users,
		Enrollments: group.Enrollments,
	}

	if len(repos) == 0 {
		if request.Name != "" && newGroup.TeamID < 1 {
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
		newGroup.TeamID = team.ID
		// when updating a group for an existing team, name changes are not allowed.
		// this to avoid a mismatch between database group name and SCM team name
		s.logger.Debugf("updateGroup: SCM team name: %s, requested group name: %s", team.Name, request.Name)
		if team.Name != request.Name {
			newGroup.Name = team.Name
		}
	}

	// if there are changes in group membership, update SCM team
	if !group.ContainsAll(newGroup) {
		if err := updateGroupTeam(ctx, sc, newGroup, course.GetOrganizationID()); err != nil {
			return err
		}
	}

	// approve and update the group in the database
	newGroup.Status = pb.Group_APPROVED
	return s.db.UpdateGroup(newGroup)
}

// getGroupUsers returns the users of the specified group request, and checks
// that the group's users are enrolled in the course,
// that the enrollment has been accepted, and
// that the group's users are not already enrolled in another group.
func (s *AutograderService) getGroupUsers(request *pb.Group) ([]*pb.User, error) {
	if len(request.Users) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no users in group")
	}
	var userIds []uint64
	for _, user := range request.Users {
		enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.CourseID, user.ID)
		switch {
		case err == gorm.ErrRecordNotFound:
			return nil, status.Errorf(codes.NotFound, "user not enrolled in this course")
		case err != nil:
			return nil, err
		case enrollment.GroupID > 0 && request.ID == 0:
			// new group check (request group ID should be 0)
			return nil, status.Errorf(codes.InvalidArgument, "user already enrolled in another group")
		case enrollment.GroupID > 0 && enrollment.GroupID != request.ID:
			// update group check (request group ID should be non-0)
			return nil, status.Errorf(codes.InvalidArgument, "user already enrolled in another group")
		case enrollment.Status < pb.Enrollment_STUDENT:
			return nil, status.Errorf(codes.InvalidArgument, "user not yet accepted for this course")
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

// isValidGroupName ensures that SCM team and repository names for the given group
// will not coincide with one of the existing approved groups
func (s *AutograderService) isValidGroupName(courseID uint64, groupName string) bool {
	courseGroups, _ := s.db.GetGroupsByCourse(courseID)
	for _, group := range courseGroups {
		if slug.Make(groupName) == slug.Make(group.GetName()) {
			s.logger.Errorf("Failed to create group %s, another group % already exists, both names will result in %s on GitHub", groupName, group.Name, slug.Make(groupName))
			return false
		}
	}
	return true
}

// getCourseGroupRepos returns the group, the group's repositories and the organization ID
// for the given course and group specified in the GroupRequest.
func (s *AutograderService) getCourseGroupRepos(request *pb.GroupRequest) (*pb.Group, []*pb.Repository, *pb.Course, error) {
	group, err := s.db.GetGroup(request.GetGroupID())
	if err != nil {
		return nil, nil, nil, err
	}
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return nil, nil, nil, err
	}

	groupRepoQuery := &pb.Repository{
		OrganizationID: course.OrganizationID,
		GroupID:        group.GetID(),
		RepoType:       pb.Repository_GROUP,
	}
	repos, err := s.db.GetRepositories(groupRepoQuery)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, nil, err
	}
	return group, repos, course, nil
}
