package qf_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// newGroup returns a group enrolled in the shared test course, initialized for slip-day tests.
func newGroup() *qf.Group {
	return &qf.Group{
		ID:           1,
		CourseID:     course.GetID(),
		UsedSlipDays: make([]*qf.UsedSlipDays, 0),
	}
}

// TestGroupApprovalRequiresAllMembers checks the one behavior that actually differs
// between *Group and *Enrollment slip-day accrual: a group submission only counts as
// approved (and thus stops accruing slip days) once every member's grade is approved,
// whereas an enrollment only checks a single user's grade (see submission.IsAllApproved
// vs submission.IsApproved).
func TestGroupApprovalRequiresAllMembers(t *testing.T) {
	testNow = time.Now()
	deadlinePassed := a(-2)
	deadlinePassed.ID = 1

	tests := []struct {
		name      string
		grades    []*qf.Grade
		remaining uint32
	}{
		{
			name:      "NoMembersApproved",
			grades:    []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}},
			remaining: course.GetSlipDays() - 2,
		},
		{
			name:      "OneMemberApproved",
			grades:    []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}, {UserID: 2, Status: qf.Submission_NONE}},
			remaining: course.GetSlipDays() - 2,
		},
		{
			name:      "AllMembersApproved",
			grades:    []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}, {UserID: 2, Status: qf.Submission_APPROVED}},
			remaining: course.GetSlipDays(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := newGroup()
			submission := &qf.Submission{
				AssignmentID: deadlinePassed.GetID(),
				GroupID:      group.GetID(),
				Grades:       tt.grades,
				Score:        50,
				BuildInfo: &score.BuildInfo{
					BuildDate:      timestamppb.New(testNow),
					SubmissionDate: timestamppb.New(testNow),
				},
			}
			if err := group.UpdateSlipDays(deadlinePassed, submission); err != nil {
				t.Fatal(err)
			}
			group.SetSlipDays(course)
			if group.GetSlipDaysRemaining() != tt.remaining {
				t.Errorf("%s: SlipDaysRemaining() = %d, want %d", tt.name, group.GetSlipDaysRemaining(), tt.remaining)
			}
		})
	}
}

// TestGroupGetUsedSlipDays checks that UpdateSlipDays records used slip days against
// the group (GroupID set), which is the other behavior that differs from *Enrollment
// (which sets EnrollmentID instead; see TestEnrollmentGetUsedSlipDays).
func TestGroupGetUsedSlipDays(t *testing.T) {
	testNow = time.Now()
	group := newGroup()
	lab := a(-2)
	lab.ID = 1
	submission := &qf.Submission{
		AssignmentID: lab.GetID(),
		GroupID:      group.GetID(),
		Grades:       []*qf.Grade{{UserID: 1, Status: qf.Submission_NONE}, {UserID: 2, Status: qf.Submission_NONE}},
		BuildInfo: &score.BuildInfo{
			BuildDate:      timestamppb.New(testNow),
			SubmissionDate: timestamppb.New(testNow),
		},
	}
	if err := group.UpdateSlipDays(lab, submission); err != nil {
		t.Fatal(err)
	}
	want := []*qf.UsedSlipDays{
		{
			AssignmentID: 1,
			GroupID:      group.GetID(),
			UsedDays:     2,
		},
	}
	if diff := cmp.Diff(want, group.GetUsedSlipDays(), protocmp.Transform()); diff != "" {
		t.Errorf("GetUsedSlipDays() mismatch (-want +got):\n%s", diff)
	}
}
