package ci

import (
	"fmt"
	"time"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/stream"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RecordResults for the course and assignment given by the run data structure.
// If the results argument is nil, then the submission is considered to be a manual review.
func (r RunData) RecordResults(logger *zap.SugaredLogger, db database.Database, s *stream.StreamServices, results *score.Results) (*qf.Submission, error) {
	logger.Debugf("Fetching (if any) previous submission for %s", r)
	previous, err := r.previousSubmission(db)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get previous submission: %w", err)
	}
	if previous == nil {
		logger.Debugf("Recording new submission for %s", r)
	} else {
		logger.Debugf("Updating submission %d for %s", previous.GetID(), r)
	}

	resType, newSubmission := r.newSubmission(previous, results)
	if err = db.CreateSubmission(newSubmission); err != nil {
		return nil, fmt.Errorf("failed to record submission %d for %s: %w", previous.GetID(), r, err)
	}
	logger.Debugf("Recorded %s for %s with status %s and score %d", resType, r, newSubmission.GetStatus(), newSubmission.GetScore())

	if !r.Rebuild {
		if err := r.updateSlipDays(db, newSubmission); err != nil {
			return nil, fmt.Errorf("failed to update slip days for %s: %w", r, err)
		}
		logger.Debugf("Updated slip days for %s", r)
	}

	switch {
	case r.Repo.IsUserRepo():
		s.Submission.SendTo(newSubmission, r.Repo.GetUserID())
	case r.Repo.IsGroupRepo():
		if group, err := db.GetGroup(r.Repo.GetGroupID()); err != nil {
			logger.Debugf("Failed stream submission to group %d: %v", r.Repo.GetGroupID(), err)
		} else {
			s.Submission.SendTo(newSubmission, group.UserIDs()...)
		}
	}
	return newSubmission, nil
}

func (r RunData) previousSubmission(db database.Database) (*qf.Submission, error) {
	submissionQuery := &qf.Submission{
		AssignmentID: r.Assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
	}
	return db.GetSubmission(submissionQuery)
}

func (r RunData) newSubmission(previous *qf.Submission, results *score.Results) (string, *qf.Submission) {
	if results != nil {
		return "test execution", r.newTestRunSubmission(previous, results)
	}
	return "manual review", r.newManualReviewSubmission(previous)
}

func (r RunData) newManualReviewSubmission(previous *qf.Submission) *qf.Submission {
	return &qf.Submission{
		ID:           previous.GetID(),
		AssignmentID: r.Assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
		CommitHash:   r.CommitID,
		Score:        previous.GetScore(),
		Status:       previous.GetStatus(),
		Released:     previous.GetReleased(),
		BuildInfo: &score.BuildInfo{
			BuildDate: time.Now().Format(qf.TimeLayout),
			BuildLog:  "No automated tests for this assignment",
			ExecTime:  1,
		},
	}
}

func (r RunData) newTestRunSubmission(previous *qf.Submission, results *score.Results) *qf.Submission {
	if r.Rebuild && previous != nil && previous.BuildInfo != nil {
		// Keep previous submission's delivery date if this is a rebuild.
		results.BuildInfo.BuildDate = previous.BuildInfo.BuildDate
	}
	score := results.Sum()
	return &qf.Submission{
		ID:           previous.GetID(),
		AssignmentID: r.Assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
		CommitHash:   r.CommitID,
		Score:        score,
		Status:       r.Assignment.IsApproved(previous, score),
		BuildInfo:    results.BuildInfo,
		Scores:       results.Scores,
	}
}

func (r RunData) updateSlipDays(db database.Database, submission *qf.Submission) error {
	buildDate := submission.GetBuildInfo().GetBuildDate()
	buildTime, err := time.Parse(qf.TimeLayout, buildDate)
	if err != nil {
		return fmt.Errorf("failed to parse time from build date (%s): %w", buildDate, err)
	}

	enrollments := make([]*qf.Enrollment, 0)
	if submission.GroupID > 0 {
		group, err := db.GetGroup(submission.GroupID)
		if err != nil {
			return fmt.Errorf("failed to get group %d: %w", submission.GroupID, err)
		}
		enrollments = append(enrollments, group.Enrollments...)
	} else {
		enrol, err := db.GetEnrollmentByCourseAndUser(r.Assignment.CourseID, submission.UserID)
		if err != nil {
			return fmt.Errorf("failed to get enrollment for user %d in course %d: %w", submission.UserID, r.Assignment.CourseID, err)
		}
		enrollments = append(enrollments, enrol)
	}

	for _, enrol := range enrollments {
		if err := enrol.UpdateSlipDays(buildTime, r.Assignment, submission); err != nil {
			return fmt.Errorf("failed to update slip days for user %d in course %d: %w", enrol.UserID, r.Assignment.CourseID, err)
		}
		if err := db.UpdateSlipDays(enrol.UsedSlipDays); err != nil {
			return fmt.Errorf("failed to update slip days for enrollment %d (user %d) (course %d): %w", enrol.ID, enrol.UserID, enrol.CourseID, err)
		}
	}
	return nil
}
