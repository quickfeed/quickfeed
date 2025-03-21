package database

import (
	"errors"
	"fmt"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
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
func (db *GormDB) CreateSubmission(submission *qf.Submission) error {
	if err := db.check(submission); err != nil {
		return err
	}

	// Make a new submission struct for the database query to check
	// whether a submission record for the given lab and user/group
	// already exists. We cannot reuse the incoming submission
	// because the query would attempt to match all the test result
	// fields as well.
	query := &qf.Submission{
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
		if submission.GetID() != 0 {
			if err := tx.First(&qf.Submission{}, &qf.Submission{ID: submission.GetID()}).Error; err != nil {
				return err // will rollback transaction
			}
			if err := tx.Where("submission_id = ?", submission.GetID()).Delete(&score.Score{}).Error; err != nil {
				return err // will rollback transaction
			}
			if err := tx.Where("submission_id = ?", submission.GetID()).Delete(&score.BuildInfo{}).Error; err != nil {
				return err // will rollback transaction
			}
			if submission.GetBuildInfo() != nil {
				submission.BuildInfo.SubmissionID = submission.GetID()
			}
			for _, sc := range submission.GetScores() {
				sc.SubmissionID = submission.GetID()
			}
		} else {
			// Initialize grades for the new submission
			if err := setGrades(tx, submission); err != nil {
				return err // will rollback transaction
			}
		}
		// Full save associations is required to save any nested grades
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(submission).Error; err != nil {
			return err // will rollback transaction
		}
		return nil // will commit transaction
	})
}

// setGrades adds grades for any user or group related to the submission
// which are then saved to the database upon creation of the submission.
func setGrades(tx *gorm.DB, s *qf.Submission) error {
	if s.GetUserID() > 0 {
		// Add a grade for the user if the submission is not a group submission.
		// Create a new grade for the user.
		s.Grades = []*qf.Grade{{
			UserID: s.GetUserID(),
			Status: s.GetStatusByUser(s.GetUserID()),
		}}
	}
	if s.GetGroupID() > 0 {
		// If the submission is for a group, create a new grade for each user in the group.
		// Get all the user IDs in the group
		var userIDs []uint64
		tx.Model(&qf.Enrollment{}).Where("group_id = ?", s.GetGroupID()).Pluck("user_id", &userIDs)

		s.Grades = make([]*qf.Grade, len(userIDs))
		for idx, id := range userIDs {
			// Create a grade for each user in the group
			s.Grades[idx] = &qf.Grade{
				UserID: id,
			}
		}
	}
	// Find the assignment associated with the submission
	var assignment qf.Assignment
	if err := tx.First(&assignment, s.GetAssignmentID()).Error; err != nil {
		return err
	}
	// Set the submission status based on the assignment's auto-approve settings
	s.Grades = assignment.SubmissionStatus(s, s.GetScore())
	return nil
}

