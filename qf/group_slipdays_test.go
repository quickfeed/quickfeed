package qf_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

var groupSlipTests = []struct {
	name        string
	labs        []*qf.Assignment
	submissions [][]int32
	remaining   [][]int32
}{
	{
		"One assignment with deadline two days ago, two submissions same day",
		[]*qf.Assignment{a(-2)},
		[][]int32{{0, 0}},
		[][]int32{{3, 3}},
	},
	{
		"One assignment with deadline in two days, two submissions same day",
		[]*qf.Assignment{a(2)},
		[][]int32{{0, 0}},
		[][]int32{{5, 5}},
	},
	{
		"One assignment with deadline in two days, five submissions one day apart",
		[]*qf.Assignment{a(2)},
		[][]int32{{0, 1, 1, 1, 1}},
		[][]int32{{5, 5, 5, 4, 3}},
	},
	{
		"One assignment with deadline in two days, ten submissions one day apart",
		[]*qf.Assignment{a(2)},
		[][]int32{{0, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		[][]int32{{5, 5, 5, 4, 3, 2, 1, 0, -1, -2}},
	},
	{
		"Four assignments with different deadlines, five or more submissions for each assignment",
		[]*qf.Assignment{a(0), a(2), a(5), a(20)},
		[][]int32{
			{0, 0, 0, 0, 0},
			{0, 1, 0, 1, 0},
			{0, 3, 1, 1, 1},
			{0, 10, 1, 1, 1, 1, 1},
		},
		[][]int32{
			{5, 5, 5, 5, 5},
			{5, 5, 5, 5, 5},
			{5, 5, 4, 3, 2},
			{2, 2, 2, 2, 1, 0, -1},
		},
	},
}

func TestGroupSlipDays(t *testing.T) {
	for _, sd := range groupSlipTests {
		testNow = time.Now()
		group := &qf.Group{
			ID:           1,
			CourseID:     course.GetID(),
			UsedSlipDays: make([]*qf.UsedSlipDays, 0),
		}

		for i := range sd.labs {
			t.Run(fmt.Sprintf("%s#%d", sd.name, i), func(t *testing.T) {
				if len(sd.submissions) != len(sd.remaining) {
					t.Fatalf("faulty test case: len(sd.submissions)=%d != len(sd.remaining)=%d", len(sd.submissions), len(sd.remaining))
				}
				sd.labs[i].ID = uint64(i + 1)
				for j := range sd.submissions[i] {
					if len(sd.submissions[i]) != len(sd.remaining[i]) {
						t.Fatalf("faulty test case: len(sd.submissions[%d])=%d != len(sd.remaining[%d])=%d", i, len(sd.submissions[i]), i, len(sd.remaining[i]))
					}

					// emulate advancing time for this submission
					testNow = testNow.Add(time.Duration(sd.submissions[i][j]) * days)
					submission := &qf.Submission{
						AssignmentID: sd.labs[i].GetID(),
						GroupID:      group.GetID(),
						Grades:       []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}},
						Score:        50,
						BuildInfo: &score.BuildInfo{
							BuildDate:      timestamppb.New(testNow),
							SubmissionDate: timestamppb.New(testNow),
						},
					}

					// functions to test
					err := group.UpdateSlipDays(sd.labs[i], submission)
					if err != nil {
						t.Fatal(err)
					}
					remaining := group.RemainingSlipDays(course)
					if remaining != sd.remaining[i][j] {
						t.Errorf("UpdateSlipDays(%q, %q, %q, %q) == %d, want %d", testNow.Format(qf.TimeLayout), sd.labs[i], submission, group, remaining, sd.remaining[i][j])
					}
				}
			})
		}
	}
}

