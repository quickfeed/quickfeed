package database

import (
	"errors"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	// ErrDuplicateIdentity is returned when trying to associate a remote identity
	// with a user account and the identity is already in use.
	ErrDuplicateIdentity = errors.New("remote identity registered with another user")
	// ErrEmptyGroup is returned when trying to create a group without users.
	ErrEmptyGroup = errors.New("cannot create group without users")
	// ErrDuplicateGroup is returned when trying to create a group with the same
	// name as a previously registered group.
	ErrDuplicateGroup = errors.New("group with this name already registered")
	// ErrUpdateGroup is returned when updating a group's enrollment fails.
	ErrUpdateGroup = errors.New("failed to update group enrollment")
	// ErrCourseExists is returned when trying to create an association in
	// the database for a DirectoryId that already exists in the database.
	ErrCourseExists = errors.New("course already exists on git provider")
	// ErrInsufficientAccess is returned when trying to update database
	// with insufficient access privileges.
	ErrInsufficientAccess = errors.New("user must be admin to perform this operation")
	// ErrCreateRepo is returned when trying to create repository with wrong argument.
	ErrCreateRepo = errors.New("failed to create repository; invalid arguments")
	// ErrNotEnrolled is returned when the requested user or group do not have
	// the expected association with the given course
	ErrNotEnrolled = errors.New("user or group not enrolled in the course")
)

// GormDB implements the Database interface.
type GormDB struct {
	conn *gorm.DB
}

// NewGormDB creates a new gorm database using the provided driver.
func NewGormDB(path string, logger *zap.Logger) (*GormDB, error) {
	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: NewGORMLogger(logger),
	})
	if err != nil {
		return nil, err
	}

	schema.RegisterSerializer("timestamp", &TimestampSerializer{})

	if err := conn.AutoMigrate(
		&qf.User{},
		&qf.RemoteIdentity{},
		&qf.Course{},
		&qf.Enrollment{},
		&qf.Assignment{},
		&qf.Submission{},
		&qf.Group{},
		&qf.Repository{},
		&qf.UsedSlipDays{},
		&qf.GradingBenchmark{},
		&qf.GradingCriterion{},
		&qf.Review{},
		&qf.Issue{},
		&qf.Task{},
		&qf.PullRequest{},
		&score.BuildInfo{},
		&score.Score{},
	); err != nil {
		return nil, err
	}

	return &GormDB{conn}, nil
}

///  Remote Identities ///

// CreateUserFromRemoteIdentity creates new user record from remote identity, sets user with ID 1 as admin.
func (db *GormDB) CreateUserFromRemoteIdentity(user *qf.User, remoteIdentity *qf.RemoteIdentity) error {
	user.RemoteIdentities = []*qf.RemoteIdentity{remoteIdentity}
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
	var count int64
	if err := db.conn.
		Model(&qf.RemoteIdentity{}).
		Where(&qf.RemoteIdentity{
			Provider: provider,
			RemoteID: remoteID,
		}).
		Not(&qf.RemoteIdentity{
			UserID: uid,
		}).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateIdentity
	}

	var remoteIdentity qf.RemoteIdentity
	return db.conn.
		Where(qf.RemoteIdentity{Provider: provider, RemoteID: remoteID, UserID: uid}).
		Assign(qf.RemoteIdentity{AccessToken: accessToken}).
		FirstOrCreate(&remoteIdentity).Error
}

// UpdateAccessToken refreshes the token info for the given remote identity.
func (db *GormDB) UpdateAccessToken(remote *qf.RemoteIdentity) error {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity qf.RemoteIdentity
	if err := tx.
		Where(&qf.RemoteIdentity{
			Provider: remote.Provider,
			RemoteID: remote.RemoteID,
		}).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&remoteIdentity).Update("access_token", remote.AccessToken).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (db *GormDB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
