package database

import (
	"context"
	"errors"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
)

// CreateCourse creates a new course if user with given ID is admin, enrolls user as course teacher.
func (db *BunDB) CreateCourse(courseCreatorID uint64, course *qf.Course) error {
	ctx := context.Background()
	courseCreator, err := db.GetUser(courseCreatorID)
	if err != nil {
		return err
	}
	if !courseCreator.GetIsAdmin() {
		return ErrInsufficientAccess
	}

	exists, err := db.conn.NewSelect().Model((*qf.Course)(nil)).
		Where("scm_organization_id = ?", course.GetScmOrganizationID()).
		WhereOr("(code = ? AND year = ?)", course.GetCode(), course.GetYear()).
		Exists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return ErrCourseExists
	}

	course.CourseCreatorID = courseCreatorID
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(course).Exec(ctx); err != nil {
			return err
		}
		enrollment := &qf.Enrollment{
			UserID:   courseCreatorID,
			CourseID: course.GetID(),
			Status:   qf.Enrollment_TEACHER,
			State:    qf.Enrollment_VISIBLE,
		}
		_, err := tx.NewInsert().Model(enrollment).Exec(ctx)
		return err
	})
}

// GetCourse fetches course by ID.
func (db *BunDB) GetCourse(courseID uint64) (*qf.Course, error) {
	ctx := context.Background()
	var course qf.Course
	if err := db.conn.NewSelect().Model(&course).Where("id = ?", courseID).Scan(ctx); err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourseByStatus fetches course by ID with preloaded data based on enrollment status.
func (db *BunDB) GetCourseByStatus(courseID uint64, status qf.Enrollment_UserStatus) (*qf.Course, error) {
	ctx := context.Background()
	var course qf.Course
	q := db.conn.NewSelect().Model(&course).Where("course.id = ?", courseID)
	switch status {
	case qf.Enrollment_NONE, qf.Enrollment_PENDING:
		// no preloaded data
	case qf.Enrollment_STUDENT:
		userStates := []qf.Enrollment_UserStatus{
			qf.Enrollment_STUDENT,
			qf.Enrollment_TEACHER,
		}
		q = q.
			Relation("Assignments").
			Relation("Enrollments", func(sq *bun.SelectQuery) *bun.SelectQuery {
				return sq.Where("status IN (?)", bun.In(userStates))
			}).
			Relation("Enrollments.User")
	case qf.Enrollment_TEACHER:
		q = q.
			Relation("Assignments").
			Relation("Enrollments").
			Relation("Enrollments.User").
			Relation("Enrollments.Group").
			Relation("Enrollments.UsedSlipDays").
			Relation("Groups", func(sq *bun.SelectQuery) *bun.SelectQuery {
				return sq.Where("course_id = ?", courseID)
			}).
			Relation("Groups.Users")
	default:
		return nil, errors.New("invalid enrollment status")
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	if status == qf.Enrollment_TEACHER {
		course.PopulateSlipDays()
	}
	return &course, nil
}

// GetCourseByOrganizationID fetches course by organization ID.
func (db *BunDB) GetCourseByOrganizationID(organizationID uint64) (*qf.Course, error) {
	ctx := context.Background()
	var course qf.Course
	if err := db.conn.NewSelect().Model(&course).Where("scm_organization_id = ?", organizationID).Scan(ctx); err != nil {
		return nil, err
	}
	return &course, nil
}

// GetCourses returns a list of courses. If course IDs are provided, only those courses are returned.
func (db *BunDB) GetCourses(courseIDs ...uint64) ([]*qf.Course, error) {
	ctx := context.Background()
	var courses []*qf.Course
	q := db.conn.NewSelect().Model(&courses)
	if len(courseIDs) > 0 {
		q = q.Where("id IN (?)", bun.In(courseIDs))
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return courses, nil
}

// GetCoursesByUser returns all courses for the given user, optionally filtered by enrollment status.
func (db *BunDB) GetCoursesByUser(userID uint64, statuses ...qf.Enrollment_UserStatus) ([]*qf.Course, error) {
	enrollments, err := db.getEnrollments("user", userID, statuses...)
	if err != nil {
		return nil, err
	}

	var courseIDs []uint64
	m := make(map[uint64]*qf.Enrollment)
	for _, enrollment := range enrollments {
		m[enrollment.GetCourseID()] = enrollment
		courseIDs = append(courseIDs, enrollment.GetCourseID())
	}

	if len(statuses) == 0 {
		courseIDs = nil
	} else if len(courseIDs) == 0 {
		return []*qf.Course{}, nil
	}

	courses, err := db.GetCourses(courseIDs...)
	if err != nil {
		return nil, err
	}

	for _, course := range courses {
		course.Enrolled = qf.Enrollment_NONE
		if enrollment, ok := m[course.GetID()]; ok {
			course.Enrolled = enrollment.GetStatus()
		}
	}
	return courses, nil
}

// GetCourseTeachers returns a list of all teachers in a course.
func (db *BunDB) GetCourseTeachers(query *qf.Course) ([]*qf.User, error) {
	ctx := context.Background()
	var enrollments []*qf.Enrollment
	if err := db.conn.NewSelect().
		Model(&enrollments).
		Where("course_id = ? AND status = ?", query.GetID(), qf.Enrollment_TEACHER).
		Scan(ctx); err != nil {
		return nil, err
	}
	if len(enrollments) == 0 {
		return nil, errors.New("course has no teachers")
	}
	teachers := make([]*qf.User, 0, len(enrollments))
	for _, e := range enrollments {
		teacher, err := db.GetUser(e.GetUserID())
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

// UpdateCourse updates course information.
func (db *BunDB) UpdateCourse(course *qf.Course) error {
	ctx := context.Background()
	_, err := db.conn.NewUpdate().Model(course).OmitZero().WherePK().Exec(ctx)
	return err
}
