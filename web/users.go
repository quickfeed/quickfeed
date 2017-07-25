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
	IsAdmin *bool `json:"isadmin"`
}

func (uur *UpdateUserRequest) isSetIsAdmin() bool {
	return uur.IsAdmin != nil
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
		if uur.isSetIsAdmin() && *uur.IsAdmin {
			if err := db.SetAdmin(id); err != nil {
				return err
			}
			status = http.StatusOK
		}
		return c.NoContent(status)
	}
}
