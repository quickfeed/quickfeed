package web

import (
	"context"

	"github.com/autograde/aguis/scm"
	"github.com/labstack/echo"

	"google.golang.org/grpc/codes"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/status"
)

// GetGroup returns the group for the given gid.
func GetGroup(request *pb.RecordRequest, db database.Database) (*pb.Group, error) {
	group, err := db.GetGroup(request.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Group not found")
		}
		return nil, err
	}
	return group, nil
}

// GetGroupByUserAndCourse returns a single group of a user for a course
func GetGroupByUserAndCourse(request *pb.ActionRequest, db database.Database) (*pb.Group, error) {

	enrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseID, request.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not enrolled in course")
		}
		return nil, err

	}
	if enrollment.GroupID > 0 {
		group, err := db.GetGroup(enrollment.GroupID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "group not found")
		}
		return group, nil
	}
	return nil, status.Errorf(codes.NotFound, "group not found")
}

// GetGroups returns all groups for the given course cid.
func GetGroups(request *pb.RecordRequest, db database.Database) (*pb.Groups, error) {
	//TODO(Vera): add a corner case with non-existent course to the unit test
	groups, err := db.GetGroupsByCourse(request.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "group not found")
		}
		return nil, err
	}
	return &pb.Groups{Groups: groups}, nil
}

// DeleteGroup deletes a pending or rejected group for the given gid.
func DeleteGroup(request *pb.Group, db database.Database) error {
	group, err := db.GetGroup(request.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "group not found")
		}
		return err
	}
	if group.Status > pb.Group_Rejected {
		return status.Errorf(codes.Aborted, "accepted group cannot be deleted")
	}
	if err := db.DeleteGroup(request.ID); err != nil {
		return err
	}
	return nil
}

