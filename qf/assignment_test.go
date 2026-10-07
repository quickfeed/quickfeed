package qf

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/score"
)

type timeoutContextKey struct{}

func TestAssignmentWithTimeoutPreservesParent(t *testing.T) {
	parent := context.WithValue(context.Background(), timeoutContextKey{}, "value")
	ctx, cancel := (&Assignment{}).WithTimeout(parent, time.Minute)
	defer cancel()
	if got := ctx.Value(timeoutContextKey{}); got != "value" {
		t.Errorf("WithTimeout() context value = %v, want value", got)
	}
	if _, ok := ctx.Deadline(); !ok {
		t.Error("WithTimeout() context has no deadline")
	}
}

func TestAssignmentZeroScoreTests(t *testing.T) {
	tests := []struct {
		name       string
		assignment *Assignment
		wantScores []*score.Score
	}{
		{
			name:       "NoExpectedTests",
			assignment: &Assignment{},
			wantScores: nil,
		},
		{
			name:       "SingleExpectedTest",
			assignment: &Assignment{ExpectedTests: []*TestInfo{{TestName: "TestA", MaxScore: 10, Weight: 5}}},
			wantScores: []*score.Score{{TestName: "TestA", MaxScore: 10, Weight: 5}},
		},
		{
			name:       "MultipleExpectedTests",
			assignment: &Assignment{ExpectedTests: []*TestInfo{{TestName: "TestA", MaxScore: 10, Weight: 5}, {TestName: "TestB", MaxScore: 20, Weight: 10}}},
			wantScores: []*score.Score{{TestName: "TestA", MaxScore: 10, Weight: 5}, {TestName: "TestB", MaxScore: 20, Weight: 10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.assignment.ZeroScoreTests()
			if len(got) != len(tt.assignment.ExpectedTests) {
				t.Errorf("ZeroScoreTests() returned %d tests, expected %d", len(got), len(tt.assignment.ExpectedTests))
			}
			if diff := cmp.Diff(tt.wantScores, got); diff != "" {
				t.Errorf("ZeroScoreTests() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAssignmentDefaultReviewWeight(t *testing.T) {
	tests := []*TestInfo{{TestName: "TestA", MaxScore: 1, Weight: 1}}
	cases := []struct {
		name       string
		assignment *Assignment
		want       uint32
	}{
		{name: "TestsOnly", assignment: &Assignment{ExpectedTests: tests}, want: 0},
		{name: "ReviewsOnly", assignment: &Assignment{Reviewers: 1}, want: 100},
		{name: "TestsAndReviews", assignment: &Assignment{Reviewers: 2, ExpectedTests: tests}, want: 50},
		{name: "Neither", assignment: &Assignment{}, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.assignment.DefaultReviewWeight(); got != tc.want {
				t.Errorf("DefaultReviewWeight() = %d, want %d", got, tc.want)
			}
		})
	}
}
