package ci

import (
	"fmt"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// RecordResults for the course and assignment given by the run data structure.
// If the results argument is nil, then the submission is considered to be a manual review.
func (r *RunData) RecordResults(logger *zap.SugaredLogger, db database.Database, results *score.Results) (*qf.Submission, error) {
	defer func() {
		if m := recover(); m != nil {
			logger.Errorf("Recovered from panic: %v", m)
		}
	}()
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
	logger.Debugf("Recorded %s for %s with status %s and score %d", resType, r, newSubmission.GetStatuses(), newSubmission.GetScore())

	if !r.Rebuild {
		if err := r.updateSlipDays(logger, db, newSubmission); err != nil {
			return nil, fmt.Errorf("failed to update slip days for %s: %w", r, err)
		}
		logger.Debugf("Updated slip days for %s", r)
	}
	return newSubmission, nil
}

func (r *RunData) previousSubmission(db database.Database) (*qf.Submission, error) {
	submissionQuery := &qf.Submission{
		AssignmentID: r.Assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
	}
	return db.GetSubmission(submissionQuery)
}

func (r *RunData) newSubmission(previous *qf.Submission, results *score.Results) (string, *qf.Submission) {
	if results != nil {
		return "test execution", r.newTestRunSubmission(previous, results)
	}
	return "manual review", r.newManualReviewSubmission(previous)
}

func (r *RunData) newManualReviewSubmission(previous *qf.Submission) *qf.Submission {
	return &qf.Submission{
		ID:           previous.GetID(),
		AssignmentID: r.Assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
		CommitHash:   r.CommitID,
		Score:        previous.GetScore(),
		Grades:       previous.GetGrades(),
		BuildInfo: &score.BuildInfo{
			SubmissionDate: timestamppb.Now(),
			BuildDate:      timestamppb.Now(),
			BuildLog:       "",
			ExecTime:       1,
		},
	}
}

func (r *RunData) newTestRunSubmission(previous *qf.Submission, results *score.Results) *qf.Submission {
	if r.Rebuild && previous != nil && previous.GetBuildInfo() != nil {
		// Keep previous submission's delivery date if this is a rebuild.
		results.BuildInfo.SubmissionDate = previous.GetBuildInfo().GetSubmissionDate()
	}
	score := results.Sum()
	previous.SetGradesIfApproved(r.Assignment, score)
	return &qf.Submission{
		ID:           previous.GetID(),
		AssignmentID: r.Assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
		CommitHash:   r.CommitID,
		Score:        score,
		Grades:       previous.GetGrades(),
		BuildInfo:    results.GetBuildInfo(),
		Scores:       results.Scores,
	}
}

// slipDayUpdater is satisfied by both *qf.Group and *qf.Enrollment, letting the
// group and individual submission paths share the same slip-day update logic in RunData.updateSlipDays.
type slipDayUpdater interface {
	GetID() uint64
	GetUsedSlipDays() []*qf.UsedSlipDays
	UpdateSlipDays(assignment *qf.Assignment, submission *qf.Submission) error
}

func (r *RunData) updateSlipDays(logger *zap.SugaredLogger, db database.Database, submission *qf.Submission) (err error) {
	var holder slipDayUpdater
	if submission.GetGroupID() > 0 {
		if !r.Assignment.GetIsGroupLab() {
			// A group submission to a non-group lab should not update slip days.
			logger.Debugf("Skipping slip-day update: group %d pushed to non-group lab %d", submission.GetGroupID(), r.Assignment.GetID())
			return nil
		}
		holder, err = db.GetGroup(submission.GetGroupID())
		if err != nil {
			return fmt.Errorf("failed to get group %d: %w", submission.GetGroupID(), err)
		}
	} else {
		holder, err = db.GetEnrollmentByCourseAndUser(r.Assignment.GetCourseID(), submission.GetUserID())
		if err != nil {
			return fmt.Errorf("failed to get enrollment for user %d in course %d: %w", submission.GetUserID(), r.Assignment.GetCourseID(), err)
		}
	}
	if err := holder.UpdateSlipDays(r.Assignment, submission); err != nil {
		return fmt.Errorf("failed to update slip days for %s (id %d) in course %d: %w", r, holder.GetID(), r.Assignment.GetCourseID(), err)
	}
	if err := db.UpdateSlipDays(holder.GetUsedSlipDays()); err != nil {
		return fmt.Errorf("failed to update slip days for %s (id %d) in course %d: %w", r, holder.GetID(), r.Assignment.GetCourseID(), err)
	}
	return nil
}

// GetOwners returns the UserIDs of a user or group repository's owners.
// Returns an error if no owners could be found.
// This method should only be called for a user or group repository.
func (r *RunData) GetOwners(db database.Database) ([]uint64, error) {
	var owners []uint64
	if r.Repo.IsUserRepo() {
		owners = []uint64{r.Repo.GetUserID()}
	}
	if r.Repo.IsGroupRepo() {
		group, err := db.GetGroup(r.Repo.GetGroupID())
		if err == nil {
			owners = group.UserIDs()
		}
	}
	if len(owners) == 0 {
		return nil, fmt.Errorf("failed to get owners for %s", r)
	}
	return owners, nil
}
