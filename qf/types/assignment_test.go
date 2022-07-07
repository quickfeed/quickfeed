package types_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf/types"
)

func TestIsApproved(t *testing.T) {
	a := &types.Assignment{}
	b := &types.Assignment{
		ScoreLimit: 80,
	}
	c := &types.Assignment{
		AutoApprove: true,
		ScoreLimit:  80,
	}
	isApprovedTests := []struct {
		name       string
		assignment *types.Assignment
		submission *types.Submission
		score      uint32
		expected   types.Submission_Status
	}{
		{
			name:       "Assignment:ScoreLimit=0:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: a,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 50},
			score:      55,
			expected:   types.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: b,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 50},
			score:      55,
			expected:   types.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:80",
			assignment: b,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 50},
			score:      80,
			expected:   types.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=80,NewScore:75",
			assignment: b,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 80},
			score:      75,
			expected:   types.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=80,NewScore:85",
			assignment: b,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 80},
			score:      85,
			expected:   types.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=REJECTED:OldScore=50,NewScore:80",
			assignment: b,
			submission: &types.Submission{Status: types.Submission_REJECTED, Score: 50},
			score:      80,
			expected:   types.Submission_REJECTED,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=REVISION:OldScore=50,NewScore:80",
			assignment: b,
			submission: &types.Submission{Status: types.Submission_REVISION, Score: 50},
			score:      80,
			expected:   types.Submission_REVISION,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:80",
			assignment: b,
			submission: &types.Submission{Status: types.Submission_APPROVED, Score: 50},
			score:      80,
			expected:   types.Submission_APPROVED,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: c,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 50},
			score:      55,
			expected:   types.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:79",
			assignment: c,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 50},
			score:      79,
			expected:   types.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:80",
			assignment: c,
			submission: &types.Submission{Status: types.Submission_NONE, Score: 50},
			score:      80,
			expected:   types.Submission_APPROVED,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:0",
			assignment: c,
			submission: &types.Submission{Status: types.Submission_APPROVED, Score: 50},
			score:      0,
			expected:   types.Submission_APPROVED,
		},
	}

	for _, test := range isApprovedTests {
		t.Run(test.name, func(t *testing.T) {
			got := test.assignment.IsApproved(test.submission, test.score)
			if got != test.expected {
				t.Errorf("IsApproved(%v, %v, %d) = %v, expected %v", test.assignment, test.submission, test.score, got, test.expected)
			}
		})
	}
}
