package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// let the database return err if enrollment has no group
	return s.db.GetGroup(enrollment.GroupID)
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
		if err = s.db.DeleteRepository(repo.GetID()); err != nil {
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

	if request.Status == pb.Group_REJECTED {
		// if the group is rejected or deleted, it is enough to update its entry in the database.
		if err := s.db.UpdateGroupStatus(request); err != nil {
			return err
		}
		// if we delete a previously accepted group, reset the group's members enrollment status,
		// so that they can later join other groups.
		for _, member := range request.Users {
			if err = s.db.UpdateGroupEnrollment(member.ID, course.ID); err != nil {
				return err
			}
		}
		return nil
	}

	// the group is being updated or approved;
	// will create group repository and team and set group status to approved

	// get users of group, check consistency of group request
	users, err := s.getGroupUsers(request)
	if err != nil {
		return err
	}

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

	group.Status = pb.Group_APPROVED
	group.Users = users

	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.OrganizationID})
	if err != nil {
		return fmt.Errorf("updateGroup: organization not found: %w", err)
	}

	if len(repos) == 0 && group.GetTeamID() < 1 {
		// found no repos for the group; create group repo and team
		if request.GetName() != "" {
			group.Name = request.Name
		}

		repo, team, err := createRepoAndTeam(ctx, sc, org, group.Name, group.Name, group.UserNames())
		if err != nil {
			return err
		}
		// create database entry for group repository
		groupRepo := &pb.Repository{
			OrganizationID: course.OrganizationID,
			RepositoryID:   repo.ID,
			GroupID:        request.ID,
			HTMLURL:        repo.WebURL,
			RepoType:       pb.Repository_GROUP,
		}
		if err := s.db.CreateRepository(groupRepo); err != nil {
			return err
		}
		group.TeamID = team.ID
	} else {
		// github team already exists, update its members
		// use the group's existing team ID obtained from the database above.
		if err := updateGroupTeam(ctx, sc, org, group); err != nil {
			return err
		}
	}

	// approve the updated group
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
