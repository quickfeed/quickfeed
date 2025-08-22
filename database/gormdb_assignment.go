package database

import (
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

var ErrInvalidCourseRelation = errors.New("entity belongs to a different course")

// CreateAssignment creates a new assignment record.
func (db *GormDB) CreateAssignment(assignment *qf.Assignment) error {
	// Course id and assignment order must be given.
	if assignment.GetCourseID() < 1 || assignment.GetOrder() < 1 {
		return gorm.ErrRecordNotFound
	}

	var course int64
	if err := db.conn.Model(&qf.Course{}).Where(&qf.Course{
		ID: assignment.GetCourseID(),
	}).Count(&course).Error; err != nil {
		return err
	}

	if course != 1 {
		return gorm.ErrRecordNotFound
	}
	return db.conn.
		Where(qf.Assignment{
			CourseID: assignment.GetCourseID(),
			Order:    assignment.GetOrder(),
		}).
		Assign(map[string]interface{}{
			"name":              assignment.GetName(),
			"order":             assignment.GetOrder(),
			"deadline":          assignment.GetDeadline().AsTime(),
			"auto_approve":      assignment.GetAutoApprove(),
			"score_limit":       assignment.GetScoreLimit(),
			"is_group_lab":      assignment.GetIsGroupLab(),
			"reviewers":         assignment.GetReviewers(),
			"container_timeout": assignment.GetContainerTimeout(),
			"tasks":             assignment.GetTasks(),
		}).Omit("Tasks").FirstOrCreate(assignment).Error
}

// GetAssignment returns assignment with the given ID.
func (db *GormDB) GetAssignment(query *qf.Assignment) (*qf.Assignment, error) {
	var assignment qf.Assignment
	if err := db.conn.Where(query).
		Preload("ExpectedTests").
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
	if err := db.conn.
		Preload("Assignments").
		Preload("Assignments.ExpectedTests").
		First(&course, courseID).Error; err != nil {
		return nil, err
	}
	for _, a := range course.GetAssignments() {
		a.GradingBenchmarks, err = db.GetBenchmarks(&qf.Assignment{ID: a.GetID()})
		if err != nil {
			return nil, err
		}
	}
	return course.GetAssignments(), nil
}

// UpdateAssignments updates assignment information.
func (db *GormDB) UpdateAssignments(assignments []*qf.Assignment) error {
	var errs error
	for _, v := range assignments {
		err := db.conn.Transaction(func(tx *gorm.DB) error {
			if err := check(tx, v); err != nil {
				return err
			}

			var assignment qf.Assignment
			if tx.Model(&qf.Assignment{}).FirstOrInit(&assignment,
				&qf.Assignment{
					CourseID: v.GetCourseID(),
					Order:    v.GetOrder(),
				},
			).RowsAffected == 0 {
				// Zero rows affected indicates that the assignment does not exist
				return tx.Model(&qf.Assignment{}).Create(v).Error
			}

			// Assign the existing assignment ID to the incoming assignment
			v.ID = assignment.GetID()
			if err := db.updateGradingCriteria(tx, v); err != nil {
				return err // will rollback transaction
			}
			// This sets the assignment ID (and ID if it already exists) for each expected test.
			// This is required to avoid duplicates in the database.
			for _, info := range v.GetExpectedTests() {
				if err := tx.Model(&qf.TestInfo{}).Where(&qf.TestInfo{
					AssignmentID: v.GetID(),
					TestName:     info.GetTestName(),
				}).FirstOrInit(info).Error; err != nil {
					return err // will rollback transaction
				}
			}

			if err := tx.Model(v).Where(&qf.Assignment{
				ID: assignment.GetID(),
			}).Select("*").Updates(&qf.Assignment{
				ID:               v.GetID(),
				CourseID:         v.GetCourseID(),
				Name:             v.GetName(),
				Deadline:         v.GetDeadline(),
				AutoApprove:      v.GetAutoApprove(),
				Order:            v.GetOrder(),
				IsGroupLab:       v.GetIsGroupLab(),
				ScoreLimit:       v.GetScoreLimit(),
				Reviewers:        v.GetReviewers(),
				ContainerTimeout: v.GetContainerTimeout(),
				// Submissions:       v.GetSubmissions(),
				Tasks:             v.GetTasks(),
				GradingBenchmarks: v.GetGradingBenchmarks(),
				ExpectedTests:     v.GetExpectedTests(),
			}).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func check(tx *gorm.DB, assignment *qf.Assignment) error {
	// Course id and assignment order must be given.
	if assignment.GetCourseID() < 1 || assignment.GetOrder() < 1 {
		return gorm.ErrRecordNotFound
	}
	var course int64
	if err := tx.Model(&qf.Course{}).Where(&qf.Course{
		ID: assignment.GetCourseID(),
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
			ID: assignment.GetID(),
		})
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// a new assignment, no actions required
				return nil
			}
			return fmt.Errorf("failed to fetch assignment %s from database: %w", assignment.GetName(), err)
		}
		if len(gradingBenchmarks) > 0 {
			if cmp.Equal(assignment.GetGradingBenchmarks(), gradingBenchmarks, cmp.Options{
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
					for _, c := range bm.GetCriteria() {
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

// CreateBenchmark creates a new grading benchmark
func (db *GormDB) CreateBenchmark(query *qf.GradingBenchmark) error {
	if _, err := db.GetAssignment(&qf.Assignment{
		ID: query.GetAssignmentID(),
	}); err != nil {
		return err
	}
	return db.conn.Create(query).Error
}

// getBenchmark fetches a benchmark by its ID
func (db *GormDB) getBenchmark(benchmarkID uint64) (*qf.GradingBenchmark, error) {
	var benchmark qf.GradingBenchmark
	if err := db.conn.First(&benchmark, benchmarkID).Error; err != nil {
		return nil, err
	}
	return &benchmark, nil
}

// UpdateBenchmark updates the given benchmark
func (db *GormDB) UpdateBenchmark(query *qf.GradingBenchmark) error {
	return db.conn.Select("*").
		Where(&qf.GradingBenchmark{
			ID:           query.GetID(),
			AssignmentID: query.GetAssignmentID(),
			ReviewID:     query.GetReviewID(),
		}).Updates(query).Error
}

// DeleteBenchmark removes the given benchmark
func (db *GormDB) DeleteBenchmark(query *qf.GradingBenchmark) error {
	db.conn.Where("benchmark_id = ?", query.GetID()).Delete(&qf.GradingCriterion{})
	return db.conn.Delete(query).Error
}

// CreateCriterion creates a new grading criterion
func (db *GormDB) CreateCriterion(query *qf.GradingCriterion) error {
	// check that the given criterion's benchmark exists
	benchmark, err := db.getBenchmark(query.GetBenchmarkID())
	if err != nil {
		return err
	}
	// check that the given criterion's course belongs to the corresponding benchmark
	if benchmark.GetCourseID() != query.GetCourseID() {
		return ErrInvalidCourseRelation
	}
	return db.conn.Create(query).Error
}

// UpdateCriterion updates the given criterion
func (db *GormDB) UpdateCriterion(query *qf.GradingCriterion) error {
	return db.conn.Select("*").
		Where(&qf.GradingCriterion{
			ID:          query.GetID(),
			BenchmarkID: query.GetBenchmarkID(),
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
		Where("assignment_id = ?", assignment.GetID()).
		Where("review_id = ?", 0).
		Find(&benchmarks).Error; err != nil {
		return nil, err
	}

	for _, b := range benchmarks {
		var criteria []*qf.GradingCriterion
		if err := db.conn.Where("benchmark_id = ?", b.GetID()).Find(&criteria).Error; err != nil {
			return nil, err
		}
		b.Criteria = criteria
	}
	return benchmarks, nil
}
