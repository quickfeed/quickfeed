package qf_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

func TestIsApproved(t *testing.T) {
	a := &qf.Assignment{}
	b := &qf.Assignment{
		ScoreLimit: 80,
	}
	c := &qf.Assignment{
		AutoApprove: true,
		ScoreLimit:  80,
	}
	d := &qf.Assignment{
		IsGroupLab:  false, // making it explicit that it isn't a group lab
		AutoApprove: true,
		ScoreLimit:  90,
	}
	e := &qf.Assignment{
		IsGroupLab:  true, // making it explicit that it is a group lab
		AutoApprove: true,
		ScoreLimit:  90,
	}
	isApprovedTests := []struct {
		name       string
		assignment *qf.Assignment
		submission *qf.Submission
		score      uint32
		expected   qf.Submission_Status
	}{
		{
			name:       "Assignment:ScoreLimit=0:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: a,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50},
			score:      55,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: b,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50},
			score:      55,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50},
			score:      80,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=80,NewScore:75",
			assignment: b,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 80},
			score:      75,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=80,NewScore:85",
			assignment: b,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 80},
			score:      85,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=REJECTED:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Status: qf.Submission_REJECTED, Score: 50},
			score:      80,
			expected:   qf.Submission_REJECTED,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=REVISION:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Status: qf.Submission_REVISION, Score: 50},
			score:      80,
			expected:   qf.Submission_REVISION,
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Status: qf.Submission_APPROVED, Score: 50},
			score:      80,
			expected:   qf.Submission_APPROVED,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: c,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50},
			score:      55,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:79",
			assignment: c,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50},
			score:      79,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:80",
			assignment: c,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50},
			score:      80,
			expected:   qf.Submission_APPROVED,
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:0",
			assignment: c,
			submission: &qf.Submission{Status: qf.Submission_APPROVED, Score: 50},
			score:      0,
			expected:   qf.Submission_APPROVED,
		},
		{
			name:       "Assignment:ScoreLimit=90:AutoApprove,Submission:GroupId=5:Status=NONE:OldScore=0,NewScore:50",
			assignment: d,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50, GroupID: 5},
			score:      50,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=90:AutoApprove,Submission:GroupId=5:Status=NONE:OldScore=0,NewScore:95",
			assignment: d,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 95, GroupID: 5},
			score:      95,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=90:AutoApprove:IsGroupLab,Submission:UserId=15:Status=NONE:OldScore=0,NewScore:50",
			assignment: e,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 50, UserID: 15},
			score:      50,
			expected:   qf.Submission_NONE,
		},
		{
			name:       "Assignment:ScoreLimit=90:AutoApprove:IsGroupLab,Submission:UserId=15:Status=NONE:OldScore=0,NewScore:95",
			assignment: e,
			submission: &qf.Submission{Status: qf.Submission_NONE, Score: 95, UserID: 15},
			score:      95,
			expected:   qf.Submission_NONE,
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
