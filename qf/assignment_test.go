package qf_test

import (
	"reflect"
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
	isApprovedTests := []struct {
		name       string
		assignment *qf.Assignment
		submission *qf.Submission
		score      uint32
		expected   []*qf.Grade
	}{
		{
			name:       "Assignment:ScoreLimit=0:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: a,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 50},
			score:      55,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: b,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 50},
			score:      55,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 50},
			score:      80,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=80,NewScore:75",
			assignment: b,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 80},
			score:      75,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=80,NewScore:85",
			assignment: b,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 80},
			score:      85,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=REJECTED:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_REJECTED}}, Score: 50},
			score:      80,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_REJECTED}},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=REVISION:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_REVISION}}, Score: 50},
			score:      80,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_REVISION}},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:80",
			assignment: b,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}, Score: 50},
			score:      80,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:55",
			assignment: c,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 50},
			score:      55,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:79",
			assignment: c,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 50},
			score:      79,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=50,NewScore:80",
			assignment: c,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}, Score: 50},
			score:      80,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:0",
			assignment: c,
			submission: &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}, Score: 50},
			score:      0,
			expected:   []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:80",
			assignment: c,
			submission: &qf.Submission{
				Grades: []*qf.Grade{
					{UserID: 1, Status: qf.Submission_APPROVED},
					{UserID: 2, Status: qf.Submission_APPROVED}},
				Score: 50,
			},
			score: 80,
			expected: []*qf.Grade{
				{UserID: 1, Status: qf.Submission_APPROVED},
				{UserID: 2, Status: qf.Submission_APPROVED},
			},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:0",
			assignment: c,
			submission: &qf.Submission{
				Grades: []*qf.Grade{
					{UserID: 1, Status: qf.Submission_NONE},
					{UserID: 2, Status: qf.Submission_APPROVED},
				},
				Score: 50,
			},
			score: 0,
			expected: []*qf.Grade{
				{UserID: 1, Status: qf.Submission_NONE},
				{UserID: 2, Status: qf.Submission_APPROVED},
			},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=APPROVED:OldScore=50,NewScore:80",
			assignment: c,
			submission: &qf.Submission{
				Grades: []*qf.Grade{
					{UserID: 1, Status: qf.Submission_NONE},
					{UserID: 2, Status: qf.Submission_NONE},
				},
				Score: 50,
			},
			score: 80,
			expected: []*qf.Grade{
				{UserID: 1, Status: qf.Submission_APPROVED},
				{UserID: 2, Status: qf.Submission_APPROVED},
			},
		},
		{
			name:       "Assignment:ScoreLimit=0:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:95",
			assignment: a,
			submission: &qf.Submission{
				Grades: []*qf.Grade{
					{UserID: 1, Status: qf.Submission_NONE},
					{UserID: 2, Status: qf.Submission_NONE},
					{UserID: 3, Status: qf.Submission_NONE},
				},
				Score: 50,
			},
			score: 95,
			expected: []*qf.Grade{
				{UserID: 1, Status: qf.Submission_NONE},
				{UserID: 2, Status: qf.Submission_NONE},
				{UserID: 3, Status: qf.Submission_NONE},
			},
		},
		{
			name:       "Assignment:ScoreLimit=80:NoAutoApprove,Submission:Status=NONE:OldScore=50,NewScore:95",
			assignment: b,
			submission: &qf.Submission{
				Grades: []*qf.Grade{
					{UserID: 1, Status: qf.Submission_NONE},
					{UserID: 2, Status: qf.Submission_NONE},
					{UserID: 3, Status: qf.Submission_NONE},
				},
				Score: 50,
			},
			score: 95,
			expected: []*qf.Grade{
				{UserID: 1, Status: qf.Submission_NONE},
				{UserID: 2, Status: qf.Submission_NONE},
				{UserID: 3, Status: qf.Submission_NONE},
			},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=0,NewScore:79",
			assignment: c,
			submission: &qf.Submission{
				Grades: []*qf.Grade{
					{UserID: 1, Status: qf.Submission_NONE},
					{UserID: 2, Status: qf.Submission_NONE},
					{UserID: 3, Status: qf.Submission_NONE},
				},
				Score: 0,
			},
			score: 79,
			expected: []*qf.Grade{
				{UserID: 1, Status: qf.Submission_NONE},
				{UserID: 2, Status: qf.Submission_NONE},
				{UserID: 3, Status: qf.Submission_NONE},
			},
		},
		{
			name:       "Assignment:ScoreLimit=80:AutoApprove,Submission:Status=NONE:OldScore=0,NewScore:100",
			assignment: c,
			submission: &qf.Submission{
				Grades: []*qf.Grade{
					{UserID: 1, Status: qf.Submission_NONE},
					{UserID: 2, Status: qf.Submission_NONE},
					{UserID: 3, Status: qf.Submission_NONE},
				},
				Score: 0,
			},
			score: 100,
			expected: []*qf.Grade{
				{UserID: 1, Status: qf.Submission_APPROVED},
				{UserID: 2, Status: qf.Submission_APPROVED},
				{UserID: 3, Status: qf.Submission_APPROVED},
			},
		},
	}

	for _, test := range isApprovedTests {
		t.Run(test.name, func(t *testing.T) {
			got := test.assignment.IsApproved(test.submission, test.score)
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("IsApproved(%v, %v, %d) = %v, expected %v", test.assignment, test.submission, test.score, got, test.expected)
			}
		})
	}
}
