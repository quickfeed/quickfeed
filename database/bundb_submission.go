package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateSubmission creates a new submission record or updates the most
// recent submission, as defined by the provided submission query.
func (db *BunDB) CreateSubmission(submission *qf.Submission) error {
	if err := db.checkSubmission(submission); err != nil {
		return err
	}
	ctx := context.Background()
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if submission.GetID() != 0 {
			exists, err := tx.NewSelect().Model((*qf.Submission)(nil)).
				Where("id = ?", submission.GetID()).Exists(ctx)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("submission %d not found", submission.GetID())
			}
			if _, err := tx.NewDelete().Model((*score.Score)(nil)).
				Where("submission_id = ?", submission.GetID()).Exec(ctx); err != nil {
				return err
			}
			if _, err := tx.NewDelete().Model((*score.BuildInfo)(nil)).
				Where("submission_id = ?", submission.GetID()).Exec(ctx); err != nil {
				return err
			}
			if submission.GetBuildInfo() != nil {
				submission.BuildInfo.SubmissionID = submission.GetID()
			}
			for _, sc := range submission.GetScores() {
				sc.SubmissionID = submission.GetID()
			}
		} else {
			if err := bunSetGrades(ctx, tx, submission); err != nil {
				return err
			}
		}
		_, err := tx.NewInsert().Model(submission).
			On("CONFLICT (id) DO UPDATE").
			Set("score = EXCLUDED.score, status = EXCLUDED.status").
			Exec(ctx)
		if err != nil {
			return err
		}
		// Bun doesn't automatically save associations, so insert grades explicitly
		if len(submission.GetGrades()) > 0 {
			for _, grade := range submission.GetGrades() {
				grade.SubmissionID = submission.GetID()
			}
			_, err = tx.NewInsert().Model(&submission.Grades).
				On("CONFLICT (submission_id, user_id) DO UPDATE").
				Set("status = EXCLUDED.status").
				Exec(ctx)
			if err != nil {
				return err
			}
		}
		if submission.GetBuildInfo() != nil {
			submission.BuildInfo.ID = 0
			submission.BuildInfo.SubmissionID = submission.GetID()
			if _, err = tx.NewInsert().Model(submission.BuildInfo).Exec(ctx); err != nil {
				return err
			}
		}
		if len(submission.GetScores()) > 0 {
			for _, sc := range submission.GetScores() {
				sc.ID = 0
				sc.SubmissionID = submission.GetID()
			}
			if _, err = tx.NewInsert().Model(&submission.Scores).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// bunSetGrades initializes grade records for a new submission.
func bunSetGrades(ctx context.Context, tx bun.Tx, submission *qf.Submission) error {
	var userIDs []uint64
	if submission.GetUserID() > 0 {
		userIDs = []uint64{submission.GetUserID()}
	}
	if submission.GetGroupID() > 0 {
		if err := tx.NewSelect().Model((*qf.Enrollment)(nil)).
			Column("user_id").
			Where("group_id = ?", submission.GetGroupID()).
			Scan(ctx, &userIDs); err != nil {
			return err
		}
	}
	if submission.GetGrades() == nil {
		submission.Grades = make([]*qf.Grade, len(userIDs))
		for i, userID := range userIDs {
			submission.Grades[i] = &qf.Grade{UserID: userID}
		}
	}
	var assignment qf.Assignment
	if err := tx.NewSelect().Model(&assignment).
		Where("id = ?", submission.GetAssignmentID()).Scan(ctx); err != nil {
		return err
	}
	submission.SetGradesIfApproved(&assignment, submission.GetScore())
	return nil
}

// checkSubmission returns an error if the submission is invalid.
func (db *BunDB) checkSubmission(submission *qf.Submission) error {
	ctx := context.Background()
	if submission.GetAssignmentID() < 1 {
		return ErrInvalidAssignmentID
	}
	switch {
	case submission.GetUserID() > 0 && submission.GetGroupID() > 0:
		return ErrInvalidSubmission
	case submission.GetUserID() > 0:
		exists, err := db.conn.NewSelect().Model((*qf.User)(nil)).
			Where("id = ?", submission.GetUserID()).Exists(ctx)
		if err != nil {
			return fmt.Errorf("user %d not found for submission: %w", submission.GetUserID(), err)
		}
		if !exists {
			return sql.ErrNoRows
		}
	case submission.GetGroupID() > 0:
		exists, err := db.conn.NewSelect().Model((*qf.Group)(nil)).
			Where("id = ?", submission.GetGroupID()).Exists(ctx)
		if err != nil {
			return fmt.Errorf("group %d not found for submission: %w", submission.GetGroupID(), err)
		}
		if !exists {
			return sql.ErrNoRows
		}
	default:
		return ErrInvalidSubmission
	}
	exists, err := db.conn.NewSelect().Model((*qf.Assignment)(nil)).
		Where("id = ?", submission.GetAssignmentID()).Exists(ctx)
	if err != nil {
		return fmt.Errorf("assignment %d not found: %w", submission.GetAssignmentID(), err)
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

// GetSubmission fetches a submission record matching the given query.
func (db *BunDB) GetSubmission(query *qf.Submission) (*qf.Submission, error) {
	ctx := context.Background()
	var submission qf.Submission
	q := db.conn.NewSelect().
		Model(&submission).
		Relation("Reviews").
		Relation("BuildInfo").
		Relation("Scores").
		Relation("Grades").
		Relation("Reviews.GradingBenchmarks").
		Relation("Reviews.GradingBenchmarks.Criteria")
	if query.GetID() > 0 {
		q = q.Where("submission.id = ?", query.GetID())
	}
	if query.GetAssignmentID() > 0 {
		q = q.Where("submission.assignment_id = ?", query.GetAssignmentID())
	}
	if query.GetUserID() > 0 {
		q = q.Where("submission.user_id = ?", query.GetUserID())
	}
	if query.GetGroupID() > 0 {
		q = q.Where("submission.group_id = ?", query.GetGroupID())
	}
	if err := q.OrderExpr("submission.id DESC").Limit(1).Scan(ctx); err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetLastSubmission returns the last submission for the given query and course ID.
func (db *BunDB) GetLastSubmission(courseID uint64, query *qf.Submission) (*qf.Submission, error) {
	submission, err := db.GetSubmission(query)
	if err != nil {
		return nil, err
	}
	exists, err := db.conn.NewSelect().Model((*qf.Assignment)(nil)).
		Where("id = ? AND course_id = ?", submission.GetAssignmentID(), courseID).
		Exists(context.Background())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("assignment %d not found for course %d", submission.GetAssignmentID(), courseID)
	}
	return submission, nil
}

// GetLastSubmissions returns all latest submissions for the given course and query.
func (db *BunDB) GetLastSubmissions(courseID uint64, query *qf.Submission) ([]*qf.Submission, error) {
	ctx := context.Background()
	var course qf.Course
	if err := db.conn.NewSelect().
		Model(&course).
		Relation("Assignments").
		Where("course.id = ?", courseID).
		Scan(ctx); err != nil {
		return nil, err
	}
	var latestSubs []*qf.Submission
	for _, a := range course.GetAssignments() {
		query.AssignmentID = a.GetID()
		temp, err := db.GetSubmission(query)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, err
		}
		latestSubs = append(latestSubs, temp)
	}
	return latestSubs, nil
}

// GetSubmissions returns all submissions matching the query.
func (db *BunDB) GetSubmissions(query *qf.Submission) ([]*qf.Submission, error) {
	if _, err := db.GetAssignment(&qf.Assignment{ID: query.GetAssignmentID()}); err != nil {
		return nil, err
	}
	ctx := context.Background()
	var submissions []*qf.Submission
	q := db.conn.NewSelect().Model(&submissions).Relation("Grades")
	if query.GetAssignmentID() > 0 {
		q = q.Where("submission.assignment_id = ?", query.GetAssignmentID())
	}
	if query.GetUserID() > 0 {
		q = q.Where("submission.user_id = ?", query.GetUserID())
	}
	if query.GetGroupID() > 0 {
		q = q.Where("submission.group_id = ?", query.GetGroupID())
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return submissions, nil
}

// UpdateSubmission updates submission with the given approved status.
func (db *BunDB) UpdateSubmission(query *qf.Submission) error {
	ctx := context.Background()
	// Update submission fields
	if _, err := db.conn.NewUpdate().Model(query).WherePK().Exec(ctx); err != nil {
		return err
	}
	// Update associated grades (similar to gorm's FullSaveAssociations)
	if len(query.GetGrades()) > 0 {
		for _, grade := range query.GetGrades() {
			if _, err := db.conn.NewUpdate().Model(grade).
				Where("submission_id = ? AND user_id = ?", grade.GetSubmissionID(), grade.GetUserID()).
				Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetReview fetches a review matching the given query.
func (db *BunDB) GetReview(query *qf.Review) (*qf.Review, error) {
	ctx := context.Background()
	var review qf.Review
	q := db.conn.NewSelect().
		Model(&review).
		Relation("GradingBenchmarks").
		Relation("GradingBenchmarks.Criteria")
	if query.GetID() > 0 {
		q = q.Where("review.id = ?", query.GetID())
	}
	if query.GetSubmissionID() > 0 {
		q = q.Where("review.submission_id = ?", query.GetSubmissionID())
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return &review, nil
}

// CreateReview creates a new submission review.
func (db *BunDB) CreateReview(query *qf.Review) error {
	submission, err := db.GetSubmission(&qf.Submission{ID: query.GetSubmissionID()})
	if err != nil {
		return err
	}
	assignment, err := db.GetAssignment(&qf.Assignment{ID: submission.GetAssignmentID()})
	if err != nil {
		return err
	}
	if len(submission.GetReviews()) >= int(assignment.GetReviewers()) {
		return ErrAllReviewsCreated(submission.GetID(), assignment.GetName(), assignment.GetReviewers())
	}
	query.Edited = timestamppb.Now()
	query.ComputeScore()
	benchmarks, err := db.GetBenchmarks(&qf.Assignment{ID: submission.GetAssignmentID()})
	if err != nil {
		return err
	}
	query.GradingBenchmarks = benchmarks
	for _, bm := range query.GetGradingBenchmarks() {
		bm.ID = 0
		for _, c := range bm.GetCriteria() {
			c.ID = 0
		}
	}
	ctx := context.Background()
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err = tx.NewInsert().Model(query).Exec(ctx); err != nil {
			return err
		}
		return db.bunReplaceReviewBenchmarks(ctx, tx, query)
	})
}

// UpdateReview updates a review.
func (db *BunDB) UpdateReview(query *qf.Review) error {
	if query.GetID() == 0 {
		return ErrEmptyReviewID
	}
	submission, err := db.GetSubmission(&qf.Submission{ID: query.GetSubmissionID()})
	if err != nil {
		return err
	}
	query.Edited = timestamppb.Now()
	query.ComputeScore()
	ctx := context.Background()
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewUpdate().Model(query).WherePK().Exec(ctx); err != nil {
			return err
		}
		if err := db.bunReplaceReviewBenchmarks(ctx, tx, query); err != nil {
			return err
		}
		submission.Score = query.GetScore()
		if _, err := tx.NewUpdate().Model(submission).
			Column("score").
			WherePK().
			Exec(ctx); err != nil {
			return err
		}
		return nil
	})
}

func (db *BunDB) bunReplaceReviewBenchmarks(ctx context.Context, tx bun.Tx, review *qf.Review) error {
	if _, err := tx.NewDelete().Model((*qf.GradingCriterion)(nil)).
		Where("benchmark_id IN (SELECT id FROM grading_benchmarks WHERE review_id = ?)", review.GetID()).
		Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.NewDelete().Model((*qf.GradingBenchmark)(nil)).
		Where("review_id = ?", review.GetID()).
		Exec(ctx); err != nil {
		return err
	}
	for _, benchmark := range review.GetGradingBenchmarks() {
		benchmark.ID = 0
		benchmark.ReviewID = review.GetID()
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

// DeleteReview removes all reviews matching the query.
func (db *BunDB) DeleteReview(query *qf.Review) error {
	ctx := context.Background()
	q := db.conn.NewDelete().Model((*qf.Review)(nil))
	if query.GetID() > 0 {
		q = q.Where("id = ?", query.GetID())
	}
	if query.GetSubmissionID() > 0 {
		q = q.Where("submission_id = ?", query.GetSubmissionID())
	}
	_, err := q.Exec(ctx)
	return err
}
