package web

import (
	"fmt"
	"log"
	"net/http"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"

	//"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// User is a test model to improve marshalling of user structure for authentication
type User struct {
	ID uint64 `json:"id"`

	IsAdmin *bool `json:"isadmin"`

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
		fmt.Println("GetSelf got user from session context: ", usr.ID)
		user, err := db.GetUser(usr.ID)
		fmt.Println("GetSelf got user from the database with ID: ", usr.ID)

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		jsonUser := User{ID: user.ID, IsAdmin: &user.IsAdmin, Name: user.Name, StudentID: user.StudentID, Email: user.Email, AvatarURL: user.AvatarURL}
		log.Println("Marshalled user is ", jsonUser)
		return c.JSONPretty(http.StatusFound, jsonUser, "\t")
	}
}

// GetUser returns information about the provided user id.
func GetUser(request *pb.RecordRequest, db database.Database) (*pb.User, error) {
	if request.ID < 1 {
		fmt.Println("GetUser requested user with ID: ", request.ID)
		return nil, status.Errorf(codes.Aborted, "invalid argument")
	}
	user, err := db.GetUser(request.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, err
	}
	return user, nil
}

// GetUsers returns all the users in the database.
func GetUsers(db database.Database) (*pb.Users, error) {
	users, err := db.GetUsers()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "no users found")
		}
		return nil, err
	}
	return &pb.Users{Users: users}, nil
}

// PatchUser promotes a user to an administrator or makes other changes to the user database entry.
func PatchUser(currentUser *pb.User, request *pb.User, db database.Database) (*pb.User, error) {
	updateUser, err := db.GetUser(request.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
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

	// no need to check IsAdmin field for nil any more, it is type safe - it is always boolean and cannot be nil
	if currentUser.IsAdmin {
		updateUser.IsAdmin = request.IsAdmin
	}
	if err := db.UpdateUser(updateUser); err != nil {
		return nil, err
	}
	return updateUser, nil
}
