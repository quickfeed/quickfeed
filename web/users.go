package web

import (
	"net/http"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// JSONuser is a model to improve marshalling of user structure for authentication
type JSONuser struct {
	ID        uint64 `json:"id"`
	IsAdmin   *bool  `json:"isadmin"`
	Name      string `json:"name"`
	StudentID string `json:"studentid"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarurl"`
}

// GetSelf redirects to GetUser with the current user's id.
func GetSelf(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		usr := c.Get("user").(*pb.User)

		// defer closing the http request body
		defer c.Request().Body.Close()

		user, err := db.GetUser(usr.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		jsonUser := JSONuser{ID: user.ID, IsAdmin: &user.IsAdmin, Name: user.Name, StudentID: user.StudentID, Email: user.Email, AvatarURL: user.AvatarURL}
		return c.JSONPretty(http.StatusFound, jsonUser, "\t")
	}
}

// getUsers returns all the users in the database.
func (s *AutograderService) getUsers() (*pb.Users, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	return &pb.Users{Users: users}, nil
}

// updateUser promotes the given user to administrator and/or
// make other changes to the user database entry for the given user.
// The isAdmin flag must be true to promote the given user to admin.
func (s *AutograderService) updateUser(isAdmin bool, request *pb.User) (*pb.User, error) {
	updateUser, err := s.db.GetUser(request.ID)
	if err != nil {
		return nil, err
	}

	if request.Name != "" {
		updateUser.Name = request.Name
	}
	if request.StudentID != "" {
		updateUser.StudentID = request.StudentID
	}
	if request.Email != "" {
		updateUser.Email = request.Email
	}
	if request.AvatarURL != "" {
		updateUser.AvatarURL = request.AvatarURL
	}
	// current user must be admin to change admin status of another user
	// admin status of super admin (user with ID 1) cannot be changed
	if isAdmin && request.ID > 1 {
		updateUser.IsAdmin = request.IsAdmin
	}

	err = s.db.UpdateUser(updateUser)
	return updateUser, err
}
