package ci

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// RecordResults for the course and assignment given by the run data structure.
// If the results argument is nil, then the submission is considered to be a manual review.
func (r *RunData) RecordResults(ctx context.Context, db database.Database, results *score.Results) (*qf.Submission, error) {
	// The run identifies the course, assignment, and job owner for every record below.
	logger := qlog.FromContext(ctx).With(runLabel, r.String())
	defer func() {
		if m := recover(); m != nil {
			// A panic here is rare and leaves the submission unrecorded, so keep
			// the stack: the recover swallows it, and the run attributes alone are
			// not enough to locate the cause.
			logger.Error("recovered while recording results", "panic", m, "stack", string(debug.Stack()))
		}
	}()
	logger.Debug("fetching previous submission")
	previous, err := r.previousSubmission(db)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("getting previous submission: %w", err)
	}
	if previous == nil {
		logger.Debug("recording new submission")
	} else {
		logger.Debug("updating submission", label.SubmissionID, previous.GetID())
	}

	resType, newSubmission := r.newSubmission(previous, results)
	if err = db.CreateSubmission(newSubmission); err != nil {
		return nil, fmt.Errorf("recording submission %d for %s: %w", previous.GetID(), r, err)
	}
	logger.Debug("recorded submission", "result_type", resType, "status", newSubmission.GetStatuses(), "score", newSubmission.GetScore())

	// Only a student push updates slip-day accounting, independent of whether
	// the test run failed; a failed run still delivers the push, and its date
	// is the date the student submitted. A rebuild is triggered from the
	// frontend, keeps the original submission date, and must never account for
	// the same push a second time.
	if !r.Rebuild {
		if err := r.updateSlipDays(logger, db, newSubmission); err != nil {
			return nil, fmt.Errorf("updating slip days for %s: %w", r, err)
		}
		logger.Debug("updated slip days")
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
	if !results.ScoresValid() {
		return r.newFailedRunSubmission(previous, results)
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

// newFailedRunSubmission records a failed run without overwriting the previous
// submission's score, grades, and scores. The build info carries the failed
// attempt's status, log, and submission date, while the commit hash names the
// failed commit so a rebuild retries it.
//
// The failed attempt's submission date is the date the student pushed, and is
// therefore the date that RecordResults accounts for slip days; a rebuild of
// the failed commit keeps the original date, set by newTestRunSubmission.
func (r *RunData) newFailedRunSubmission(previous *qf.Submission, results *score.Results) *qf.Submission {
	return &qf.Submission{
		ID:           previous.GetID(),
		AssignmentID: r.Assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
		CommitHash:   r.CommitID,
		Score:        previous.GetScore(),
		Grades:       previous.GetGrades(),
		BuildInfo:    results.GetBuildInfo(),
		Scores:       previous.GetScores(),
	}
}

// slipDayUpdater is satisfied by both *qf.Group and *qf.Enrollment, letting the
// group and individual submission paths share the same slip-day update logic in RunData.updateSlipDays.
type slipDayUpdater interface {
	GetID() uint64
	GetUsedSlipDays() []*qf.UsedSlipDays
	UpdateSlipDays(assignment *qf.Assignment, submission *qf.Submission) error
}

func (r *RunData) updateSlipDays(logger *slog.Logger, db database.Database, submission *qf.Submission) (err error) {
	var holder slipDayUpdater
	if submission.GetGroupID() > 0 {
		if !r.Assignment.GetIsGroupLab() {
			// A group submission to a non-group lab should not update slip days.
			logger.Debug("skipping slip-day update for group submission to individual assignment", label.GroupID, submission.GetGroupID(), label.AssignmentID, r.Assignment.GetID())
			return nil
		}
		holder, err = db.GetGroup(submission.GetGroupID())
		if err != nil {
			return fmt.Errorf("getting group %d: %w", submission.GetGroupID(), err)
		}
	} else {
		holder, err = db.GetEnrollmentByCourseAndUser(r.Assignment.GetCourseID(), submission.GetUserID())
		if err != nil {
			return fmt.Errorf("getting enrollment for user %d in course %d: %w", submission.GetUserID(), r.Assignment.GetCourseID(), err)
		}
	}
	if err := holder.UpdateSlipDays(r.Assignment, submission); err != nil {
		return fmt.Errorf("updating slip days for %s (id %d) in course %d: %w", r, holder.GetID(), r.Assignment.GetCourseID(), err)
	}
	if err := db.UpdateSlipDays(holder.GetUsedSlipDays()); err != nil {
		return fmt.Errorf("updating slip days for %s (id %d) in course %d: %w", r, holder.GetID(), r.Assignment.GetCourseID(), err)
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
		return nil, fmt.Errorf("no owners for %s", r)
	}
	return owners, nil
}