func TestScoreLimitGroupSlipDays(t *testing.T) {
	testNow = time.Now()
	neg2, a2 := a(-2), a(2)

	scoreLimitSlipDayTests := []struct {
		name       string
		assignment *qf.Assignment
		submission *qf.Submission
		remaining  uint32
	}{
		{
			name:       "DeadlineNotPassed,NotApproved,NoScoreLimit",
			assignment: a2,
			submission: &qf.Submission{AssignmentID: a2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}}, Score: 50},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlineNotPassed,NotApproved,ScoreLimit",
			assignment: a2,
			submission: &qf.Submission{AssignmentID: a2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlineNotPassed,AllApproved,NoScoreLimit",
			assignment: a2,
			submission: &qf.Submission{AssignmentID: a2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}, {UserID: 2, Status: qf.Submission_APPROVED}}, Score: 50},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlineNotPassed,AllApproved,ScoreLimit",
			assignment: a2,
			submission: &qf.Submission{AssignmentID: a2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}, {UserID: 2, Status: qf.Submission_APPROVED}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlinePassed,NotApproved,NoScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}}, Score: 50},
			remaining:  course.GetSlipDays() - 2,
		},
		{
			name:       "DeadlinePassed,AllApproved,NoScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}, {UserID: 2, Status: qf.Submission_APPROVED}}, Score: 50},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlinePassed,NotApproved,ScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlinePassed,AllApproved,ScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), GroupID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}, {UserID: 2, Status: qf.Submission_APPROVED}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
	}

	for _, tt := range scoreLimitSlipDayTests {
		t.Run(tt.name, func(t *testing.T) {
			group := &qf.Group{
				ID:           1,
				CourseID:     course.GetID(),
				UsedSlipDays: make([]*qf.UsedSlipDays, 0),
			}
			tt.submission.BuildInfo = &score.BuildInfo{
				BuildDate:      timestamppb.New(testNow),
				SubmissionDate: timestamppb.New(testNow),
			}
			err := group.UpdateSlipDays(tt.assignment, tt.submission)
			if err != nil {
				t.Fatal(err)
			}
			group.SetSlipDays(course)
			if group.GetSlipDaysRemaining() != tt.remaining {
				t.Errorf("%s: got %d, want %d", tt.name, group.GetSlipDaysRemaining(), tt.remaining)
			}
		})
	}
}

func TestGracePeriodGroupSlipDays(t *testing.T) {
	testNow = time.Now()
	// Both cases use a deadline at testNow (just passed) so submission times
	// relative to the deadline are easy to reason about.
	deadlineNow := a(0)

	gracePeriodTests := []struct {
		name             string
		submissionOffset time.Duration // time after the deadline
		remaining        uint32
	}{
		{
			// 1h after deadline: within 2h grace period, no slip day charged
			name:             "SubmittedOneHourAfterDeadline_WithinGracePeriod",
			submissionOffset: time.Hour,
			remaining:        course.GetSlipDays(),
		},
		{
			// 3h after deadline: outside 2h grace period, 1 slip day charged
			name:             "SubmittedThreeHoursAfterDeadline_OutsideGracePeriod",
			submissionOffset: 3 * time.Hour,
			remaining:        course.GetSlipDays() - 1,
		},
	}

	for i, tt := range gracePeriodTests {
		t.Run(tt.name, func(t *testing.T) {
			deadlineNow.ID = uint64(i + 1)
			group := &qf.Group{
				ID:           1,
				CourseID:     course.GetID(),
				UsedSlipDays: make([]*qf.UsedSlipDays, 0),
			}
			submissionTime := testNow.Add(tt.submissionOffset)
			submission := &qf.Submission{
				AssignmentID: deadlineNow.GetID(),
				GroupID:      1,
				Grades:       []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}},
				Score:        50,
				BuildInfo: &score.BuildInfo{
					BuildDate:      timestamppb.New(submissionTime),
					SubmissionDate: timestamppb.New(submissionTime),
				},
			}
			err := group.UpdateSlipDays(deadlineNow, submission)
			if err != nil {
				t.Fatal(err)
			}
			group.SetSlipDays(course)
			if group.GetSlipDaysRemaining() != tt.remaining {
				t.Errorf("%s: got %d, want %d", tt.name, group.GetSlipDaysRemaining(), tt.remaining)
			}
		})
	}
}
