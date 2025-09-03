package qf

import (
	context "context"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
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

// ZeroScoreTests returns a slice of score.Score objects with zero scores
// for all expected tests in this assignment.
func (a *Assignment) ZeroScoreTests() []*score.Score {
	expectedTests := a.GetExpectedTests()
	if len(expectedTests) == 0 {
		return nil
	}

	scores := make([]*score.Score, len(expectedTests))
	for i, testInfo := range expectedTests {
		scores[i] = &score.Score{
			TestName: testInfo.GetTestName(),
			MaxScore: testInfo.GetMaxScore(),
			Weight:   testInfo.GetWeight(),
		}
	}
	return scores
}
