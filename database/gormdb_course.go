package database

import (
	"errors"

	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

// CreateCourse creates a new course if user with given ID is admin, enrolls user as course teacher.
// The provided course must have a unique (GitHub) OrganizationID not already associated with existing course.
// Similarly, the course must have a unique course code and year.
func (db *GormDB) CreateCourse(courseCreatorID uint64, course *qf.Course) error {
	courseCreator, err := db.GetUser(courseCreatorID)
	if err != nil {
		return err
	}
	if !courseCreator.IsAdmin {
		return ErrInsufficientAccess
	}

	var courses int64
	if err := db.conn.Model(&qf.Course{}).Where(&qf.Course{
		ScmOrganizationID: course.ScmOrganizationID,
	}).Or(&qf.Course{
		Code: course.Code,
		Year: course.Year,
	}).Count(&courses).Error; err != nil {
		return err
	}
	if courses > 0 {
		return ErrCourseExists
	}

	course.CourseCreatorID = courseCreatorID

	tx := db.conn.Begin()
	if err := tx.Create(course).Error; err != nil {
		tx.Rollback()
		return err
	}

	// enroll course creator as teacher for course and mark as visible
	if err := tx.Create(&qf.Enrollment{
		UserID:   courseCreatorID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
		State:    qf.Enrollment_VISIBLE,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetCourse fetches course by ID.
func (db *GormDB) GetCourse(courseID uint64) (*qf.Course, error) {
	var course qf.Course
	if err := db.conn.First(&course, courseID).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourseByStatus fetches course by ID. Depending on the enrollment status,
// it preloads course assignments, active enrollments, users, and groups.
func (db *GormDB) GetCourseByStatus(courseID uint64, status qf.Enrollment_UserStatus) (*qf.Course, error) {
	m := db.conn
	var course qf.Course
	switch status {
	case qf.Enrollment_NONE, qf.Enrollment_PENDING:
		// no preloaded data
	case qf.Enrollment_STUDENT:
		userStates := []qf.Enrollment_UserStatus{
			qf.Enrollment_STUDENT,
			qf.Enrollment_TEACHER,
		}
		m = m.Preload("Assignments").
			Preload("Enrollments", "status in (?)", userStates, func(db *gorm.DB) *gorm.DB {
				return db.Omit("UsedSlipDays", "LastActivityDate")
			}).
			Preload("Enrollments.User", func(db *gorm.DB) *gorm.DB {
				return db.Omit("Login", "UpdateToken", "Email")
			})
	case qf.Enrollment_TEACHER:
		// Preload all data
		modelGroup := &qf.Group{Status: qf.Group_APPROVED, CourseID: courseID}
		m = m.Preload("Assignments").
			Preload("Enrollments").
			Preload("Enrollments.User").
			Preload("Enrollments.Group").
			Preload("Enrollments.UsedSlipDays").
			Preload("Groups", modelGroup).
			Preload("Groups.Users")
	default:
		return nil, errors.New("invalid enrollment status")
	}
	if err := m.First(&course, courseID).Error; err != nil {
		return nil, err
	}
	if status == qf.Enrollment_TEACHER {
		course.PopulateSlipDays()
	}
	return &course, nil
}

// GetCourseByOrganizationID fetches course by organization ID.
func (db *GormDB) GetCourseByOrganizationID(did uint64) (*qf.Course, error) {
	var course qf.Course
	if err := db.conn.First(&course, &qf.Course{ScmOrganizationID: did}).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourses returns a list of courses. If one or more course ids are provided,
// the corresponding courses are returned. Otherwise, all courses are returned.
func (db *GormDB) GetCourses(courseIDs ...uint64) ([]*qf.Course, error) {
	m := db.conn
	if len(courseIDs) > 0 {
		m = m.Where(courseIDs)
	}
	var courses []*qf.Course
	if err := m.Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

// GetCoursesByUser returns all courses (with enrollment status)
// for the given user id.
// If enrollment statuses is provided, the set of courses returned
// is filtered according to these enrollment statuses.
func (db *GormDB) GetCoursesByUser(userID uint64, statuses ...qf.Enrollment_UserStatus) ([]*qf.Course, error) {
	enrollments, err := db.getEnrollments(&qf.User{ID: userID}, statuses...)
	if err != nil {
		return nil, err
	}

	var courseIDs []uint64
	m := make(map[uint64]*qf.Enrollment)
	for _, enrollment := range enrollments {
		m[enrollment.CourseID] = enrollment
		courseIDs = append(courseIDs, enrollment.CourseID)
	}

	if len(statuses) == 0 {
		courseIDs = nil
	} else if len(courseIDs) == 0 {
		// No need to query database since user have no enrolled courses.
		return []*qf.Course{}, nil
	}
	courses, err := db.GetCourses(courseIDs...)
	if err != nil {
		return nil, err
	}

	for _, course := range courses {
		course.Enrolled = qf.Enrollment_NONE
		if enrollment, ok := m[course.ID]; ok {
			course.Enrolled = enrollment.Status
		}
	}
	return courses, nil
}

// GetCourseTeachers returns a list of all teachers in a course.
func (db *GormDB) GetCourseTeachers(query *qf.Course) ([]*qf.User, error) {
	var course qf.Course
	if err := db.conn.Where(query).Preload("Enrollments").First(&course).Error; err != nil {
		return nil, err
	}
	teachers := []*qf.User{}
	for _, teacherEnrollment := range course.TeacherEnrollments() {
		teacher, err := db.GetUser(teacherEnrollment.GetUserID())
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	if len(teachers) == 0 {
		return nil, errors.New("course has no teachers")
	}
	return teachers, nil
}

// UpdateCourse updates course information.
func (db *GormDB) UpdateCourse(course *qf.Course) error {
	return db.conn.Model(&qf.Course{}).
		Where(&qf.Course{ID: course.GetID()}).
		Updates(course).Error
}
