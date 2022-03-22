package database

import (
	pb "github.com/autograde/quickfeed/ag"
)

// CreateCourse creates a new course if user with given ID is admin, enrolls user as course teacher.
// The provided course must have a unique (GitHub) OrganizationID not already associated with existing course.
// Similarly, the course must have a unique course code and year.
func (db *GormDB) CreateCourse(courseCreatorID uint64, course *pb.Course) error {
	courseCreator, err := db.GetUser(courseCreatorID)
	if err != nil {
		return err
	}
	if !courseCreator.IsAdmin {
		return ErrInsufficientAccess
	}

	var courses int64
	if err := db.conn.Model(&pb.Course{}).Where(&pb.Course{
		OrganizationID: course.OrganizationID,
	}).Or(&pb.Course{
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
	if err := tx.Create(&pb.Enrollment{
		UserID:   courseCreatorID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
		State:    pb.Enrollment_VISIBLE,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// fetch course creator's most recent access token
	accessToken, err := courseCreator.GetAccessToken(course.GetProvider())
	if err != nil {
		tx.Rollback()
		return err
	}
	// update the access token cache for course
	pb.SetAccessToken(course.GetID(), accessToken)
	return tx.Commit().Error
}

// GetCourse fetches course by ID. If withInfo is true, preloads course
// assignments, active enrollments and groups.
func (db *GormDB) GetCourse(courseID uint64, withEnrollments bool) (*pb.Course, error) {
	m := db.conn
	var course pb.Course

	if withEnrollments {
		// we only want submission from users enrolled in the course
		userStates := []pb.Enrollment_UserStatus{
			pb.Enrollment_STUDENT,
			pb.Enrollment_TEACHER,
		}
		// and only group submissions from approved groups
		modelGroup := &pb.Group{Status: pb.Group_APPROVED, CourseID: courseID}
		if err := m.Preload("Assignments").
			Preload("Enrollments", "status in (?)", userStates).
			Preload("Enrollments.User").
			Preload("Enrollments.Group").
			Preload("Enrollments.UsedSlipDays").
			Preload("Groups", modelGroup).
			First(&course, courseID).Error; err != nil {
			return nil, err
		}
	} else {
		if err := m.First(&course, courseID).Error; err != nil {
			return nil, err
		}
	}
	if err := db.updateCourseAccessTokenIfEmpty(&course); err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourseByOrganizationID fetches course by organization ID.
func (db *GormDB) GetCourseByOrganizationID(did uint64) (*pb.Course, error) {
	var course pb.Course
	if err := db.conn.First(&course, &pb.Course{OrganizationID: did}).Error; err != nil {
		return nil, err
	}
	if err := db.updateCourseAccessTokenIfEmpty(&course); err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourses returns a list of courses. If one or more course ids are provided,
// the corresponding courses are returned. Otherwise, all courses are returned.
func (db *GormDB) GetCourses(courseIDs ...uint64) ([]*pb.Course, error) {
	m := db.conn
	if len(courseIDs) > 0 {
		m = m.Where(courseIDs)
	}
	var courses []*pb.Course
	if err := m.Find(&courses).Error; err != nil {
		return nil, err
	}
	for _, course := range courses {
		if err := db.updateCourseAccessTokenIfEmpty(course); err != nil {
			return nil, err
		}
	}
	return courses, nil
}

// GetCoursesByUser returns all courses (with enrollment status)
// for the given user id.
// If enrollment statuses is provided, the set of courses returned
// is filtered according to these enrollment statuses.
func (db *GormDB) GetCoursesByUser(userID uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Course, error) {
	enrollments, err := db.getEnrollments(&pb.User{ID: userID}, statuses...)
	if err != nil {
		return nil, err
	}

	var courseIDs []uint64
	m := make(map[uint64]*pb.Enrollment)
	for _, enrollment := range enrollments {
		m[enrollment.CourseID] = enrollment
		courseIDs = append(courseIDs, enrollment.CourseID)
	}

	if len(statuses) == 0 {
		courseIDs = nil
	} else if len(courseIDs) == 0 {
		// No need to query database since user have no enrolled courses.
		return []*pb.Course{}, nil
	}
	courses, err := db.GetCourses(courseIDs...)
	if err != nil {
		return nil, err
	}

	for _, course := range courses {
		if err := db.updateCourseAccessTokenIfEmpty(course); err != nil {
			return nil, err
		}
		course.Enrolled = pb.Enrollment_NONE
		if enrollment, ok := m[course.ID]; ok {
			course.Enrolled = enrollment.Status
		}
	}
	return courses, nil
}

// UpdateCourse updates course information.
func (db *GormDB) UpdateCourse(course *pb.Course) error {
	return db.conn.Model(&pb.Course{}).
		Where(&pb.Course{ID: course.GetID()}).
		Updates(course).Error
}