// check returns an error if the submission query is invalid; otherwise nil is returned.
func (db *GormDB) check(submission *qf.Submission) error {
	// Foreign key must be greater than 0.
	if submission.GetAssignmentID() < 1 {
		return ErrInvalidAssignmentID
	}

	// Either user or group id must be set, but not both.
	var m *gorm.DB
	switch {
	case submission.GetUserID() > 0 && submission.GetGroupID() > 0:
		return ErrInvalidSubmission
	case submission.GetUserID() > 0:
		m = db.conn.First(&qf.User{ID: submission.GetUserID()})
	case submission.GetGroupID() > 0:
		m = db.conn.First(&qf.Group{ID: submission.GetGroupID()})
	default:
		// neither UserID nor GroupID are not set
		return ErrInvalidSubmission
	}

	// Check that user/group with given ID exists.
	var idCount int64
	if err := m.Count(&idCount).Error; err != nil {
		if submission.GetUserID() > 0 {
			return fmt.Errorf("user %d not found for submission: %+v: %w", submission.GetUserID(), submission, err)
		} else {
			return fmt.Errorf("group %d not found for submission: %+v: %w", submission.GetGroupID(), submission, err)
		}
	}

	// Checks that the assignment exists.
	var assignment int64
	if err := db.conn.Model(&qf.Assignment{}).Where(&qf.Assignment{
		ID: submission.GetAssignmentID(),
	}).Count(&assignment).Error; err != nil {
		return fmt.Errorf("assignment %d not found: %w", submission.GetAssignmentID(), err)
	}

	// Exactly one assignment and user/group must exist together.
	if assignment+idCount != 2 {
		return fmt.Errorf("inconsistent database state: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

// GetSubmission fetches a submission record.
func (db *GormDB) GetSubmission(query *qf.Submission) (*qf.Submission, error) {
	var submission qf.Submission
	if err := db.conn.Preload("Reviews").
		Preload("BuildInfo").
		Preload("Scores").
		Preload("Grades").
		Preload("Reviews.GradingBenchmarks").
		Preload("Reviews.GradingBenchmarks.Criteria").
		Where(query).Last(&submission).Error; err != nil {
		return nil, err
	}

	return &submission, nil
}

// GetLastSubmission returns the last submission for the given submission query and course ID.
// If no assignment matches the found submission's assignment ID and provided course ID, an error is returned.
func (db *GormDB) GetLastSubmission(courseID uint64, query *qf.Submission) (*qf.Submission, error) {
	submission, err := db.GetSubmission(query)
	if err != nil {
		return nil, err
	}
	var assignment qf.Assignment
	if err := db.conn.Model(&qf.Assignment{}).Where(
		&qf.Assignment{ID: submission.GetAssignmentID(), CourseID: courseID},
	).First(&assignment).Error; err != nil {
		return nil, err
	}
	return submission, nil
}

// GetLastSubmissions returns all submissions for the active assignment for the given course.
// The query may specify both UserID and GroupID to fetch both user and group submissions.
func (db *GormDB) GetLastSubmissions(courseID uint64, query *qf.Submission) ([]*qf.Submission, error) {
	var course qf.Course
	if err := db.conn.Preload("Assignments").First(&course, courseID).Error; err != nil {
		return nil, err
	}

	var latestSubs []*qf.Submission
	for _, a := range course.GetAssignments() {
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
func (db *GormDB) GetSubmissions(query *qf.Submission) ([]*qf.Submission, error) {
	var submissions []*qf.Submission
	if err := db.conn.Preload("Grades").Find(&submissions, &query).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}

// UpdateSubmission updates submission with the given approved status.
func (db *GormDB) UpdateSubmission(query *qf.Submission) error {
	// full save associations is required to update the nested grades
	return db.conn.Session(&gorm.Session{FullSaveAssociations: true}).Updates(query).Error
}

// UpdateSubmissions approves and/or releases all submissions that have score
// equal or above the provided score for the given assignment ID
func (db *GormDB) UpdateSubmissions(query *qf.Submission, approve bool) error {
	return db.conn.Transaction(func(tx *gorm.DB) error {
		var submissionIDs []*uint64
		if err := tx.Model(&qf.Submission{}).
			Where("assignment_id = ? AND score >= ?", query.GetAssignmentID(), query.GetScore()).
			Pluck("id", &submissionIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(query).
			// Update the released status of all submissions that have score equal or above the provided score
			Where("id IN (?)", submissionIDs).
			Updates(&qf.Submission{
				Released: query.GetReleased(),
			}).Error; err != nil {
			return err
		}

		status := qf.Submission_APPROVED
		if !approve {
			status = qf.Submission_REJECTED
		}

		// Approve all Grades for the submissions
		err := tx.Model(&qf.Grade{}).
			Where("submission_id IN (?)", submissionIDs).
			Updates(&qf.Grade{
				Status: status,
			}).Error
		if err != nil {
			return err
		}

		return nil
	})
}

// GetReview fetches a review
func (db *GormDB) GetReview(query *qf.Review) (*qf.Review, error) {
	var review qf.Review
	if err := db.conn.Where(query).
		Preload("GradingBenchmarks", "review_id = (?)", query.GetID()).
		Preload("GradingBenchmarks.Criteria").
		First(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// CreateReview creates a new submission review.
func (db *GormDB) CreateReview(query *qf.Review) error {
	return db.conn.Create(query).Error
}

// UpdateReview updates a review.
func (db *GormDB) UpdateReview(query *qf.Review) error {
	// By default, Gorm will not update zero value fields; such as the Ready bool field.
	// Therefore we use Select before the Updates call. For additional context, see
	// https://github.com/quickfeed/quickfeed/issues/569#issuecomment-1013729572
	return db.conn.Model(&query).Select("*").Updates(&qf.Review{
		ID:           query.GetID(),
		SubmissionID: query.GetSubmissionID(),
		Feedback:     query.GetFeedback(),
		Ready:        query.GetReady(),
		Score:        query.GetScore(),
		ReviewerID:   query.GetReviewerID(),
		Edited:       query.GetEdited(),
	}).Error
}

// DeleteReview removes all reviews matching the query.
func (db *GormDB) DeleteReview(query *qf.Review) error {
	return db.conn.Delete(&qf.Review{}, &query).Error
}
