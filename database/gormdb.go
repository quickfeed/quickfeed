package database

import (
	"errors"

	pb "github.com/autograde/aguis/ag"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrDuplicateIdentity is returned when trying to associate a remote identity
	// with a user account and the identity is already in use.
	ErrDuplicateIdentity = errors.New("remote identity registered with another user")
	// ErrEmptyGroup is returned when trying to create a group without users.
	ErrEmptyGroup = errors.New("cannot create group without users")
	// ErrDuplicateGroup is returned when trying to create a group with the same
	// name as a previously registered group.
	ErrDuplicateGroup = status.Error(codes.InvalidArgument, "group with this name already registered")
	// ErrUpdateGroup is returned when updating a group's enrollment fails.
	ErrUpdateGroup = errors.New("failed to update group enrollment")
	// ErrCourseExists is returned when trying to create an association in
	// the database for a DirectoryId that already exists in the database.
	ErrCourseExists = errors.New("course already exists on git provider")
	// ErrInsufficientAccess is returned when trying to update database
	// with insufficient access priviledges.
	ErrInsufficientAccess = errors.New("user must be admin to perform this operation")
	// ErrCreateRepo is returned when trying to create repository with wrong argument.
	ErrCreateRepo = errors.New("failed to create repository; invalid arguments")
)

// GormDB implements the Database interface.
type GormDB struct {
	conn *gorm.DB
}

// NewGormDB creates a new gorm database using the provided driver.
func NewGormDB(driver, path string, logger GormLogger) (*GormDB, error) {
	conn, err := gorm.Open(driver, path)
	if err != nil {
		return nil, err
	}

	if logger != nil {
		conn.SetLogger(logger)
	}
	conn.LogMode(logger != nil)

	if err := conn.AutoMigrate(
		&pb.User{},
		&pb.RemoteIdentity{},
		&pb.Course{},
		&pb.Enrollment{},
		&pb.Assignment{},
		&pb.Submission{},
		&pb.Group{},
		&pb.Repository{},
	).Error; err != nil {
		return nil, err
	}

	return &GormDB{conn}, nil
}

///  Remote Identities ///

// GetRemoteIdentity fetches remote identity by provider and ID.
func (db *GormDB) GetRemoteIdentity(provider string, remoteID uint64) (*pb.RemoteIdentity, error) {
	var remoteIdentity pb.RemoteIdentity
	if err := db.conn.Model(&pb.RemoteIdentity{}).
		Where(&pb.RemoteIdentity{
			Provider: provider,
			RemoteID: remoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		return nil, err
	}
	return &remoteIdentity, nil
}

// CreateUserFromRemoteIdentity creates new user record from remote identity, sets user with ID 1 as admin.
func (db *GormDB) CreateUserFromRemoteIdentity(user *pb.User, remoteIdentity *pb.RemoteIdentity) error {
	user.RemoteIdentities = []*pb.RemoteIdentity{remoteIdentity}
	if err := db.conn.Create(&user).Error; err != nil {
		return err
	}
	// The first user defaults to admin user.
	if user.ID == 1 {
		user.IsAdmin = true
		if err := db.UpdateUser(user); err != nil {
			return err
		}
	}
	return nil
}

// AssociateUserWithRemoteIdentity associates remote identity with the user with given ID.
func (db *GormDB) AssociateUserWithRemoteIdentity(uid uint64, provider string, remoteID uint64, accessToken string) error {
	var count uint64
	if err := db.conn.
		Model(&pb.RemoteIdentity{}).
		Where(&pb.RemoteIdentity{
			Provider: provider,
			RemoteID: remoteID,
		}).
		Not(&pb.RemoteIdentity{
			UserID: uid,
		}).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateIdentity
	}

	var remoteIdentity pb.RemoteIdentity
	return db.conn.
		Where(pb.RemoteIdentity{Provider: provider, RemoteID: remoteID, UserID: uid}).
		Assign(pb.RemoteIdentity{AccessToken: accessToken}).
		FirstOrCreate(&remoteIdentity).Error
}

// UpdateAccessToken refreshes the token info for the given remote identity.
func (db *GormDB) UpdateAccessToken(remote *pb.RemoteIdentity) error {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity pb.RemoteIdentity
	if err := tx.
		Where(&pb.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update the access token.
	if err := tx.Model(&remoteIdentity).Update("access_token", remote.AccessToken).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// updateAccessTokenCache caches the access token for the course
// to allow easy access elsewhere.
func (db *GormDB) updateAccessTokenCache(course *pb.Course) {
	existingToken := course.GetAccessToken()
	if existingToken != "" {
		// no need to cache again
		return
	}
	// only need to query db if not in cache
	courseCreator, err := db.GetUser(course.GetCourseCreatorID())
	if err != nil {
		// failed to get course creator; ignore
		return
	}
	accessToken, err := courseCreator.GetAccessToken(course.GetProvider())
	if err != nil {
		// failed to get access token for course creator; ignore
		return
	}
	// update the access token cache
	pb.SetAccessToken(course.GetID(), accessToken)
}
