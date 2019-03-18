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
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
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
		// we don't want remote identities of users returned to the frontend
		users, err := db.GetUsers(false)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		return c.JSONPretty(http.StatusFound, users, "\t")
	}
}

// PatchUser updates a user's information, including promoting to administrator.
// Only existing administrators can promote another user.
func PatchUser(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := parseUint(c.Param("uid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		var uur UpdateUserRequest
		if err := c.Bind(&uur); err != nil {
			return err
		}

		status := http.StatusNotModified

		// get user to update
		updateUser, err := db.GetUser(id)
		if err != nil {
			return err
		}

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
		// get current user
		currentUser := c.Get("user").(*models.User)
		// promote other user to admin, only if current user has admin privileges
		if currentUser.IAdmin() && uur.IsAdmin != nil {
			updateUser.IsAdmin = uur.IsAdmin
			status = http.StatusOK
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
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		cid, err := parseUint(c.Param("cid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}
		enrollment, err := db.GetEnrollmentByCourseAndUser(cid, uid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		if enrollment.GroupID > 0 {
			// no need for remote identities
			group, err := db.GetGroup(false, enrollment.GroupID)
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}
			return c.JSONPretty(http.StatusFound, group, "\t")
		}
		return c.NoContent(http.StatusNotFound)
	}
}
