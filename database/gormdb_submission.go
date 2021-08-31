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
	var group int64
	if err := m.Count(&group).Error; err != nil {
		return fmt.Errorf("submission not found: %+v: %w", submission, err)
	}

	// Checks that the assignment exists.
	var assignment int64
	if err := db.conn.Model(&pb.Assignment{}).Where(&pb.Assignment{
		ID: submission.AssignmentID,
	}).Count(&assignment).Error; err != nil {
		return fmt.Errorf("assignment %d not found: %w", submission.AssignmentID, err)
	}

	if assignment+group != 2 {
		// Exactly one assignment and user/group must exist together.
		return fmt.Errorf("inconsistent database state: %w", gorm.ErrRecordNotFound)
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

	// We want the last record as there can be multiple submissions
	// for the same student/group and lab in the database.
	if err := db.conn.Last(query, query).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// TODO(meling) Rewrite this into a transaction.

	if submission.ID != 0 {
		// A submission will have a build info without any ID, using save will create a duplicate build info
		// in case the submission already exists in the database. We have to update build info explicitly.
		// There should be a less hacky way to do it.
		if submission.BuildInfo != nil {
			var buildInfo score.BuildInfo
			// TODO(meling); this query should probably do `err != nil && err != gorm.ErrRecordNotFound`
			// But add tests to ensure that this path is taken.
			if err := db.conn.Where("submission_id = ?", submission.ID).Last(&buildInfo).Error; err != nil {
				return fmt.Errorf("could not find old build info for submission %d: %w", submission.ID, err)
			}
			submission.BuildInfo.ID = buildInfo.ID
			if err := db.conn.Save(submission.BuildInfo).Error; err != nil {
				return fmt.Errorf("failed to save build info %+v: %w", submission.BuildInfo, err)
			}
		}
		for _, newScore := range submission.Scores {
			var oldScore score.Score
			query := &score.Score{
				SubmissionID: submission.ID,
				TestName:     newScore.TestName,
				MaxScore:     newScore.MaxScore,
				Weight:       newScore.Weight,
			}
			// TODO(meling); this query should probably do `err != nil && err != gorm.ErrRecordNotFound`
			// But add tests to ensure that this path is taken.
			if err := db.conn.Where(query).Last(&oldScore).Error; err != nil {
				return fmt.Errorf("could not find old score for %+v: %w", query, err)
			}
			newScore.ID = oldScore.ID
			if err := db.conn.Save(newScore).Error; err != nil {
				return fmt.Errorf("failed to save score %+v: %w", newScore, err)
			}
		}
	}
	err := db.conn.Save(submission).Error
	if err != nil {
		return fmt.Errorf("failed to save submission %+v: %w", submission, err)
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
	return db.conn.Model(&pb.Review{ID: query.ID}).Updates(&pb.Review{
		Feedback:   query.Feedback,
		Ready:      query.Ready,
		Score:      query.Score,
		ReviewerID: query.ReviewerID,
		Edited:     query.Edited,
	}).Error
}

// DeleteReview removes all reviews matching the query
func (db *GormDB) DeleteReview(query *pb.Review) error {
	return db.conn.Delete(&pb.Review{}, &query).Error
}
