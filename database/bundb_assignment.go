package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/testing/protocmp"
)

// CreateAssignment creates a new assignment record or updates an existing one.
func (db *BunDB) CreateAssignment(assignment *qf.Assignment) error {
	ctx := context.Background()
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := bunCheckAssignment(ctx, tx, assignment); err != nil {
			return err
		}
		var existing qf.Assignment
		err := tx.NewSelect().Model(&existing).
			Where("course_id = ? AND \"order\" = ?", assignment.GetCourseID(), assignment.GetOrder()).
			Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if _, createErr := tx.NewInsert().Model(assignment).Exec(ctx); createErr != nil {
					return createErr
				}
				return db.bunSaveAssignmentAssociations(ctx, tx, assignment)
			}
			return err
		}
		assignment.ID = existing.GetID()
		if _, err = tx.NewUpdate().Model(assignment).WherePK().Exec(ctx); err != nil {
			return err
		}
		if len(assignment.GetGradingBenchmarks()) > 0 || len(assignment.GetExpectedTests()) > 0 {
			if err := db.bunUpdateGradingCriteria(ctx, tx, assignment); err != nil {
				return err
			}
			if err := db.bunUpdateExpectedTests(ctx, tx, assignment); err != nil {
				return err
			}
		}
		return db.bunSaveAssignmentAssociations(ctx, tx, assignment)
	})
}

// GetAssignment returns the assignment matching the given query.
func (db *BunDB) GetAssignment(query *qf.Assignment) (*qf.Assignment, error) {
	ctx := context.Background()
	var assignment qf.Assignment
	q := db.conn.NewSelect().
		Model(&assignment).
		Relation("ExpectedTests").
		Relation("GradingBenchmarks").
		Relation("GradingBenchmarks.Criteria")
	if query.GetID() > 0 {
		q = q.Where("assignment.id = ?", query.GetID())
	}
	if query.GetCourseID() > 0 {
		q = q.Where("assignment.course_id = ?", query.GetCourseID())
	}
	if query.GetOrder() > 0 {
		q = q.Where("assignment.\"order\" = ?", query.GetOrder())
	}
	if query.GetName() != "" {
		q = q.Where("assignment.name = ?", query.GetName())
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetAssignmentsByCourse fetches all assignments for the given course ID.
func (db *BunDB) GetAssignmentsByCourse(courseID uint64) ([]*qf.Assignment, error) {
	ctx := context.Background()
	var course qf.Course
	if err := db.conn.NewSelect().
		Model(&course).
		Relation("Assignments").
		Relation("Assignments.ExpectedTests").
		Where("course.id = ?", courseID).
		Scan(ctx); err != nil {
		return nil, err
	}
	var err error
	for _, a := range course.GetAssignments() {
		a.GradingBenchmarks, err = db.GetBenchmarks(&qf.Assignment{ID: a.GetID()})
		if err != nil {
			return nil, err
		}
	}
	return course.GetAssignments(), nil
}

// UpdateAssignments updates the specified list of assignments.
func (db *BunDB) UpdateAssignments(assignments []*qf.Assignment) error {
	ctx := context.Background()
	var errs error
	for _, v := range assignments {
		err := db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			if err := bunCheckAssignment(ctx, tx, v); err != nil {
				return err
			}

			var existing qf.Assignment
			err := tx.NewSelect().Model(&existing).
				Where("course_id = ? AND \"order\" = ?", v.GetCourseID(), v.GetOrder()).
				Scan(ctx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					_, createErr := tx.NewInsert().Model(v).Exec(ctx)
					return createErr
				}
				return err
			}

			v.ID = existing.GetID()
			if err := db.bunUpdateGradingCriteria(ctx, tx, v); err != nil {
				return err
			}
			if err := db.bunUpdateExpectedTests(ctx, tx, v); err != nil {
				return err
			}

			if _, err = tx.NewUpdate().Model(v).WherePK().Exec(ctx); err != nil {
				return err
			}
			return db.bunSaveAssignmentAssociations(ctx, tx, v)
		})
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

// bunSaveAssignmentAssociations explicitly persists child records that gorm handled via implicit association saves.
func (db *BunDB) bunSaveAssignmentAssociations(ctx context.Context, tx bun.Tx, assignment *qf.Assignment) error {
	for _, test := range assignment.GetExpectedTests() {
		test.ID = 0
		test.AssignmentID = assignment.GetID()
		if _, err := tx.NewInsert().Model(test).Exec(ctx); err != nil {
			return err
		}
	}

	for _, benchmark := range assignment.GetGradingBenchmarks() {
		benchmark.ID = 0
		benchmark.AssignmentID = assignment.GetID()
		if _, err := tx.NewInsert().Model(benchmark).Exec(ctx); err != nil {
			return err
		}
		for _, criterion := range benchmark.GetCriteria() {
			criterion.ID = 0
			criterion.BenchmarkID = benchmark.GetID()
			if _, err := tx.NewInsert().Model(criterion).Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// bunCheckAssignment validates that course ID and order are present, and that the course exists.
func bunCheckAssignment(ctx context.Context, tx bun.Tx, assignment *qf.Assignment) error {
	if assignment.GetCourseID() < 1 || assignment.GetOrder() < 1 {
		return sql.ErrNoRows
	}
	exists, err := tx.NewSelect().Model((*qf.Course)(nil)).Where("id = ?", assignment.GetCourseID()).Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

// bunUpdateExpectedTests removes old expected tests when the assignment is updated.
func (db *BunDB) bunUpdateExpectedTests(ctx context.Context, tx bun.Tx, assignment *qf.Assignment) error {
	if len(assignment.GetExpectedTests()) == 0 {
		return nil
	}
	var expectedTests []*qf.TestInfo
	if err := tx.NewSelect().Model(&expectedTests).
		Where("assignment_id = ?", assignment.GetID()).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to fetch assignment %s from database: %w", assignment.GetName(), err)
	}
	for _, test := range expectedTests {
		if _, err := tx.NewDelete().Model(test).WherePK().Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete expected test %d: %w", test.GetID(), err)
		}
	}
	return nil
}

// bunUpdateGradingCriteria removes old grading criteria and reviews when criteria.json is updated.
func (db *BunDB) bunUpdateGradingCriteria(ctx context.Context, tx bun.Tx, assignment *qf.Assignment) error {
	if len(assignment.GetGradingBenchmarks()) == 0 {
		return nil
	}
	gradingBenchmarks, err := db.GetBenchmarks(&qf.Assignment{ID: assignment.GetID()})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to fetch assignment %s from database: %w", assignment.GetName(), err)
	}
	if len(gradingBenchmarks) == 0 {
		return nil
	}
	if cmp.Equal(assignment.GetGradingBenchmarks(), gradingBenchmarks, cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(&qf.GradingBenchmark{}, "ID", "AssignmentID", "ReviewID"),
		protocmp.IgnoreFields(&qf.GradingCriterion{}, "ID", "BenchmarkID"),
		protocmp.IgnoreEnums(),
	}) {
		assignment.GradingBenchmarks = nil
		return nil
	}
	for _, bm := range gradingBenchmarks {
		for _, c := range bm.GetCriteria() {
			if _, err := tx.NewDelete().Model(c).WherePK().Exec(ctx); err != nil {
				return fmt.Errorf("failed to delete criterion %d: %w", c.GetID(), err)
			}
		}
		if _, err := tx.NewDelete().Model(bm).WherePK().Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete benchmark %d: %w", bm.GetID(), err)
		}
	}
	return nil
}

