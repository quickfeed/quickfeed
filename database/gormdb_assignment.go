package database

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
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
			if err := db.updateGradingCriteria(v); err != nil {
				return err // will rollback transaction
			}
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

// updateGradingCriteria will remove old grading criteria and related reviews when criteria.json gets updated.
func (db *GormDB) updateGradingCriteria(assignment *qf.Assignment) error {
	if len(assignment.GetGradingBenchmarks()) > 0 {
		gradingBenchmarks, err := db.GetBenchmarks(&qf.Assignment{
			CourseID: assignment.CourseID,
			Order:    assignment.Order,
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
						if err := db.DeleteCriterion(c); err != nil {
							return fmt.Errorf("failed to delete criterion %d: %w", c.GetID(), err)
						}
					}
					if err := db.DeleteBenchmark(bm); err != nil {
						return fmt.Errorf("failed to delete benchmark %d: %w", bm.GetID(), err)
					}
				}
			}
		}
	}
	return nil
}

// GetAssignmentsWithSubmissions returns all course assignments
// of requested type with preloaded submissions.
func (db *GormDB) GetAssignmentsWithSubmissions(request *qf.SubmissionsForCourseRequest) ([]*qf.Assignment, error) {
	m := db.conn.Preload("Submissions").
		Preload("Submissions.Reviews").
		Preload("Submissions.Reviews.GradingBenchmarks").
		Preload("Submissions.Reviews.GradingBenchmarks.Criteria").
		Preload("Submissions.Scores")
	if request.GetWithBuildInfo() {
		m.Preload("Submissions.BuildInfo")
	}
	// the 'order' field must be in 'quotes', otherwise it will be interpreted as SQL.
	var assignments []*qf.Assignment
	if err := m.Where(&qf.Assignment{CourseID: request.GetCourseID()}).
		Order("'order'").
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	if request.IncludeAll() {
		return assignments, nil
	}
	filteredAssignments := make([]*qf.Assignment, 0)
	for _, a := range assignments {
		if request.Include(a) {
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

// GetBenchmarks returns all benchmarks and associated criteria without reviews for a given assignment.
func (db *GormDB) GetBenchmarks(query *qf.Assignment) ([]*qf.GradingBenchmark, error) {
	var benchmarks []*qf.GradingBenchmark
	err := db.conn.Transaction(func(tx *gorm.DB) error {
		// Lookup the assignment; may be based on e.g., CourseID and Order fields.
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
