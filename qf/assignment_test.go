package qf

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/score"
)

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
