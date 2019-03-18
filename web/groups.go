package web

import (
	"context"
	"net/http"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

// NewGroupRequest represents a new group.
type NewGroupRequest struct {
	Name     string   `json:"name"`
	CourseID uint64   `json:"courseid"`
	UserIDs  []uint64 `json:"userids"`
}

func (grp *NewGroupRequest) valid() bool {
	return grp != nil &&
		grp.Name != "" &&
		len(grp.UserIDs) > 0
}

// UpdateGroupRequest updates group
type UpdateGroupRequest struct {
	//TODO(meling) make separate GroupStatus iota for group (to make type safe use??)
	Status uint `json:"status"`
}

// GetGroups returns all groups for the given course cid.
func GetGroups(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		//TODO(meling) is this really necessary? The GetGroupsByCourse() should fail in this case too.
		if _, err := db.GetCourse(cid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}
		groups, err := db.GetGroupsByCourse(cid)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, groups, "\t")
	}
}

// NewGroup creates a new group for the given course cid.
// This function is typically called by a student when creating
// a group, which will later be (optionally) edited and approved
// by a teacher of the course using the UpdateGroup function below.
func NewGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if _, err := db.GetCourse(cid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}

		var grp NewGroupRequest
		if err := c.Bind(&grp); err != nil {
			return err
		}
		if !grp.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// don't add remote identities here since these users are returned to the client.
		users, err := db.GetUsers(false, grp.UserIDs...)
		if err != nil {
			return err
		}
		// sanity check: are provided user IDs actual users in database
		if len(users) != len(grp.UserIDs) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		// signed in student user must be member of the group
		signedInUser := c.Get("user").(*models.User)
		signedInUserEnrollment, err := db.GetEnrollmentByCourseAndUser(cid, signedInUser.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "unable to retrieve enrollment for signed in user")
		}
		signedInUserInGroup := false

		// only enrolled users can join a group
		// prevent group override if a student is already in a group in this course
		for _, user := range users {
			enrollment, err := db.GetEnrollmentByCourseAndUser(cid, user.ID)
			switch {
			case err == gorm.ErrRecordNotFound:
				return echo.NewHTTPError(http.StatusNotFound, "user not enrolled in course")
			case err != nil:
				return err
			case enrollment.GroupID > 0:
				return echo.NewHTTPError(http.StatusBadRequest, "user already enrolled in another group")
			case enrollment.Status < models.Student:
				return echo.NewHTTPError(http.StatusBadRequest, "user not yet accepted for this course")
			case enrollment.Status == models.Teacher && signedInUserEnrollment.Status != models.Teacher:
				return echo.NewHTTPError(http.StatusBadRequest, "only teachers can create group with a teacher")
			case signedInUser.ID == user.ID && enrollment.Status == models.Student:
				signedInUserInGroup = true
			}
		}

		// if signed in user is teacher we proceed to create group with the enrolled users
		if signedInUserEnrollment.Status == models.Teacher {
			signedInUserInGroup = true
		}
		if !signedInUserInGroup {
			return echo.NewHTTPError(http.StatusBadRequest, "student must be member of new group")
		}

		group := models.Group{
			Name:     grp.Name,
			CourseID: cid,
			Users:    users,
		}
		// create a new group and update group_id in enrollment table
		if err := db.CreateGroup(&group); err != nil {
			if err == database.ErrDuplicateGroup {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return err
		}

		return c.JSONPretty(http.StatusCreated, &group, "\t")
	}
}

// UpdateGroup updates the group for the given gid and course cid.
// Only teachers can invoke this, and allows the teacher to add or remove
// members from a group, before a repository is created on the SCM and
// the member details are updated in the database.
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
			case enrollment.GroupID > 0 && enrollment.GroupID != gid:
				return echo.NewHTTPError(http.StatusBadRequest, "user is already in another group")
			case enrollment.Status < models.Student:
				return echo.NewHTTPError(http.StatusBadRequest, "user is not yet accepted to this course")
			}
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
		logger.WithField("repo", repo).Println("Created new group repository in database")

		if err := db.UpdateGroupStatus(&models.Group{
			ID:     oldgrp.ID,
			Status: ngrp.Status,
		}); err != nil {
			logger.WithField("status", ngrp.Status).WithField("gid", oldgrp.ID).WithError(err).Warn("Failed to update group status in database")
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

// createGroupRepoAndTeam creates the given group in course on the provided SCM.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createGroupRepoAndTeam(c echo.Context, course *models.Course, group *models.Group) (*scm.Repository, error) {
	s, err := getSCM(c, course.Provider)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
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

func fetchGitUserNames(ctx context.Context, s scm.SCM, course *models.Course, users ...*models.User) ([]string, error) {
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

// GetGroup returns the group for the given gid.
func GetGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return err
		}
		group, err := db.GetGroup(false, gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		return c.JSONPretty(http.StatusOK, group, "\t")
	}
}

// DeleteGroup deletes a pending or rejected group for the given gid.
func DeleteGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return err
		}
		group, err := db.GetGroup(false, gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		//TODO(meling) make separate GroupStatus iota for group to avoid using models.Teacher.
		if group.Status > models.Rejected {
			return echo.NewHTTPError(http.StatusForbidden, "accepted group cannot be deleted")
		}
		if err := db.DeleteGroup(gid); err != nil {
			return nil
		}
		return c.NoContent(http.StatusOK)
	}
}
