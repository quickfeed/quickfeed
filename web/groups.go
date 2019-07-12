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
func (s *AutograderService) getGroup(request *pb.RecordRequest) (*pb.Group, error) {
	return s.db.GetGroup(request.ID)
}

// getGroups returns all groups for the given course ID.
func (s *AutograderService) getGroups(request *pb.RecordRequest) (*pb.Groups, error) {
	groups, err := s.db.GetGroupsByCourse(request.ID)
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

// DeleteGroup deletes a pending or rejected group for the given gid.
func (s *AutograderService) deleteGroup(request *pb.RecordRequest) error {
	group, err := s.db.GetGroup(request.ID)
	if err != nil {
		return err
	}
	if group.Status > pb.Group_REJECTED {
		return status.Errorf(codes.Aborted, "accepted group cannot be deleted")
	}
	return s.db.DeleteGroup(request.ID)
}

// createGroup creates a new group for the given course and users.
// This function is typically called by a student when creating
// a group, which will later be (optionally) edited and approved
// by a teacher of the course using the updateGroup function below.
func (s *AutograderService) createGroup(request *pb.Group, currentUser *pb.User) (*pb.Group, error) {
	if _, err := s.db.GetCourse(request.CourseID); err != nil {
		return nil, status.Errorf(codes.NotFound, "course not found")
	}
	// get users of group, check consistency of group request
	if _, err := s.getGroupUsers(request, currentUser); err != nil {
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
func (s *AutograderService) updateGroup(ctx context.Context, request *pb.Group, currentUser *pb.User, sc scm.SCM) error {
	// course must exist in the database
	course, err := s.db.GetCourse(request.CourseID)
	if err != nil {
		return status.Errorf(codes.NotFound, "course not found")
	}
	// group must exist in the database
	group, err := s.db.GetGroup(request.ID)
	if err != nil {
		return status.Errorf(codes.NotFound, "group not found")
	}

	if request.Status == pb.Group_REJECTED || request.Status == pb.Group_DELETED {
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
	// will create group repository and set group status to approved

	// get users of group, check consistency of group request
	users, err := s.getGroupUsers(request, currentUser)
	if err != nil {
		return err
	}
	s.logger.Info("getGroupUsers got list of users: ", users)

	// check whether the group repo already exists
	groupRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		GroupID:        request.GetID(),
	}
	repos, err := s.db.GetRepositories(groupRepoQuery)
	if err != nil {
		return err
	}
	// request.Users from frontend may only have IDs
	// we need to get full user information from database
	request.Users = users

	if len(repos) == 0 {
		// found no repos for the group; create group repo and team
		repo, team, err := createGroupRepoAndTeam(ctx, sc, course, request)
		if err != nil {
			return err
		}
		// create database entry for group repository
		groupRepo := &pb.Repository{
			OrganizationID: course.OrganizationID,
			RepositoryID:   repo.ID,
			UserID:         0,
			GroupID:        request.ID,
			HTMLURL:        repo.WebURL,
			RepoType:       pb.Repository_USER, // TODO(meling) should we distinguish GroupRepo?
		}
		if err := s.db.CreateRepository(groupRepo); err != nil {
			return err
		}
		request.TeamID = team.ID
	} else {
		// github team already exists, update its members
		// use the group's existing team ID obtained from the database above.
		request.TeamID = group.TeamID
		if err := updateGroupTeam(ctx, sc, course, request); err != nil {
			return err
		}
	}

	// approve the updated group
	return s.db.UpdateGroup(&pb.Group{
		ID:       request.ID,
		Name:     request.Name,
		CourseID: request.CourseID,
		TeamID:   request.TeamID,
		Users:    users,
		Status:   pb.Group_APPROVED,
	})
}

// getGroupUsers returns the users of the specified group request, and checks
// that the current signed in user is part of the group,
// that the group's users are enrolled in the course,
// that the enrollment has been accepted, and
// that the group's users are not already enrolled in another group.
func (s *AutograderService) getGroupUsers(request *pb.Group, currentUser *pb.User) ([]*pb.User, error) {
	if len(request.Users) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no users in group")
	}
	var userIds []uint64
	currUserInGroup := false
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
		case enrollment.Status == pb.Enrollment_TEACHER && !s.isTeacher(currentUser.ID, request.CourseID):
			return nil, status.Errorf(codes.InvalidArgument, "only teachers can create group with a teacher")
		case currentUser.ID == user.ID:
			currUserInGroup = true
		}
		userIds = append(userIds, user.ID)
	}
	// current user must be member of the group or teacher
	if !(currUserInGroup || s.isTeacher(currentUser.ID, request.CourseID)) {
		return nil, status.Errorf(codes.InvalidArgument, "signed in user not in group or is not teacher")
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

// createGroupRepoAndTeam creates the given group in course on the provided SCM.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createGroupRepoAndTeam(ctx context.Context, s scm.SCM, course *pb.Course, group *pb.Group) (*scm.Repository, *scm.Team, error) {
	org, err := s.GetOrganization(ctx, course.OrganizationID)
	if err != nil {
		return nil, nil, status.Errorf(codes.NotFound, "organization not found")
	}

	opt := &scm.CreateRepositoryOptions{
		Organization: org,
		Path:         group.Name,
		Private:      true,
	}
	return s.CreateRepoAndTeam(ctx, opt, group.Name, gitUserNames(group))
}

func updateGroupTeam(ctx context.Context, s scm.SCM, course *pb.Course, group *pb.Group) error {
	org, err := s.GetOrganization(ctx, course.OrganizationID)
	if err != nil {
		return status.Errorf(codes.NotFound, "organization not found")
	}

	opt := &scm.CreateTeamOptions{
		Organization: org,
		TeamName:     group.Name,
		TeamID:       group.TeamID,
		Users:        gitUserNames(group),
	}
	return s.UpdateTeamMembers(ctx, opt)
}

func gitUserNames(g *pb.Group) []string {
	var gitUserNames []string
	for _, user := range g.Users {
		gitUserNames = append(gitUserNames, user.GetLogin())
	}
	return gitUserNames
}
