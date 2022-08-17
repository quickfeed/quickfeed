package database

import (
	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

/// Assignments ///

// CreateAssignment creates a new assignment record.
func (db *GormDB) CreateAssignment(assignment *qf.Assignment) error {
	// Course id and assignment order must be given.
	if assignment.CourseID < 1 || assignment.Order < 1 {
		return gorm.ErrRecordNotFound
	}

	var course int64
	if err := db.conn.Model(&qf.Course{}).Where(&qf.Course{
		ID: assignment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}

	if course != 1 {
		return gorm.ErrRecordNotFound
	}
	return db.conn.
		Where(qf.Assignment{
			CourseID: assignment.CourseID,
			Order:    assignment.Order,
		}).
		Assign(map[string]interface{}{
			"name":               assignment.Name,
			"order":              assignment.Order,
			"run_script_content": assignment.RunScriptContent,
			"deadline":           assignment.Deadline,
			"auto_approve":       assignment.AutoApprove,
			"score_limit":        assignment.ScoreLimit,
			"is_group_lab":       assignment.IsGroupLab,
			"reviewers":          assignment.Reviewers,
			"container_timeout":  assignment.ContainerTimeout,
			"tasks":              assignment.Tasks,
		}).Omit("Tasks").FirstOrCreate(assignment).Error
}

// GetAssignment returns assignment with the given ID.
func (db *GormDB) GetAssignment(query *qf.Assignment) (*qf.Assignment, error) {
	var assignment qf.Assignment
	if err := db.conn.Where(query).
		Preload("GradingBenchmarks").
		Preload("GradingBenchmarks.Criteria").
		First(&assignment).Error; err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetAssignmentsByCourse fetches all assignments for the given course ID.
func (db *GormDB) GetAssignmentsByCourse(courseID uint64, withGrading bool) ([]*qf.Assignment, error) {
	var course qf.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}
	assignments := course.Assignments
	if withGrading {
		for _, a := range assignments {
			var benchmarks []*qf.GradingBenchmark
			if err := db.conn.
				Where("assignment_id = ?", a.ID).
				Where("review_id = ?", 0).
				Find(&benchmarks).Error; err != nil {
				return nil, err
			}
			a.GradingBenchmarks = benchmarks
			for _, b := range a.GradingBenchmarks {
				var criteria []*qf.GradingCriterion
				if err := db.conn.Where("benchmark_id = ?", b.ID).Find(&criteria).Error; err != nil {
					return nil, err
				}
				b.Criteria = criteria
			}
		}
	}
	return assignments, nil
}

// UpdateAssignments updates assignment information.
func (db *GormDB) UpdateAssignments(assignments []*qf.Assignment) error {
	// TODO(meling) Updating the database may need locking?? Or maybe rewrite as a single query or txn.
	for _, v := range assignments {
		// this will create or update an existing assignment
		if err := db.CreateAssignment(v); err != nil {
			return err
		}
	}
	return nil
}

// GetAssignmentsWithSubmissions returns all course assignments
// of requested type with preloaded submissions.
func (db *GormDB) GetAssignmentsWithSubmissions(courseID uint64, submissionType qf.SubmissionsForCourseRequest_Type, withBuildInfo bool) ([]*qf.Assignment, error) {
	var assignments []*qf.Assignment
	// the 'order' field of qf.Assignment must be in 'quotes' since otherwise it will be interpreted as SQL
	m := db.conn.Preload("Submissions").
		Preload("Submissions.Grades").
		Preload("Submissions.Reviews").
		Preload("Submissions.Reviews.GradingBenchmarks").
		Preload("Submissions.Reviews.GradingBenchmarks.Criteria").
		Preload("Submissions.Scores")
	if withBuildInfo {
		m.Preload("Submissions.BuildInfo")
	}
	if err := m.Where(&qf.Assignment{CourseID: courseID}).
		Order("'order'").
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	if submissionType == qf.SubmissionsForCourseRequest_ALL {
		return assignments, nil
	}
	wantGroupLabs := submissionType == qf.SubmissionsForCourseRequest_GROUP
	filteredAssignments := make([]*qf.Assignment, 0)
	for _, a := range assignments {
		if a.IsGroupLab == wantGroupLabs {
			filteredAssignments = append(filteredAssignments, a)
		}
	}
	return filteredAssignments, nil
}

// CreateBenchmark creates a new grading benchmark
func (db *GormDB) CreateBenchmark(query *qf.GradingBenchmark) error {
	return db.conn.Create(query).Error
}

// UpdateBenchmark updates the given benchmark
func (db *GormDB) UpdateBenchmark(query *qf.GradingBenchmark) error {
	return db.conn.
		Where(&qf.GradingBenchmark{
			ID:           query.ID,
			AssignmentID: query.AssignmentID,
			ReviewID:     query.ReviewID,
		}).
		Updates(&qf.GradingBenchmark{
			Heading: query.Heading,
			Comment: query.Comment,
		}).Error
}

// DeleteBenchmark removes the given benchmark
func (db *GormDB) DeleteBenchmark(query *qf.GradingBenchmark) error {
	db.conn.Where("benchmark_id = ?", query.GetID()).Delete(&qf.GradingCriterion{})
	return db.conn.Delete(query).Error
}

// CreateCriterion creates a new grading criterion
func (db *GormDB) CreateCriterion(query *qf.GradingCriterion) error {
	return db.conn.Create(query).Error
}

// UpdateCriterion updates the given criterion
func (db *GormDB) UpdateCriterion(query *qf.GradingCriterion) error {
	return db.conn.
		Where(&qf.GradingCriterion{ID: query.ID, BenchmarkID: query.BenchmarkID}).
		Updates(&qf.GradingCriterion{Description: query.Description, Comment: query.Comment, Grade: query.Grade, Points: query.Points}).Error
}

// DeleteCriterion removes the given criterion
func (db *GormDB) DeleteCriterion(query *qf.GradingCriterion) error {
	return db.conn.Delete(query).Error
}

// GetBenchmarks returns all benchmarks and associated criteria for a given assignment ID
func (db *GormDB) GetBenchmarks(query *qf.Assignment) ([]*qf.GradingBenchmark, error) {
	var benchmarks []*qf.GradingBenchmark

	var assignment qf.Assignment
	if err := db.conn.Where(query).
		First(&assignment).Error; err != nil {
		return nil, err
	}
	if err := db.conn.
		Where("assignment_id = ?", assignment.ID).
		Where("review_id = ?", 0).
		Find(&benchmarks).Error; err != nil {
		return nil, err
	}

	for _, b := range benchmarks {
		var criteria []*qf.GradingCriterion
		if err := db.conn.Where("benchmark_id = ?", b.ID).Find(&criteria).Error; err != nil {
			return nil, err
		}
		b.Criteria = criteria
	}
	return benchmarks, nil
}
