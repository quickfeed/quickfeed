package web

import (
	"fmt"
	"net/http"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// UpdateUserRequest updates a user object in the database.
type UpdateUserRequest struct {
	Name      string `json:"name"`
	StudentID string `json:"studentid"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarurl"`
	IsAdmin   *bool  `json:"isadmin"`
}

// GetSelf redirects to GetUser with the current user's id.
func GetSelf() echo.HandlerFunc {
	return func(c echo.Context) error {
		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		user := c.Get("user").(*models.User)
		return c.Redirect(http.StatusFound, fmt.Sprintf("/api/v1/users/%d", user.ID))
	}
}

// GetUser returns information about the provided user id.
func GetUser(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("uid"))
		if err != nil {
			return err
		}

		user, err := db.GetUser(id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		return c.JSONPretty(http.StatusFound, user, "\t")
	}
}

// GetUsers returns all the users in the database.
func GetUsers(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		users, err := db.GetUsers()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		return c.JSONPretty(http.StatusFound, users, "\t")
	}
}

// PatchUser promotes a user to an administrator
func PatchUser(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("uid"))
		if err != nil {
			return err
		}
		var uur UpdateUserRequest
		if err := c.Bind(&uur); err != nil {
			return err
		}

		status := http.StatusNotModified

		// Get user to update
		updateUser, err := db.GetUser(id)
		if err != nil {
			return err
		}

		// Get current user
		currentUser := c.Get("user").(*models.User)

		if uur.Name != "" {
			updateUser.Name = uur.Name
			status = http.StatusOK
		}
		if uur.StudentID != "" {
			updateUser.StudentID = uur.StudentID
			status = http.StatusOK
		}
		if uur.Email != "" {
			updateUser.Email = uur.Email
			status = http.StatusOK
		}
		if uur.AvatarURL != "" {
			updateUser.AvatarURL = uur.AvatarURL
			status = http.StatusOK
		}

		// Checking if user has admin privliges, and if user is trying to promote another user.
		if uur.IsAdmin != nil && *uur.IsAdmin != updateUser.IsAdmin && currentUser.IsAdmin {
			updateUser.IsAdmin = *uur.IsAdmin
			status = http.StatusOK
		} else if uur.IsAdmin != nil && *uur.IsAdmin != updateUser.IsAdmin && !currentUser.IsAdmin {
			status = http.StatusUnauthorized
		}

		if err := db.UpdateUser(updateUser); err != nil {
			return err
		}

		return c.NoContent(status)
	}
}

// GetGroupByUserAndCourse returns a single group of a user for a course
func GetGroupByUserAndCourse(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		uid, err := parseUint(c.Param("uid"))
		if err != nil {
			return err
		}
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return nil
		}
		enrollment, err := db.GetEnrollmentByCourseAndUser(cid, uid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		if enrollment.GroupID > 0 {
			group, err := db.GetGroup(enrollment.GroupID)
			if err != nil {
				return nil
			}
			return c.JSONPretty(http.StatusFound, group, "\t")
		}
		return c.NoContent(http.StatusNotFound)
	}
}
