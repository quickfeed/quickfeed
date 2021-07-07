package ag

import (
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

// IsApproved returns true if this assignment is already approved for the
// latest submission, or if the score of the latest submission is sufficient
// to autoapprove the assignment.
func (a *Assignment) IsApproved(latest *Submission, score uint32) bool {
	// keep approved status if already approved
	approved := latest.GetStatus() == Submission_APPROVED
	if a.GetAutoApprove() && score >= a.GetScoreLimit() {
		approved = true
	}
	return approved
}

// IsManual returns true if the assignment is manually graded.
func (a *Assignment) IsManual() bool {
	return len(a.GetGradingBenchmarks()) > 0
}

// HasTests returns true if the assignment has tests to be executed by QuickFeed's CI.
func (a *Assignment) HasTests() bool {
	return !a.IsManual()
}

// CloneWithoutSubmissions returns a deep copy of the given assignment
// without submissions
func (a *Assignment) CloneWithoutSubmissions() *Assignment {
	return &Assignment{
		ID:                a.ID,
		CourseID:          a.CourseID,
		Name:              a.Name,
		ScriptFile:        a.ScriptFile,
		Deadline:          a.Deadline,
		AutoApprove:       a.AutoApprove,
		Order:             a.Order,
		IsGroupLab:        a.IsGroupLab,
		ScoreLimit:        a.ScoreLimit,
		Reviewers:         a.Reviewers,
		SkipTests:         a.SkipTests,
		GradingBenchmarks: a.GradingBenchmarks,
	}
}
