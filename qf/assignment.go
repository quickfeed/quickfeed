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
func (a *Assignment) WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	t := a.GetContainerTimeout()
	if t > 0 {
		timeout = time.Duration(t) * time.Minute
	}
	return context.WithTimeout(parent, timeout)
}

// CloneWithoutSubmissions returns a deep copy of the assignment without submissions.
func (a *Assignment) CloneWithoutSubmissions() *Assignment {
	clone := proto.CloneOf(a)
	clone.Submissions = nil
	return clone
}

// GradedManually returns true if the assignment will be graded manually.
func (a *Assignment) GradedManually() bool {
	return a.GetReviewers() > 0
}

// HasTests returns true if submissions to the assignment are tested.
func (a *Assignment) HasTests() bool {
	return len(a.GetExpectedTests()) > 0
}

// RunsTests returns true if submissions to the assignment are tested: those
// to an assignment without reviewers are, even if it has no expected tests.
func (a *Assignment) RunsTests() bool {
	return !a.GradedManually() || a.HasTests()
}

// DefaultReviewWeight returns the review weight for an assignment that does not
// specify one: 0 without reviewers, 100 without tests, and 50 with both.
func (a *Assignment) DefaultReviewWeight() uint32 {
	switch {
	case !a.GradedManually():
		return 0
	case !a.HasTests():
		return 100
	default:
		return 50
	}
}

// WeightedScore returns the submission score for the given test and review
// scores, weighted by the assignment's review weight. Reviews count only if
// the assignment has reviewers.
func (a *Assignment) WeightedScore(testScore, reviewScore uint32) uint32 {
	var weight uint64
	if a.GradedManually() {
		weight = uint64(min(a.GetReviewWeight(), 100))
	}
	// Adding 50 rounds the weighted score to the nearest integer.
	return uint32((uint64(testScore)*(100-weight) + uint64(reviewScore)*weight + 50) / 100)
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
