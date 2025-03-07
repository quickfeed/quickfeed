package qf_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	days = time.Duration(24 * time.Hour)
)

var (
	testNow = time.Now()

	course = &qf.Course{
		ID:       1,
		SlipDays: 5,
		Name:     "opsys",
	}

	a = func(daysFromNow int32) *qf.Assignment {
		return &qf.Assignment{
			CourseID:   course.GetID(),
			ScoreLimit: 60,
			Deadline:   timestamppb.New(testNow.Add(time.Duration(daysFromNow) * days)),
		}
	}
)

var slipTests = []struct {
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
		[][]int32{ // each number is a submission; if >0 the number is #days since previous submission; they carry over from one line to the next.
			{0, 0, 0, 0, 0},        // 0 ==> five submissions on the same day
			{0, 1, 0, 1, 0},        // 0+2 ==> 1 submission on 0th day, two submissions on day 1, and two submissions on day 2.
			{0, 3, 1, 1, 1},        // 2+6 ==> 1 submission on day 2 (carried over from previous line), one submission on day 2+3, one submission on day 2+3+1, and so on.
			{0, 10, 1, 1, 1, 1, 1}, // 2+6+15 ==> 1 submission on day 2+6, one submission on day 2+6+10, and so on. (23 days total)
		},
		[][]int32{
			{5, 5, 5, 5, 5},        // no slip days used
			{5, 5, 5, 5, 5},        // no slip days used
			{5, 5, 4, 3, 2},        // used 3 slip days
			{2, 2, 2, 2, 1, 0, -1}, // used up all slip days and one more
		},
	},
}

func TestSlipDays(t *testing.T) {
	for _, sd := range slipTests {
		testNow = time.Now()
		enrol := &qf.Enrollment{
			Course:       course,
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
						Grades:       []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
						Score:        50,
						BuildInfo: &score.BuildInfo{
							BuildDate:      timestamppb.New(testNow),
							SubmissionDate: timestamppb.New(testNow),
						},
					}

					// functions to test
					err := enrol.UpdateSlipDays(sd.labs[i], submission)
					if err != nil {
						t.Fatal(err)
					}
					remaining := enrol.RemainingSlipDays(course)
					if remaining != sd.remaining[i][j] {
						t.Errorf("UpdateSlipDays(%q, %q, %q, %q) == %d, want %d", testNow.Format(qf.TimeLayout), sd.labs[i], submission, enrol, remaining, sd.remaining[i][j])
					}
				}
			})
		}
	}
}

func TestScoreLimitSlipDays(t *testing.T) {
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
			submission: &qf.Submission{AssignmentID: a2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 50},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlineNotPassed,NotApproved,ScoreLimit",
			assignment: a2,
			submission: &qf.Submission{AssignmentID: a2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlineNotPassed,Approved,NoScoreLimit",
			assignment: a2,
			submission: &qf.Submission{AssignmentID: a2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}, Score: 50},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlineNotPassed,Approved,ScoreLimit",
			assignment: a2,
			submission: &qf.Submission{AssignmentID: a2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlinePassed,NotApproved,NoScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 50},
			remaining:  course.GetSlipDays() - 2,
		},
		{
			name:       "DeadlinePassed,Approved,NoScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}, Score: 50},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlinePassed,NotApproved,ScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
		{
			name:       "DeadlinePassed,Approved,ScoreLimit",
			assignment: neg2,
			submission: &qf.Submission{AssignmentID: neg2.GetID(), Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}, Score: 60},
			remaining:  course.GetSlipDays(),
		},
	}
	for _, test := range scoreLimitSlipDayTests {
		enrol := &qf.Enrollment{
			Course:       course,
			CourseID:     course.GetID(),
			UsedSlipDays: make([]*qf.UsedSlipDays, 0),
			UserID:       1,
		}
		t.Run(test.name, func(t *testing.T) {
			test.submission.BuildInfo = &score.BuildInfo{
				BuildDate:      timestamppb.New(testNow),
				SubmissionDate: timestamppb.New(testNow),
			}
			err := enrol.UpdateSlipDays(test.assignment, test.submission)
			if err != nil {
				t.Fatal(err)
			}
			remaining := enrol.RemainingSlipDays(course)
			if uint32(remaining) != test.remaining {
				t.Errorf("UpdateSlipDays(%q, %q, %q, %q) = %d, want %d", testNow.Format(qf.TimeLayout), test.assignment, test.submission, enrol, remaining, test.remaining)
			}
		})
	}
}

func TestMismatchingAssignmentID(t *testing.T) {
	enrol := &qf.Enrollment{
		Course:       course,
		CourseID:     course.GetID(),
		UsedSlipDays: make([]*qf.UsedSlipDays, 0),
		UserID:       1,
	}
	// lab1's deadline is incorrectly formatted
	lab1 := &qf.Assignment{
		CourseID: course.GetID(),
		Deadline: timestamppb.New(testNow.Add(time.Duration(2) * days)),
	}
	lab1.ID = 1
	submission := &qf.Submission{
		Grades:       []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		AssignmentID: lab1.GetID() + 1,
		BuildInfo: &score.BuildInfo{
			BuildDate:      timestamppb.New(testNow),
			SubmissionDate: timestamppb.New(testNow),
		},
	}
	err := enrol.UpdateSlipDays(lab1, submission)
	if err == nil {
		t.Errorf("expected invariant violation since (assignment.GetID() != submission.GetAssignmentID())")
	}
}

