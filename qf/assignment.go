package qf

import (
	context "context"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	days = time.Duration(24 * time.Hour)
)

// SinceDeadline returns the duration since the deadline.
// A positive duration means the deadline has passed, whereas
// a negative duration means the deadline has not yet passed.
func (a *Assignment) SinceDeadline(now time.Time) time.Duration {
	return now.Sub(a.GetDeadline().AsTime())
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

// SubmissionStatus returns the existing grade submission status, or an approved submission status
// if the score of the latest submission is sufficient to autoapprove the assignment.
func (a *Assignment) SubmissionStatus(latest *Submission, score uint32, rebuild bool) []*Grade {
	if a.GetAutoApprove() && latest.GetScore() >= score && rebuild {
		// if assignment is set to autoapprove, and the latest submission has a score
		// equal to or higher than the new score, and the submission is being rebuilt,
		// then set the submission to revision and set a comment to indicate that the
		// submission has been rebuilt resulting in a lower score.
		latest.SetGradeAll(Submission_REVISION)
		latest.SetCommentAll("As a result of a rebuild, the score has been lowered.")
	} else if a.GetAutoApprove() && score >= a.GetScoreLimit() {
		// if assignment is set to autoapprove, and the latest submission has a score
		// equal to or higher than the score limit, then set the submission to approved.
		latest.SetGradeAll(Submission_APPROVED)
	}
	// keep existing status if already approved/revision/rejected
	return latest.GetGrades()
}

// CloneWithoutSubmissions returns a deep copy of the assignment without submissions.
func (a *Assignment) CloneWithoutSubmissions() *Assignment {
	clone := proto.Clone(a).(*Assignment)
	clone.Submissions = nil
	return clone
}

// GradedManually returns true if the assignment will be graded manually.
func (a *Assignment) GradedManually() bool {
	return a.GetReviewers() > 0
}