// NewGroup creates a new group for the given course cid.
// This function is typically called by a student when creating
// a group, which will later be (optionally) edited and approved
// by a teacher of the course using the UpdateGroup function below.
func NewGroup(request *pb.Group, db database.Database, currentUser *pb.User) (*pb.Group, error) {

	if _, err := db.GetCourse(request.CourseID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "course not found")
		}
		return nil, err
	}
	// make a sclice of IDs from the pb.User slice
	var userIds []uint64
	for _, user := range request.Users {
		userIds = append(userIds, user.ID)
	}
	// make sure that all users are in the database
	users, err := db.GetUsers(userIds...)
	if err != nil {
		return nil, err
	}

	if len(users) != len(request.Users) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload: users")
	}

	// signed in student must be member of the group
	signedInUserEnrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseID, currentUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "unable to retreive enrollment for signed in user")
	}
	signedInUserInGroup := false

	// only enrolled students can join a group
	// prevent group override if a student is already in a group in this course
	for _, user := range users {
		enrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseID, user.ID)
		switch {
		case err == gorm.ErrRecordNotFound:
			return nil, status.Errorf(codes.NotFound, "user not enrolled in this course")
		case err != nil:
			return nil, err
		case enrollment.GroupID > 0:
			return nil, status.Errorf(codes.InvalidArgument, "user already enrolled in another group")
		case enrollment.Status < pb.Enrollment_Student:
			return nil, status.Errorf(codes.InvalidArgument, "user not yet accepted for this course")
		case enrollment.Status == pb.Enrollment_Teacher && signedInUserEnrollment.Status != pb.Enrollment_Teacher:
			return nil, status.Errorf(codes.InvalidArgument, "only teachers can create group with a teacher")
		case currentUser.ID == user.ID && enrollment.Status == pb.Enrollment_Student:
			signedInUserInGroup = true
		}
	}

	// if signed in user is teacher we proceed to create group with the enrolled users
	if signedInUserEnrollment.Status == pb.Enrollment_Teacher {
		signedInUserInGroup = true
	}
	if !signedInUserInGroup {
		return nil, status.Errorf(codes.FailedPrecondition, "student must be member of new group")
	}

	// CreateGroup creates a new group and update groupid in enrollment table
	if err := db.CreateGroup(request); err != nil {
		if err == database.ErrDuplicateGroup {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	// database method returns error, front end method wants a new group to be returned. Get it from the database as an extra check for successful group creation
	newGroup, err := db.GetGroup(request.ID)
	if err != nil {
		return nil, err
	}
	return newGroup, nil
}

// UpdateGroup updates the group for the given gid and course cid.
// Only teachers can invoke this, and allows the teacher to add or remove
// members from a group, before a repository is created on the SCM and
// the member details are updated in the database.
func UpdateGroup(ctx context.Context, request *pb.Group, db database.Database, s scm.SCM, currentUser *pb.User) error {

	// only admin or course teacher are allowed to update groups
	signedInUserEnrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseID, currentUser.ID)
	if err != nil {
		return status.Errorf(codes.NotFound, "user not enrolled in the course")
	}
	if signedInUserEnrollment.Status != pb.Enrollment_Teacher && !currentUser.IsAdmin {
		return status.Errorf(codes.PermissionDenied, "only teacher or admin can update groups")
	}

	// course must exist in the database
	course, err := db.GetCourse(request.CourseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "course not found")
		}
		return err
	}

	// group must exist in the database
	_, err = db.GetGroup(request.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "group not found")
		}
		return status.Errorf(codes.NotFound, "no group in database")
	}

	// if the group is rejected or deleted, it is enough to update its entry in the database
	// if the group is being updated or approved, group repository will be created and group status set to approved
	if request.Status == pb.Group_Rejected || request.Status == pb.Group_Deleted {
		if err := db.UpdateGroupStatus(request); err != nil {
			return status.Errorf(codes.Aborted, "failed to update group status in database")
		}
		return nil
	}
	var userIds []uint64
	for _, user := range request.Users {
		userIds = append(userIds, user.ID)
	}
	users, err := db.GetUsers(userIds...)
	if err != nil {
		return status.Errorf(codes.Aborted, "invalid payload")
	}
	if len(request.Users) != len(users) || len(users) != len(userIds) {
		return status.Errorf(codes.Aborted, "invalid payload")
	}

	// check that each user is enrolled in the course, enrollment is accepted and user is not a member of another group
	for _, user := range request.Users {
		enrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseID, user.ID)
		switch {
		case err == gorm.ErrRecordNotFound:
			return status.Errorf(codes.NotFound, "user not enrolled in this course")
		case err != nil:
			return err
		case enrollment.GroupID > 0 && enrollment.GroupID != request.ID:
			return status.Errorf(codes.InvalidArgument, "user already in another group")
		case enrollment.Status < pb.Enrollment_Student:
			return status.Errorf(codes.InvalidArgument, "user not yet accepted to this course")
		}
	}

	// if all checks pass, create group repository
	repo, err := createGroupRepoAndTeam(ctx, s, course, request)
	if err != nil {
		return err
	}

	// create database entry for group repository
	groupRepo := &pb.Repository{
		DirectoryID:  course.DirectoryID,
		RepositoryID: repo.ID,
		UserID:       0,
		GroupID:      request.ID,
		HTMLURL:      repo.WebURL,
		RepoType:     pb.Repository_User, // TODO(meling) should we distinguish GroupRepo?
	}
	if err := db.CreateRepository(groupRepo); err != nil {
		return err
	}

	// update group
	if err := db.UpdateGroup(&pb.Group{
		ID:       request.ID,
		Name:     request.Name,
		CourseID: request.CourseID,
		Users:    users,
		Status:   pb.Group_Approved,
	}); err != nil {
		return err
	}

	return nil
}

// createGroupRepoAndTeam creates the given group in course on the provided SCM.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createGroupRepoAndTeam(ctx context.Context, s scm.SCM, course *pb.Course, group *pb.Group) (*scm.Repository, error) {

	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	dir, err := s.GetDirectory(ctx, course.DirectoryID)
	if err != nil {
		return nil, err
	}

	gitUserNames, err := fetchGitUserNames(ctx, s, course, group.Users...)
	if err != nil {
		return nil, err
	}

	opt := &scm.CreateRepositoryOptions{
		Directory: dir,
		Path:      group.Name,
		Private:   true,
	}
	return s.CreateRepoAndTeam(ctx, opt, group.Name, gitUserNames)
}

func fetchGitUserNames(ctx context.Context, s scm.SCM, course *pb.Course, users ...*pb.User) ([]string, error) {
	var gitUserNames []string
	for _, user := range users {
		remote := user.GetRemoteIDFor(course.Provider)
		if remote == nil {
			return nil, echo.ErrNotFound
		}
		// note this requires one git call per user in the group
		userName, err := s.GetUserNameByID(ctx, remote.RemoteID)
		if err != nil {
			return nil, err
		}
		gitUserNames = append(gitUserNames, userName)
	}
	return gitUserNames, nil
}