func TestMismatchingCourseID(t *testing.T) {
	enrol := &qf.Enrollment{
		Course:       course,
		CourseID:     course.GetID(),
		UsedSlipDays: make([]*qf.UsedSlipDays, 0),
		UserID:       1,
	}
	// lab1's deadline is incorrectly formatted
	lab1 := &qf.Assignment{
		CourseID: course.GetID() + 1,
		Deadline: timestamppb.New(testNow.Add(time.Duration(2) * days)),
	}
	lab1.ID = 1
	submission := &qf.Submission{
		AssignmentID: lab1.GetID(),
		Grades:       []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		BuildInfo: &score.BuildInfo{
			BuildDate:      timestamppb.New(testNow),
			SubmissionDate: timestamppb.New(testNow),
		},
	}
	err := enrol.UpdateSlipDays(lab1, submission)
	if err == nil {
		t.Errorf("expected invariant violation since (enrollment.GetCourseID() != assignment.GetCourseID())")
	}
}

func TestEnrollmentGetUsedSlipDays(t *testing.T) {
	enrol := &qf.Enrollment{
		Course:       course,
		CourseID:     course.GetID(),
		UsedSlipDays: make([]*qf.UsedSlipDays, 0),
		UserID:       1,
	}
	// lab1's deadline passed two days ago
	lab1 := a(-2)
	lab1.ID = 1
	submission := &qf.Submission{
		AssignmentID: lab1.GetID(),
		Grades:       []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}},
		BuildInfo: &score.BuildInfo{
			BuildDate:      timestamppb.New(testNow),
			SubmissionDate: timestamppb.New(testNow),
		},
	}
	usedSlipDays := enrol.GetUsedSlipDays()
	if len(usedSlipDays) != 0 {
		t.Errorf("len(usedSlipDays) = %d, expected 0", len(usedSlipDays))
	}
	err := enrol.UpdateSlipDays(lab1, submission)
	if err != nil {
		t.Error(err)
	}
	usedSlipDays = enrol.GetUsedSlipDays()
	if len(usedSlipDays) != 1 {
		t.Errorf("len(usedSlipDays) = %d, expected 1", len(usedSlipDays))
	}
	wantUsedSlipDays := []*qf.UsedSlipDays{
		{
			AssignmentID: 1,
			UsedDays:     2,
		},
	}
	if diff := cmp.Diff(wantUsedSlipDays, usedSlipDays, protocmp.Transform()); diff != "" {
		t.Errorf("GetUsedSlipDays() mismatch (-want +got):\n%s", diff)
	}
}

func TestSlipDaysWGracePeriod(t *testing.T) {
	lab := a(0)
	lab.ID = 1
	timeOfDeadline := lab.GetDeadline().AsTime()
	submission := &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}}, AssignmentID: lab.GetID()}
	submissionTimes := []struct {
		delivered    time.Time
		comment      string
		wantSlipDays uint32
	}{
		{
			delivered:    timeOfDeadline.Add(time.Duration(15) * time.Minute),
			comment:      "Delivered 15 minutes after the deadline",
			wantSlipDays: 0,
		},
		{
			delivered:    timeOfDeadline.Add(time.Duration(119) * time.Minute),
			comment:      "Delivered 1 hour and 59 minutes after the deadline",
			wantSlipDays: 0,
		},
		{
			delivered:    timeOfDeadline.Add(time.Duration(2) * time.Hour),
			comment:      "Delivered exactly 2 hours after the deadline",
			wantSlipDays: 0,
		},
		{
			delivered:    timeOfDeadline.Add(time.Duration(2)*time.Hour + time.Second),
			comment:      "Delivered 2 hours and 1 second after the deadline",
			wantSlipDays: 1,
		},
		{
			delivered:    timeOfDeadline.Add(days + time.Hour),
			comment:      "Delivered 1 day and 1 hour after the deadline",
			wantSlipDays: 1,
		},
		{
			delivered:    timeOfDeadline.Add(days + 3*time.Hour),
			comment:      "Delivered 1 day and 3 hours after the deadline",
			wantSlipDays: 2,
		},
		{
			delivered:    timeOfDeadline.Add(3*days + 6*time.Hour),
			comment:      "Delivered 3 days and 6 hours after the deadline",
			wantSlipDays: 4,
		},
	}

	for _, test := range submissionTimes {
		enrol := &qf.Enrollment{
			Course:       course,
			CourseID:     course.GetID(),
			UsedSlipDays: make([]*qf.UsedSlipDays, 0),
			UserID:       1,
		}
		t.Run(fmt.Sprintf("%s/Want UsedSlipDays:%d", test.comment, test.wantSlipDays), func(t *testing.T) {
			submission.BuildInfo = &score.BuildInfo{
				BuildDate:      timestamppb.New(test.delivered),
				SubmissionDate: timestamppb.New(test.delivered),
			}
			err := enrol.UpdateSlipDays(lab, submission)
			if err != nil {
				t.Fatal(err)
			}
			var usedSlipDays uint32
			for _, days := range enrol.GetUsedSlipDays() {
				usedSlipDays += days.GetUsedDays()
			}
			if usedSlipDays != test.wantSlipDays {
				t.Errorf("UpdateSlipDays('%v', '%v', '%v') = %d, want %d", test.delivered, lab, submission, usedSlipDays, test.wantSlipDays)
			}
		})
	}
}
