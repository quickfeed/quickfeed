package web

import (
	"fmt"
	"net/http"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// UpdateUserRequest updates a user object at the database
type UpdateUserRequest struct {
	IsAdmin *bool `json:"isadmin"`
}

func (uur *UpdateUserRequest) isSetIsAdmin() bool {
	return uur.IsAdmin != nil
}

// GetSelf redirects to GetUser with the current user's id.
func GetSelf() echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*models.User)
		return c.Redirect(http.StatusFound, fmt.Sprintf("/api/v1/users/%d", user.ID))
	}
}

// GetUser returns information about the user associated with the id query.
func GetUser(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := ParseUintParam(c.Param("uid"))
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
// TODO: Add filtering, i.e, ?course=A123.
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
		id, err := ParseUintParam(c.Param("uid"))
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
