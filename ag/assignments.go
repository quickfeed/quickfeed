package ag

import "time"

const (
	layout = "2006-01-02T15:04:05"
)

// DurationUntilDeadline returns the duration since the deadline.
func (m Assignment) DurationUntilDeadline(now time.Time) time.Duration {
	deadline, err := time.Parse(layout, m.GetDeadline())
	if err != nil {
	}
	return now.Sub(deadline)
}

// IsApproved returns true if this assignment is already approved for the
// latest submission, or if the score of the latest submission is sufficient
// to autoapprove the assignment.
func (m Assignment) IsApproved(latest *Submission, score uint32) bool {
	// keep approved status if already approved
	approved := latest.GetApproved()
	if m.GetAutoApprove() && score >= m.GetScoreLimit() {
		approved = true
	}
	return approved
}
