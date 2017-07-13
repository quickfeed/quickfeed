package database

import (
	"errors"

	"github.com/autograde/aguis/models"
	"github.com/jinzhu/gorm"
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
		conn.LogMode(true)
		conn.SetLogger(logger)
	}

	conn.AutoMigrate(
		&models.User{},
		&models.RemoteIdentity{},
		&models.Course{},
		&models.Enrollment{},
		&models.Assignment{},
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

// SetAdmin implements the Database interface.
func (db *GormDB) SetAdmin(id uint64) error {
	var user models.User
	if err := db.conn.First(&user, id).Error; err != nil {
		return err
	}
	user.IsAdmin = true
	return db.conn.Save(&user).Error
}

// CreateUserFromRemoteIdentity implements the Database interface.
func (db *GormDB) CreateUserFromRemoteIdentity(provider string, remoteID uint64, accessToken string) (*models.User, error) {
	var count int64
	if err := db.conn.
		Model(&models.RemoteIdentity{}).
		Where("provider = ? AND remote_id = ?", provider, remoteID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count != 0 {
		return nil, ErrDuplicateIdentity
	}

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

// ErrDuplicateIdentity is returned when trying to associate a remote identity
// with a user account and the identity is already in use.
var ErrDuplicateIdentity = errors.New("remote identity register with another user")

// AssociateUserWithRemoteIdentity implements the Database interface.
func (db *GormDB) AssociateUserWithRemoteIdentity(userID uint64, provider string, remoteID uint64, accessToken string) error {
	var count int64
	if err := db.conn.
		Model(&models.RemoteIdentity{}).
		Where("provider = ? AND remote_id = ? AND NOT user_id = ?", provider, remoteID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return ErrDuplicateIdentity
	}

	var remoteIdentity models.RemoteIdentity
	return db.conn.
		Where(models.RemoteIdentity{Provider: provider, RemoteID: remoteID, UserID: userID}).
		Assign(models.RemoteIdentity{AccessToken: accessToken}).
		FirstOrCreate(&remoteIdentity).Error
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

// GetCoursesForUser implements the Database interface.
func (db *GormDB) GetCoursesForUser(userID uint64) ([]*models.Enrollment, error) {
	var enrollments []*models.Enrollment
	if err := db.conn.Where(&models.Enrollment{
		UserID: userID,
	}).Find(&enrollments).Error; err != nil {
		return nil, err
	}
	return enrollments, nil
}

// GetAssignments implements the Database interface
func (db *GormDB) GetAssignments(id uint64) (*[]models.Assignment, error) {
	var course models.Course
	if err := db.conn.Preload("Assignments").First(&course, id).Error; err != nil {
		return nil, err
	}
	return &(course.Assignments), nil
}

// CreateAssignment implements the Database interface
func (db *GormDB) CreateAssignment(assignment *models.Assignment) error {
	return db.conn.Create(assignment).Error
}

// EnrollUserInCourse implements the Database interface.
func (db *GormDB) EnrollUserInCourse(userID, courseID uint64) error {
	return db.conn.Create(&models.Enrollment{UserID: userID, CourseID: courseID}).Error
}

// GetCourse implements the Database interface
func (db *GormDB) GetCourse(id uint64) (*models.Course, error) {
	var course models.Course
	if err := db.conn.First(&course, id).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// UpdateCourse implements the Database interface
func (db *GormDB) UpdateCourse(course *models.Course) error {
	return db.conn.Model(course).Updates(course).Error
}

// Close closes the gorm database.
func (db *GormDB) Close() error {
	return db.conn.Close()
}