// CreateBenchmark creates a new grading benchmark.
func (db *BunDB) CreateBenchmark(query *qf.GradingBenchmark) error {
	ctx := context.Background()
	if _, err := db.GetAssignment(&qf.Assignment{ID: query.GetAssignmentID()}); err != nil {
		return err
	}
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(query).Exec(ctx); err != nil {
			return err
		}
		for _, criterion := range query.GetCriteria() {
			criterion.ID = 0
			criterion.BenchmarkID = query.GetID()
			if _, err := tx.NewInsert().Model(criterion).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBenchmark updates the given benchmark.
func (db *BunDB) UpdateBenchmark(query *qf.GradingBenchmark) error {
	ctx := context.Background()
	_, err := db.conn.NewUpdate().Model(query).
		Where("id = ? AND assignment_id = ? AND review_id = ?",
			query.GetID(), query.GetAssignmentID(), query.GetReviewID()).
		Exec(ctx)
	return err
}

// DeleteBenchmark removes the given benchmark and all its criteria.
func (db *BunDB) DeleteBenchmark(query *qf.GradingBenchmark) error {
	ctx := context.Background()
	if _, err := db.conn.NewDelete().Model((*qf.GradingCriterion)(nil)).
		Where("benchmark_id = ?", query.GetID()).Exec(ctx); err != nil {
		return err
	}
	_, err := db.conn.NewDelete().Model(query).WherePK().Exec(ctx)
	return err
}

// CreateCriterion creates a new grading criterion.
func (db *BunDB) CreateCriterion(query *qf.GradingCriterion) error {
	ctx := context.Background()
	var benchmark qf.GradingBenchmark
	if err := db.conn.NewSelect().Model(&benchmark).Where("id = ?", query.GetBenchmarkID()).Scan(ctx); err != nil {
		return err
	}
	if benchmark.GetCourseID() != query.GetCourseID() {
		return ErrInvalidCourseRelation
	}
	_, err := db.conn.NewInsert().Model(query).Exec(ctx)
	return err
}

// UpdateCriterion updates the given criterion.
func (db *BunDB) UpdateCriterion(query *qf.GradingCriterion) error {
	ctx := context.Background()
	_, err := db.conn.NewUpdate().Model(query).
		Where("id = ? AND benchmark_id = ?", query.GetID(), query.GetBenchmarkID()).
		Exec(ctx)
	return err
}

// DeleteCriterion removes the given criterion.
func (db *BunDB) DeleteCriterion(query *qf.GradingCriterion) error {
	ctx := context.Background()
	_, err := db.conn.NewDelete().Model(query).WherePK().Exec(ctx)
	return err
}

// GetBenchmarks returns all benchmarks and associated criteria for a given assignment.
func (db *BunDB) GetBenchmarks(query *qf.Assignment) ([]*qf.GradingBenchmark, error) {
	ctx := context.Background()
	var assignment qf.Assignment
	q := db.conn.NewSelect().Model(&assignment)
	if query.GetID() > 0 {
		q = q.Where("id = ?", query.GetID())
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	var benchmarks []*qf.GradingBenchmark
	if err := db.conn.NewSelect().Model(&benchmarks).
		Where("assignment_id = ? AND review_id = 0", assignment.GetID()).
		Relation("Criteria").
		Scan(ctx); err != nil {
		return nil, err
	}
	return benchmarks, nil
}
