package web

import (
	"net/http"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"

	//"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
// Updaterequestuest updates a user object in the database.
type Updaterequestuest struct {
	Name      string `json:"name"`
	StudentID string `json:"studentid"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarurl"`
	IsAdmin   *bool  `json:"isadmin"`
}*/

// GetSelf redirects to GetUser with the current user's id.
func GetSelf(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		usr := c.Get("user").(*pb.User)
		user, err := db.GetUser(usr.GetId())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		return c.JSONPretty(http.StatusFound, user, "\t")
	}
}

// GetUser returns information about the provided user id.
func GetUser(request *pb.RecordRequest, db database.Database) (*pb.User, error) {
	user, err := db.GetUser(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, err
	}

	//TODO(vera): this can be removed - remote identities will not be sent over http
	/*
		// Remove access token for user because otherhewise anyone can get access to user tokens
		for _, remoteID := range user.GetRemoteIdentities() {
			remoteID.AccessToken = ""
		}*/

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

// UpdateUser promotes a user to an administrator or makes other changes to the user database entry.
func PatchUser(currentUser *pb.User, request *pb.User, db database.Database) (*pb.User, error) {
	updateUser, err := db.GetUser(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if request.Name != "" {
		updateUser.Name = request.Name
	}
	if request.StudentId != "" {
		updateUser.StudentId = request.StudentId
	}
	if request.Email != "" {
		updateUser.Email = request.Email
	}
	if request.AvatarUrl != "" {
		updateUser.AvatarUrl = request.AvatarUrl
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
