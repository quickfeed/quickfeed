package database

import (
	"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
)

// GormDB implements the Database interface.
type GormDB struct {
	conn *gorm.DB
}

// NewGormDB creates a new gorm database using the provided driver.
func NewGormDB(driver, path string, debug bool) (*GormDB, error) {
	conn, err := gorm.Open(driver, path)
	if err != nil {
		return nil, err
	}

	conn.LogMode(debug)
	conn.AutoMigrate(
		&models.User{},
		&models.RemoteIdentity{},
		&models.Course{},
	)

	return &GormDB{conn}, nil
}

// GetUser implements the Database interface.
func (db *GormDB) GetUser(id uint64) (*models.User, error) {
	var user models.User
	if err := db.conn.Preload("RemoteIdentities").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByRemoteIdentity implements the Database interface.
func (db *GormDB) GetUserByRemoteIdentity(provider string, id uint64, accessToken string) (*models.User, error) {
	tx := db.conn.Begin()

	// Get the remote identity.
	var remoteIdentity models.RemoteIdentity
	if err := tx.
		Where("provider = ? AND remote_id = ?", provider, id).
		First(&remoteIdentity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update the access token.
	if err := tx.Model(&remoteIdentity).Update("access_token", accessToken).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get the user.
	var user models.User
	if err := tx.Preload("RemoteIdentities").First(&user, remoteIdentity.UserID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers implements the Database interface.
func (db *GormDB) GetUsers() (*[]models.User, error) {
	var users []models.User
	if err := db.conn.Find(&users).Error; err != nil {
		return nil, err
	}
	return &users, nil
}

// NewUserFromRemoteIdentity implements the Database interface.
func (db *GormDB) NewUserFromRemoteIdentity(provider string, remoteID uint64, accessToken string) (*models.User, error) {
	user := models.User{
		RemoteIdentities: []models.RemoteIdentity{{
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: accessToken,
		}},
	}
	if err := db.conn.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AssociateUserWithRemoteIdentity implements the Database interface.
func (db *GormDB) AssociateUserWithRemoteIdentity(userID uint64, provider string, remoteID uint64, accessToken string) error {
	remoteIdentity := models.RemoteIdentity{
		Provider:    provider,
		RemoteID:    remoteID,
		AccessToken: accessToken,
		UserID:      userID,
	}
	return db.conn.Create(&remoteIdentity).Error
}

// CreateCourse implements the Database interface.
func (db *GormDB) CreateCourse(course *models.Course) error {
	return db.conn.Create(course).Error
}

// GetCourses implements the Database interface.
func (db *GormDB) GetCourses() (*[]models.Course, error) {
	var courses []models.Course
	if err := db.conn.Find(&courses).Error; err != nil {
		return nil, err
	}
	return &courses, nil
}

// Close closes the gorm database.
func (db *GormDB) Close() error {
	return db.conn.Close()
}
