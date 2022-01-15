package database

import (
	"errors"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/kit/score"
	"gorm.io/gorm"
)

var (
	// ErrInvalidSubmission is returned if the submission specify both UserID and GroupID or neither.
	ErrInvalidSubmission = errors.New("submission must specify exactly one of UserID or GroupID")
	// ErrInvalidAssignmentID is returned if assignment is not specified.
	ErrInvalidAssignmentID = errors.New("cannot create submission without an associated assignment")
)

// CreateSubmission creates a new submission record or updates the most
// recent submission, as defined by the provided submissionQuery.
// The submissionQuery must always specify the assignment, and may specify the ID of
// either an individual student or a group, but not both.
func (db *GormDB) CreateSubmission(submission *pb.Submission) error {
	if err := db.check(submission); err != nil {
		return err
	}

	// Make a new submission struct for the database query to check
	// whether a submission record for the given lab and user/group
	// already exists. We cannot reuse the incoming submission
	// because the query would attempt to match all the test result
	// fields as well.
	query := &pb.Submission{
		AssignmentID: submission.GetAssignmentID(),
		UserID:       submission.GetUserID(),
		GroupID:      submission.GetGroupID(),
	}

	return db.conn.Transaction(func(tx *gorm.DB) error {
		// We want the last record as there can be multiple submissions
		// for the same student/group and lab in the database.
		if err := tx.Last(query, query).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err // will rollback transaction
		}
		if submission.ID != 0 {
			if err := tx.First(&pb.Submission{}, &pb.Submission{ID: submission.ID}).Error; err != nil {
				return err // will rollback transaction
			}
			if err := tx.Where("submission_id = ?", submission.ID).Delete(&score.Score{}).Error; err != nil {
				return err // will rollback transaction
			}
			if err := tx.Where("submission_id = ?", submission.ID).Delete(&score.BuildInfo{}).Error; err != nil {
				return err // will rollback transaction
			}
			if submission.BuildInfo != nil {
				submission.BuildInfo.SubmissionID = submission.ID
			}
			for _, sc := range submission.Scores {
				sc.SubmissionID = submission.ID
			}
		}
		if err := tx.Save(submission).Error; err != nil {
			return err // will rollback transaction
		}
		return nil // will commit transaction
	})
}

// check returns an error if the submission query is invalid; otherwise nil is returned.
func (db *GormDB) check(submission *pb.Submission) error {
	// Foreign key must be greater than 0.
	if submission.AssignmentID < 1 {
		return ErrInvalidAssignmentID
	}

	// Either user or group id must be set, but not both.
	var m *gorm.DB
	switch {
	case submission.UserID > 0 && submission.GroupID > 0:
		return ErrInvalidSubmission
	case submission.UserID > 0:
		m = db.conn.First(&pb.User{ID: submission.UserID})
	case submission.GroupID > 0:
		m = db.conn.First(&pb.Group{ID: submission.GroupID})
	default:
		// neither UserID nor GroupID are not set
		return ErrInvalidSubmission
	}

	// Check that user/group with given ID exists.
	var idCount int64
	if err := m.Count(&idCount).Error; err != nil {
		if submission.UserID > 0 {
			return fmt.Errorf("user %d not found for submission: %+v: %w", submission.UserID, submission, err)
		} else {
			return fmt.Errorf("group %d not found for submission: %+v: %w", submission.GroupID, submission, err)
		}
	}

	// Checks that the assignment exists.
	var assignment int64
	if err := db.conn.Model(&pb.Assignment{}).Where(&pb.Assignment{
		ID: submission.AssignmentID,
	}).Count(&assignment).Error; err != nil {
		return fmt.Errorf("assignment %d not found: %w", submission.AssignmentID, err)
	}

	// Exactly one assignment and user/group must exist together.
	if assignment+idCount != 2 {
		return fmt.Errorf("inconsistent database state: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

// GetSubmission fetches a submission record.
func (db *GormDB) GetSubmission(query *pb.Submission) (*pb.Submission, error) {
	var submission pb.Submission
	if err := db.conn.Preload("Reviews").
		Preload("BuildInfo").
		Preload("Scores").
		Preload("Reviews.GradingBenchmarks").
		Preload("Reviews.GradingBenchmarks.Criteria").
		Where(query).Last(&submission).Error; err != nil {
		return nil, err
	}

	return &submission, nil
}

// GetLastSubmissions returns all submissions for the active assignment for the given course.
// The query may specify both UserID and GroupID to fetch both user and group submissions.
func (db *GormDB) GetLastSubmissions(courseID uint64, query *pb.Submission) ([]*pb.Submission, error) {
	var course pb.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}

	var latestSubs []*pb.Submission
	for _, a := range course.Assignments {
		query.AssignmentID = a.GetID()
		temp, err := db.GetSubmission(query)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return nil, err
		}
		latestSubs = append(latestSubs, temp)
	}
	return latestSubs, nil
}

// GetSubmissions returns all submissions matching the query.
func (db *GormDB) GetSubmissions(query *pb.Submission) ([]*pb.Submission, error) {
	var submissions []*pb.Submission
	if err := db.conn.Find(&submissions, &query).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}

// UpdateSubmission updates submission with the given approved status.
func (db *GormDB) UpdateSubmission(query *pb.Submission) error {
	return db.conn.Save(query).Error
}

// UpdateSubmissions approves and/or releases all submissions that have score
// equal or above the provided score for the given assignment ID
func (db *GormDB) UpdateSubmissions(courseID uint64, query *pb.Submission) error {
	return db.conn.
		Model(query).
		Where("assignment_id = ?", query.AssignmentID).
		Where("score >= ?", query.Score).
		Updates(&pb.Submission{
			Status:   query.Status,
			Released: query.Released,
		}).Error
}

// CreateReview creates a new submission review
func (db *GormDB) CreateReview(query *pb.Review) error {
	return db.conn.Create(query).Error
}

// UpdateReview updates feedback text, review and ready status
func (db *GormDB) UpdateReview(query *pb.Review) error {
	// By default, Gorm will not update zero value fields; such as the Ready bool field.
	// For additional context, see
	// https://github.com/quickfeed/quickfeed/issues/569#issuecomment-1013729572
	return db.conn.Model(&query).Select("*").Updates(&pb.Review{
		ID:           query.ID,
		SubmissionID: query.SubmissionID,
		Feedback:     query.Feedback,
		Ready:        query.Ready,
		Score:        query.Score,
		ReviewerID:   query.ReviewerID,
		Edited:       query.Edited,
	}).Error
}

// DeleteReview removes all reviews matching the query
func (db *GormDB) DeleteReview(query *pb.Review) error {
	return db.conn.Delete(&pb.Review{}, &query).Error
}
