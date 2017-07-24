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
func (db *GormDB) GetUsers() ([]*models.User, error) {
	var users []*models.User
	if err := db.conn.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUsersByCourse implements the Database interface.
func (db *GormDB) GetUsersByCourse(courseCode string) ([]*models.User, error) {
	course, err := db.GetCourseByCode(courseCode)
	if err != nil {
		return nil, err
	}

	enrollments, err := db.GetEnrollmentsByCourse(course.ID)
	if err != nil {
		return nil, err
	}

	var users []*models.User
	for _, enrollment := range enrollments {
		user, err := db.GetUser(enrollment.UserID)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
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
		RemoteIdentities: []*models.RemoteIdentity{{
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
func (db *GormDB) GetCourses() ([]*models.Course, error) {
	var courses []*models.Course
	if err := db.conn.Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

// GetAssignmentsByCourse implements the Database interface
func (db *GormDB) GetAssignmentsByCourse(id uint64) ([]*models.Assignment, error) {
	var course models.Course
	if err := db.conn.Preload("Assignments").First(&course, id).Error; err != nil {
		return nil, err
	}
	return course.Assignments, nil
}

// CreateAssignment implements the Database interface
func (db *GormDB) CreateAssignment(assignment *models.Assignment) error {
	var course uint64
	if err := db.conn.Model(&models.Course{}).Where(&models.Course{
		ID: assignment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}

	if course != 1 {
		return gorm.ErrRecordNotFound
	}
	return db.conn.
		Where(models.Assignment{CourseID: assignment.CourseID, AssignmentID: assignment.AssignmentID}).
		Assign(models.Assignment{
			Name:        assignment.Name,
			Language:    assignment.Language,
			Deadline:    assignment.Deadline,
			AutoApprove: assignment.AutoApprove,
		}).FirstOrCreate(assignment).Error
}

// CreateEnrollment implements the Database interface.
func (db *GormDB) CreateEnrollment(enrollment *models.Enrollment) error {
	var user, course uint64
	if err := db.conn.Model(&models.User{}).Where(&models.User{
		ID: enrollment.UserID,
	}).Count(&user).Error; err != nil {
		return err
	}
	if err := db.conn.Model(&models.Course{}).Where(&models.Course{
		ID: enrollment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}

	if user+course != 2 {
		return gorm.ErrRecordNotFound
	}

	return db.conn.Where(enrollment).FirstOrCreate(enrollment).Error
}

// AcceptEnrollment implements the Database interface.
func (db *GormDB) AcceptEnrollment(id uint64) error {
	return db.setEnrollment(id, models.Accepted)
}

// RejectEnrollment implements the Database interface.
func (db *GormDB) RejectEnrollment(id uint64) error {
	return db.setEnrollment(id, models.Rejected)
}

// GetEnrollmentsByUser implements the Database interface.
func (db *GormDB) GetEnrollmentsByUser(id uint64, statuses ...uint) ([]*models.Enrollment, error) {
	return db.getEnrollments(&models.User{ID: id}, statuses...)
}

// GetEnrollmentsByCourse implements the Database interface.
func (db *GormDB) GetEnrollmentsByCourse(id uint64, statuses ...uint) ([]*models.Enrollment, error) {
	return db.getEnrollments(&models.Course{ID: id}, statuses...)
}

func (db *GormDB) getEnrollments(model interface{}, statuses ...uint) ([]*models.Enrollment, error) {
	if len(statuses) == 0 {
		statuses = []uint{models.Pending, models.Rejected, models.Accepted}
	}
	var enrollments []*models.Enrollment
	if err := db.conn.Model(model).Where("status in (?)", statuses).Association("Enrollments").Find(&enrollments).Error; err != nil {
		return nil, err
	}

	return enrollments, nil
}

func (db *GormDB) setEnrollment(id uint64, status uint) error {
	if status > models.Accepted {
		panic("invalid status")
	}
	return db.conn.Model(&models.Enrollment{}).Where(&models.Enrollment{ID: id}).Update(&models.Enrollment{
		Status: int(status),
	}).Error
}

// GetCoursesByUser returns all courses with the users enrollment status
// included.
func (db *GormDB) GetCoursesByUser(id uint64) ([]*models.Course, error) {
	courses, err := db.GetCourses()
	if err != nil {
		return nil, err
	}

	enrollments, err := db.GetEnrollmentsByUser(id)
	if err != nil {
		return nil, err
	}

	m := make(map[uint64]*models.Enrollment)
	for _, enrollment := range enrollments {
		m[enrollment.CourseID] = enrollment
	}

	for _, course := range courses {
		// cannot take address of a constant, so variable none is used instead for passing address of models.None
		none := models.None
		course.Enrolled = &none
		if enrollment, ok := m[course.ID]; ok {
			course.Enrolled = &enrollment.Status
		}
	}

	return courses, nil
}

// GetActiveCoursesByUser returns all active courses of a user
func (db *GormDB) GetActiveCoursesByUser(id uint64) ([]*models.Course, error) {
	enrollments, err := db.getEnrollments(&models.User{ID: id}, models.Accepted)
	if err != nil {
		return nil, err
	}
	courseIDs := []uint64{}
	for _, enrollment := range enrollments {
		courseIDs = append(courseIDs, enrollment.CourseID)
	}

	var courses []*models.Course
	if err := db.conn.Where(courseIDs).Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

// GetCourse implements the Database interface
func (db *GormDB) GetCourse(id uint64) (*models.Course, error) {
	var course models.Course
	if err := db.conn.First(&course, id).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourseByCode implements the Database interface
func (db *GormDB) GetCourseByCode(code string) (*models.Course, error) {
	var course models.Course
	if err := db.conn.Where("code = ?", code).First(&course).Error; err != nil {
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
