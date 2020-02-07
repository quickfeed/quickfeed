package database

import (
	pb "github.com/autograde/aguis/ag"
)

// CreateCourse creates a new course if user with given ID is admin, enrolls user as course teacher.
func (db *GormDB) CreateCourse(uid uint64, course *pb.Course) error {
	user, err := db.GetUser(uid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return ErrInsufficientAccess
	}

	var courses uint64
	if err := db.conn.Model(&pb.Course{}).Where(&pb.Course{
		OrganizationID: course.OrganizationID,
	}).Count(&courses).Error; err != nil {
		return err
	}
	if courses > 0 {
		return ErrCourseExists
	}

	//TODO(meling) these db updates should be done as a transaction
	if err := db.conn.Create(course).Error; err != nil {
		return err
	}
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: uid, CourseID: course.ID}); err != nil {
		return err
	}
	if err := db.EnrollTeacher(uid, course.ID); err != nil {
		return err
	}
	return nil
}

// GetCourse fetches course by ID. If withInfo is true, preloads course
// assignments, active enrollments and groups.
func (db *GormDB) GetCourse(cid uint64, withInfo bool) (*pb.Course, error) {
	m := db.conn
	var course pb.Course

	if withInfo {
		// we only want submission from users enrolled in the course
		userStates := []pb.Enrollment_UserStatus{
			pb.Enrollment_STUDENT,
			pb.Enrollment_TEACHER,
		}
		// and only group submissions from approved groups
		modelGroup := &pb.Group{Status: pb.Group_APPROVED, CourseID: cid}
		if err := m.Preload("Assignments").Preload("Enrollments", "status in (?)", userStates).Preload("Enrollments.User").Preload("Groups", modelGroup).First(&course, cid).Error; err != nil {
			return nil, err
		}
	} else {
		if err := m.First(&course, cid).Error; err != nil {
			return nil, err
		}
	}
	db.updateAccessTokenCache(&course)
	return &course, nil
}

// GetCourseByOrganizationID fetches course by organization ID.
func (db *GormDB) GetCourseByOrganizationID(did uint64) (*pb.Course, error) {
	var course pb.Course
	if err := db.conn.First(&course, &pb.Course{OrganizationID: did}).Error; err != nil {
		return nil, err
	}
	db.updateAccessTokenCache(&course)
	return &course, nil
}

// GetCourses returns a list of courses. If one or more course ids are provided,
// the corresponding courses are returned. Otherwise, all courses are returned.
func (db *GormDB) GetCourses(cids ...uint64) ([]*pb.Course, error) {
	m := db.conn
	if len(cids) > 0 {
		m = m.Where(cids)
	}
	var courses []*pb.Course
	if err := m.Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

// GetCoursesByUser returns all courses (with enrollment status)
// for the given user id.
// If enrollment statuses is provided, the set of courses returned
// is filtered according to these enrollment statuses.
func (db *GormDB) GetCoursesByUser(uid uint64, statuses ...pb.Enrollment_UserStatus) ([]*pb.Course, error) {
	enrollments, err := db.getEnrollments(&pb.User{ID: uid}, statuses...)
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
		course.Enrolled = pb.Enrollment_NONE
		if enrollment, ok := m[course.ID]; ok {
			course.Enrolled = enrollment.Status
		}
	}
	return courses, nil
}

// UpdateCourse updates course information.
func (db *GormDB) UpdateCourse(course *pb.Course) error {
	return db.conn.Model(&pb.Course{}).Updates(course).Error
}
