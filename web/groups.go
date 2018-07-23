package web

import (
	"net/http"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// PatchGroup updates status of a group
func PatchGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("gid"))
		if err != nil {
			return err
		}
		oldgrp, err := db.GetGroup(id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		var ngrp UpdateGroupRequest
		if err := c.Bind(&ngrp); err != nil {
			return err
		}
		if ngrp.Status > models.Teacher {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		user := c.Get("user").(*models.User)
		// TODO: This check should be performed in AccessControl.
		if !user.IsAdmin {
			// Ony Admin i.e Teacher can update status of a group
			return c.NoContent(http.StatusForbidden)
		}

		users := oldgrp.Users

		var courseInfo *models.Course
		courseInfo, err = db.GetCourse(oldgrp.CourseID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "course not found")
			}
			return err
		}

		var userRemoteIdentity []*models.RemoteIdentity
		// TODO move this into the for loop above, modify db.GetUsers() to also retreive RemoteIdentity
		// so we can remove individual GetUser calls
		for _, user := range users {
			remoteIdentityUser, _ := db.GetUser(user.ID)
			if err != nil {
				return err
			}
			// TODO, figure out which remote identity to be used!
			userRemoteIdentity = append(userRemoteIdentity, remoteIdentityUser.RemoteIdentities[0])
		}

		provider := c.Get(courseInfo.Provider)
		var s scm.SCM
		if provider != nil {
			s = provider.(scm.SCM)
		} else {
			return nil // TODO decide how to handle empty provider.
		}

		// TODO move this functionality down into SCM?
		// Note: This Requires alot of calls to git.
		// Figure out all group members git-username
		var gitUserNames []string
		for _, identity := range userRemoteIdentity {
			gitName, err := s.GetUserNameByID(c.Request().Context(), identity.RemoteID)
			if err != nil {
				return err
			}
			gitUserNames = append(gitUserNames, gitName)
		}

		// Create and add repo to autograder group
		dir, err := s.GetDirectory(c.Request().Context(), courseInfo.DirectoryID)
		if err != nil {
			return err
		}
		repo, err := s.CreateRepository(c.Request().Context(), &scm.CreateRepositoryOptions{
			Directory: dir,
			Path:      oldgrp.Name,
		})
		if err != nil {
			return err
		}

		// Add repo to DB
		dbRepo := models.Repository{
			DirectoryID:  courseInfo.DirectoryID,
			RepositoryID: repo.ID,
			HTMLURL:      repo.WebURL,
			Type:         models.UserRepo,
			UserID:       0,
			GroupID:      oldgrp.ID,
		}
		if err := db.CreateRepository(&dbRepo); err != nil {
			return err
		}

		if err := db.UpdateGroupStatus(&models.Group{
			ID:     oldgrp.ID,
			Status: ngrp.Status,
		}); err != nil {
			return err
		}

		// Create git-team
		team, err := s.CreateTeam(c.Request().Context(), &scm.CreateTeamOptions{
			Directory: &scm.Directory{Path: dir.Path},
			TeamName:  oldgrp.Name,
			Users:     gitUserNames,
		})
		if err != nil {
			return err
		}
		// Adding Repo to git-team
		if err = s.AddTeamRepo(c.Request().Context(), &scm.AddTeamRepoOptions{
			TeamID: team.ID,
			Owner:  repo.Owner,
			Repo:   repo.Path,
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

// GetGroup returns a group
func GetGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return err
		}
		group, err := db.GetGroup(gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		return c.JSONPretty(http.StatusOK, group, "\t")
	}
}

// DeleteGroup deletes a pending or rejected group
func DeleteGroup(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		gid, err := parseUint(c.Param("gid"))
		if err != nil {
			return err
		}
		group, err := db.GetGroup(gid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "group not found")
			}
			return err
		}
		if group.Status > models.Rejected {
			return echo.NewHTTPError(http.StatusForbidden, "accepted group cannot be deleted")
		}
		if err := db.DeleteGroup(gid); err != nil {
			return nil
		}
		return c.NoContent(http.StatusOK)
	}
}
