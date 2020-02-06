package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrGroupNameDuplicate indicates that another group with the same name already exists on this course
var (
	ErrGroupNameDuplicate = status.Errorf(codes.AlreadyExists, "group with this name already exists. Please choose another name")
	ErrUserNotInGroup     = status.Errorf(codes.NotFound, "user is not in group")
)

// getGroup returns the group for the given group ID.
func (s *AutograderService) getGroup(request *pb.GetGroupRequest) (*pb.Group, error) {
	return s.db.GetGroup(request.GetGroupID())
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
	grp, err := s.db.GetGroup(enrollment.GroupID)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = ErrUserNotInGroup
	}
	return grp, err
}

// DeleteGroup deletes group with the provided ID.
func (s *AutograderService) deleteGroup(ctx context.Context, sc scm.SCM, request *pb.GroupRequest) error {
	group, err := s.db.GetGroup(request.GetGroupID())
	if err != nil {
		return err
	}

	// get course organization ID
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return err
	}

	// get group repositories
	repos, err := s.db.GetRepositories(&pb.Repository{OrganizationID: course.GetOrganizationID(), GroupID: group.GetID(), RepoType: pb.Repository_GROUP})
	if err != nil && err != gorm.ErrRecordNotFound {
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

		if err = deleteGroupRepoAndTeam(ctx, sc, repo.GetRepositoryID(), group.GetTeamID()); err != nil {
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
	// check that there are no other groups with the same name
	// or a name that will result in the same github name
	groups, _ := s.db.GetGroupsByCourse(request.GetCourseID())
	for _, group := range groups {
		if slug.Make(request.GetName()) == slug.Make(group.GetName()) {
			s.logger.Errorf("failed to create group %s, another group % already exists, both names will result in %s on GitHub", request.Name, group.Name, slug.Make(request.Name))
			return nil, ErrGroupNameDuplicate
		}
	}

	// get users of group, check consistency of group request
	if _, err := s.getGroupUsers(request); err != nil {
		s.logger.Errorf("CreateGroup: failed to retrieve users for group %s: %s", request.GetName(), err)
		return nil, err
	}
	// create new group and update groupid in enrollment table
	if err := s.db.CreateGroup(request); err != nil {
		return nil, err
	}
	return s.db.GetGroup(request.ID)
}

// updateGroup updates the group for the given group request.
// Only teachers can invoke this, and allows the teacher to add or remove
// members from a group, before a repository is created on the SCM and
// the member details are updated in the database.
// TODO(meling) this function must be broken up and simplified
func (s *AutograderService) updateGroup(ctx context.Context, sc scm.SCM, request *pb.Group) error {
	// course must exist in the database
	course, err := s.db.GetCourse(request.CourseID, false)
	if err != nil {
		return err
	}
	// group must exist in the database
	group, err := s.db.GetGroup(request.ID)
	if err != nil {
		return err
	}

	// get users of group, check consistency of group request
	users, err := s.getGroupUsers(request)
	if err != nil {
		return err
	}
	request.Users = users

	// check whether the group repo already exists
	groupRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		GroupID:        request.GetID(),
		RepoType:       pb.Repository_GROUP,
	}
	repos, err := s.db.GetRepositories(groupRepoQuery)
	if err != nil {
		return err
	}

	if len(repos) == 0 {

		if request.Name != "" && group.TeamID < 1 {
			group.Name = request.Name
		}

		repo, team, err := createRepoAndTeam(ctx, sc, course.GetOrganizationID(), group.Name, request.UserNames())
		if err != nil {
			return err
		}
		repo.GroupID = group.GetID()
		s.logger.Debugf("Creating group repo in the database: %+v", repo)
		if err := s.db.CreateRepository(repo); err != nil {
			return err
		}
		group.TeamID = team.ID
		// if updating group with existing team, group name will not be changed
		// to avoid a mismatch between database and github names
		s.logger.Debugf("updateGroup: name of the new GitHub team: %s, requested group name: %s", team.Name, request.Name)
		if team.Name != request.Name {
			group.Name = team.Name
		}
	}

	// if there are changes in group members, update GitHub team
	if group.ContainsAll(request) {
		if err := updateGroupTeam(ctx, sc, group); err != nil {
			return err
		}
	}

	// approve and update the group in the database
	group.Status = pb.Group_APPROVED
	group.Users = users
	return s.db.UpdateGroup(group)
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
		// TODO(vera): it seems that the next check will also check that condition
		// they can probably be merged into one
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