//TODO(vera): this can be removed after the new function is tested
/*
func UpdateGroup(logger logrus.FieldLogger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// signed in user must be teacher for course (don't need to be admin, I think?)
		signedInUser := c.Get("user").(*models.User)
		signedInUserEnrollment, err := db.GetEnrollmentByCourseAndUser(cid, signedInUser.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "user not enrolled in course")
		}
		if signedInUserEnrollment.Status != models.Teacher {
			return echo.NewHTTPError(http.StatusForbidden, "only teacher can update a group")
		}

		// the request is aimed to update the group with changes made by teacher
		// note: we reuse the NewGroupRequest also for updates
		var grp NewGroupRequest
		if err := c.Bind(&grp); err != nil {
			return err
		}
		if !grp.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// get users from database based on NewGroupRequst submitted by teacher
		grpUsers, err := db.GetUsers(false, grp.UserIDs...)
		if err != nil {
			return err
		}
		// update group in database according to request from teacher
		if err := db.UpdateGroup(&models.Group{
			ID:       gid,
			Name:     grp.Name,
			CourseID: cid,
			Users:    grpUsers,
		}); err != nil {
			if err == database.ErrDuplicateGroup {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return err
		}

		// we need the remote identities of users of the group to find their scm user names
		group, err := db.GetGroup(true, gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		gitUserNames = append(gitUserNames, gitName)
	}

	// Create and add repo to autograder group
	dir, err := s.GetDirectory(contextWithTimeout, course.DirectoryId)
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	repo, err := s.CreateRepository(contextWithTimeout, &scm.CreateRepositoryOptions{
		Directory: dir,
		Path:      request.Name,
		Private:   true,
	})
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

		// sanity check: after updating database with teacher approved group members,
		// the group should have same number of members as in the NewGroupRequest.
		if len(group.Users) != len(grp.UserIDs) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// only users enrolled in the course can join a group
		// prevent group override if a student is already in a group for this course
		for _, user := range group.Users {
			enrollment, err := db.GetEnrollmentByCourseAndUser(cid, user.ID)
			switch {
			case err == gorm.ErrRecordNotFound:
				return echo.NewHTTPError(http.StatusNotFound, "user is not enrolled to this course")
			case err != nil:
				return err
			case enrollment.GroupID > 0 && enrollment.GroupID != giD:
				return echo.NewHTTPError(http.StatusBadRequest, "user is already in another group")
			case enrollment.Status < models.Student:
				return echo.NewHTTPError(http.StatusBadRequest, "user is not yet accepted to this course")
			}
		}
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

		course, err := db.GetCourse(cid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}

		// create group repo and team on the SCM
		repo, err := createGroupRepoAndTeam(c, course, group)
		if err != nil {
			return err
		}

		groupRepo := &models.Repository{
			DirectoryID:  course.DirectoryID,
			RepositoryID: repo.ID,
			UserID:       0,
			GroupID:      group.ID,
			HTMLURL:      repo.WebURL,
			Type:         models.UserRepo, // TODO(meling) should we distinguish GroupRepo?
		}
		if err := db.CreateRepository(groupRepo); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

// PatchGroup updates status of the group for the given gid.
func PatchGroup(logger logrus.FieldLogger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("gid"))
		if err != nil {
			return err
		}
		var ngrp UpdateGroupRequest
		if err := c.Bind(&ngrp); err != nil {
			return err
		}
		//TODO(meling) make separate GroupStatus iota for group to avoid using models.Teacher.
		if ngrp.Status > models.Teacher {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		user := c.Get("user").(*models.User)
		// TODO: This check should be performed in AccessControl.
		if !user.IAdmin() {
			// Only admin / teacher can update status of a group
			return c.NoContent(http.StatusForbidden)
		}

		// we need the remote identities of the group's users
		oldgrp, err := db.GetGroup(true, id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		course, err := db.GetCourse(oldgrp.CourseID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}

		// create group repo and team on the SCM
		repo, err := createGroupRepoAndTeam(c, course, oldgrp)
		if err != nil {
			return err
		}
		logger.WithField("repo", repo).Println("Successfully created new group repository")

		// added the repository details to the database
		groupRepo := &models.Repository{
			DirectoryID:  course.DirectoryID,
			RepositoryID: repo.ID,
			UserID:       0,
			GroupID:      oldgrp.ID,
			HTMLURL:      repo.WebURL,
			Type:         models.UserRepo, // TODO(meling) should we distinguish GroupRepo?
		}
		if err := db.CreateRepository(groupRepo); err != nil {
			logger.WithField("url", repo.WebURL).WithField("gid", oldgrp.ID).WithError(err).Warn("Failed to create repository in database")
			return err
		}
		log.Println("Created new group repository in database")

		if err := db.UpdateGroupStatus(&pb.Group{
			ID:     oldGrp.ID,
			Status: request.Status,
		}); err != nil {
			log.Println("Failed to update group status in database")
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}

		return c.NoContent(http.StatusOK)
	}
}*/
