package ag

import (
	context "context"
	"time"
)

const (
	days = time.Duration(24 * time.Hour)
	zero = time.Duration(0)
)

// SinceDeadline returns the duration since the deadline.
// A positive duration means the deadline has passed, whereas
// a negative duration means the deadline has not yet passed.
func (a *Assignment) SinceDeadline(now time.Time) (time.Duration, error) {
	deadline, err := time.ParseInLocation(TimeLayout, a.GetDeadline(), now.Location())
	if err != nil {
		// this should not happen if deadlines are parsed and recorded correctly
		return zero, err
	}
	return now.Sub(deadline), nil
}

// WithTimeout returns a context with an execution timeout set to the assignment's specified
// container timeout. If the assignment has no container timeout, the provided timeout value
// is used instead.
func (a *Assignment) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	t := a.GetContainerTimeout()
	if t > 0 {
		timeout = time.Duration(t) * time.Minute
	}
	return context.WithTimeout(context.Background(), timeout)
}

// IsApproved returns an approved submission status if this assignment is already approved
// for the latest submission, or if the score of the latest submission is sufficient
// to autoapprove the assignment.
func (a *Assignment) IsApproved(latest *Submission, score uint32) Submission_Status {
	if a.GetAutoApprove() && score >= a.GetScoreLimit() {
		return Submission_APPROVED
	}
	// keep existing status if already approved/revision/rejected
	return latest.GetStatus()
}

// CloneWithoutSubmissions returns a deep copy of the given assignment
// without submissions.
func (a *Assignment) CloneWithoutSubmissions() *Assignment {
	return &Assignment{
		ID:                a.ID,
		CourseID:          a.CourseID,
		Name:              a.Name,
		Deadline:          a.Deadline,
		AutoApprove:       a.AutoApprove,
		Order:             a.Order,
		IsGroupLab:        a.IsGroupLab,
		ScoreLimit:        a.ScoreLimit,
		Reviewers:         a.Reviewers,
		GradingBenchmarks: a.GradingBenchmarks,
	}
}

// GradedManually returns true if the assignment will be graded manually.
func (a *Assignment) GradedManually() bool {
	return a.GetReviewers() > 0
}
