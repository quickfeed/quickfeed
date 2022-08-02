package database

import (
	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

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
	return db.conn.Where(&qf.Assignment{
		CourseID: assignment.CourseID,
		Order:    assignment.Order,
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
func (db *GormDB) GetAssignmentsByCourse(courseID uint64, withBenchmarkTemplate bool) (_ []*qf.Assignment, err error) {
	var course qf.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}
	if withBenchmarkTemplate {
		for _, a := range course.Assignments {
			a.GradingBenchmarks, err = db.GetBenchmarks(&qf.Assignment{ID: a.ID})
			if err != nil {
				return nil, err
			}
		}
	}
	return course.Assignments, nil
}

// UpdateAssignments updates assignment information.
func (db *GormDB) UpdateAssignments(assignments []*qf.Assignment) error {
	return db.conn.Transaction(func(tx *gorm.DB) error {
		for _, v := range assignments {
			if err := tx.Model(v).Select("*").
				Where(&qf.Assignment{
					ID: v.ID,
				}).
				Updates(&qf.Assignment{
					ID:               v.ID,
					CourseID:         v.CourseID,
					Name:             v.Name,
					RunScriptContent: v.RunScriptContent,
					Deadline:         v.Deadline,
					AutoApprove:      v.AutoApprove,
					Order:            v.Order,
					IsGroupLab:       v.IsGroupLab,
					ScoreLimit:       v.ScoreLimit,
					Reviewers:        v.Reviewers,
					ContainerTimeout: v.ContainerTimeout,
					// Submissions:       v.Submissions,
					// Tasks:             v.Tasks,
					// GradingBenchmarks: v.GradingBenchmarks,
				}).Error; err != nil {
				return err // will rollback transaction
			}
		}
		return nil
	})
}

// GetAssignmentsWithSubmissions returns all course assignments
// of requested type with preloaded submissions.
func (db *GormDB) GetAssignmentsWithSubmissions(courseID uint64, submissionType qf.SubmissionsForCourseRequest_Type, withBuildInfo bool) ([]*qf.Assignment, error) {
	m := db.conn.Preload("Submissions").
		Preload("Submissions.Reviews").
		Preload("Submissions.Reviews.GradingBenchmarks").
		Preload("Submissions.Reviews.GradingBenchmarks.Criteria").
		Preload("Submissions.Scores")
	if withBuildInfo {
		m.Preload("Submissions.BuildInfo")
	}
	// the 'order' field must be in 'quotes', otherwise it will be interpreted as SQL.
	var assignments []*qf.Assignment
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
		Where(&qf.GradingCriterion{
			ID:          query.ID,
			BenchmarkID: query.BenchmarkID,
		}).
		Updates(&qf.GradingCriterion{
			Description: query.Description,
			Comment:     query.Comment,
			Grade:       query.Grade,
			Points:      query.Points,
		}).
		Error
}

// DeleteCriterion removes the given criterion
func (db *GormDB) DeleteCriterion(query *qf.GradingCriterion) error {
	return db.conn.Delete(query).Error
}

// GetBenchmarks returns all benchmarks and associated criteria without reviews for a given assignment ID.
func (db *GormDB) GetBenchmarks(query *qf.Assignment) ([]*qf.GradingBenchmark, error) {
	var benchmarks []*qf.GradingBenchmark
	err := db.conn.Transaction(func(tx *gorm.DB) error {
		var assignment qf.Assignment
		if err := tx.Where(query).First(&assignment).Error; err != nil {
			return err // will rollback transaction
		}
		// SELECT * FROM grading_benchmarks WHERE assignment_id = 1 AND review_id = 0
		// Note that review_id = 0 ensures that only benchmarks without reviews are returned.
		if err := tx.Where(&qf.GradingBenchmark{
			AssignmentID: assignment.ID,
			ReviewID:     0,
		}, "assignment_id", "review_id").Find(&benchmarks).Error; err != nil {
			return err // will rollback transaction
		}
		for _, b := range benchmarks {
			var criteria []*qf.GradingCriterion
			if err := tx.Where(&qf.GradingCriterion{
				BenchmarkID: b.ID,
			}).Find(&criteria).Error; err != nil {
				return err // will rollback transaction
			}
			b.Criteria = criteria
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return benchmarks, nil
}
