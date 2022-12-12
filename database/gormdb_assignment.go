package database

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
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
			"name":              assignment.Name,
			"order":             assignment.Order,
			"deadline":          assignment.Deadline,
			"auto_approve":      assignment.AutoApprove,
			"score_limit":       assignment.ScoreLimit,
			"is_group_lab":      assignment.IsGroupLab,
			"reviewers":         assignment.Reviewers,
			"container_timeout": assignment.ContainerTimeout,
			"tasks":             assignment.Tasks,
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
func (db *GormDB) GetAssignmentsByCourse(courseID uint64) (_ []*qf.Assignment, err error) {
	var course qf.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}
	for _, a := range course.Assignments {
		a.GradingBenchmarks, err = db.GetBenchmarks(&qf.Assignment{ID: a.ID})
		if err != nil {
			return nil, err
		}
	}
	return course.Assignments, nil
}

// UpdateAssignments updates assignment information.
func (db *GormDB) UpdateAssignments(assignments []*qf.Assignment) error {
	return db.conn.Transaction(func(tx *gorm.DB) error {
		for _, v := range assignments {
			if err := check(tx, v); err != nil {
				return err
			}

			assignment := qf.Assignment{}
			if tx.Model(&qf.Assignment{}).FirstOrInit(&assignment,
				&qf.Assignment{
					CourseID: v.CourseID,
					Order:    v.Order,
				},
			).RowsAffected == 0 {
				// Zero rows affected indicates that the assignment does not exist
				return tx.Model(&qf.Assignment{}).Create(v).Error
			}

			// Assign the existing assignment ID to the incoming assignment
			v.ID = assignment.ID
			if err := db.updateGradingCriteria(tx, v); err != nil {
				return err // will rollback transaction
			}

			if err := tx.Model(v).Where(&qf.Assignment{
				ID: assignment.ID,
			}).Select("*").Updates(&qf.Assignment{
				ID:               v.ID,
				CourseID:         v.CourseID,
				Name:             v.Name,
				Deadline:         v.Deadline,
				AutoApprove:      v.AutoApprove,
				Order:            v.Order,
				IsGroupLab:       v.IsGroupLab,
				ScoreLimit:       v.ScoreLimit,
				Reviewers:        v.Reviewers,
				ContainerTimeout: v.ContainerTimeout,
				// Submissions:       v.Submissions,
				Tasks:             v.Tasks,
				GradingBenchmarks: v.GradingBenchmarks,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func check(tx *gorm.DB, assignment *qf.Assignment) error {
	// Course id and assignment order must be given.
	if assignment.CourseID < 1 || assignment.Order < 1 {
		return gorm.ErrRecordNotFound
	}
	var course int64
	if err := tx.Model(&qf.Course{}).Where(&qf.Course{
		ID: assignment.CourseID,
	}).Count(&course).Error; err != nil {
		return err
	}
	if course != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// updateGradingCriteria will remove old grading criteria and related reviews when criteria.json gets updated.
func (db *GormDB) updateGradingCriteria(tx *gorm.DB, assignment *qf.Assignment) error {
	if len(assignment.GetGradingBenchmarks()) > 0 {
		gradingBenchmarks, err := db.GetBenchmarks(&qf.Assignment{
			ID: assignment.ID,
		})
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// a new assignment, no actions required
				return nil
			}
			return fmt.Errorf("failed to fetch assignment %s from database: %w", assignment.Name, err)
		}
		if len(gradingBenchmarks) > 0 {
			if cmp.Equal(assignment.GradingBenchmarks, gradingBenchmarks, cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(&qf.GradingBenchmark{}, "ID", "AssignmentID", "ReviewID"),
				protocmp.IgnoreFields(&qf.GradingCriterion{}, "ID", "BenchmarkID"),
				protocmp.IgnoreEnums(),
			}) {
				// no changes in the grading criteria for this assignment (from the tests repository)
				// we set this to nil to avoid duplicates in the database
				assignment.GradingBenchmarks = nil
			} else {
				// grading criteria changed for this assignment, remove old criteria and reviews
				for _, bm := range gradingBenchmarks {
					for _, c := range bm.Criteria {
						if err := tx.Delete(c).Error; err != nil {
							return fmt.Errorf("failed to delete criterion %d: %w", c.GetID(), err)
						}
					}
					if err := tx.Delete(bm).Error; err != nil {
						return fmt.Errorf("failed to delete benchmark %d: %w", bm.GetID(), err)
					}
				}
			}
		}
	}
	return nil
}

// GetAssignmentsWithSubmissions returns all course assignments
// preloaded with submissions of the requested submission type.
func (db *GormDB) GetAssignmentsWithSubmissions(req *qf.SubmissionRequest) ([]*qf.Assignment, error) {
	var assignments []*qf.Assignment
	m := db.conn.Preload("Submissions").
		Preload("Submissions.Reviews").
		Preload("Submissions.Reviews.GradingBenchmarks").
		Preload("Submissions.Reviews.GradingBenchmarks.Criteria").
		Preload("Submissions.Scores")
	// the 'order' field of qf.Assignment must be in 'quotes' since otherwise it will be interpreted as SQL
	if err := m.Where(&qf.Assignment{CourseID: req.GetCourseID()}).
		Order("'order'").
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	if req.IncludeAll() {
		return assignments, nil
	}
	filteredAssignments := make([]*qf.Assignment, 0)
	for _, a := range assignments {
		if req.Include(a) {
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
	return db.conn.Select("*").
		Where(&qf.GradingBenchmark{
			ID:           query.ID,
			AssignmentID: query.AssignmentID,
			ReviewID:     query.ReviewID,
		}).Updates(query).Error
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
	return db.conn.Select("*").
		Where(&qf.GradingCriterion{
			ID:          query.ID,
			BenchmarkID: query.BenchmarkID,
		}).
		Updates(query).Error
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
