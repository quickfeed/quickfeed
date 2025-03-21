package qf_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewestSubmissionDate(t *testing.T) {
	submission := &qf.Submission{}
	tim := time.Now()
	newSubmissionDate := submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: &timestamppb.Timestamp{},
		},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	// Seems like the conversion from time.Time to timestamppb.Timestamp is not
	// exact, so we need to add a second to make sure the build date is newer.
	buildDate := time.Now().Add(1 * time.Second)
	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: timestamppb.New(buildDate),
		},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.After(buildDate) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}

	zero := time.Time{}
	newSubmissionDate = submission.NewestSubmissionDate(zero)
	if newSubmissionDate.Equal(zero) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.Before(zero) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.After(buildDate) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.Before(zero) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
}

func TestByUser(t *testing.T) {
	submission := &qf.Submission{}
	if submission.ByUser(0) {
		t.Errorf("ByUser(0) = true, expected false\n")
	}

	submission = &qf.Submission{
		UserID: 1,
	}
	if !submission.ByUser(1) {
		t.Errorf("ByUser(1) = false, expected true\n")
	}

	submission = &qf.Submission{
		GroupID: 1,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &qf.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}
}

func TestByGroup(t *testing.T) {
	submission := &qf.Submission{}
	if submission.ByGroup(0) {
		t.Errorf("ByGroup(0) = true, expected false\n")
	}

	submission = &qf.Submission{
		GroupID: 1,
	}
	if !submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = false, expected true\n")
	}

	submission = &qf.Submission{
		UserID: 1,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &qf.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}
}

func TestUpdateTotalApproved(t *testing.T) {
	enroll1 := &qf.Enrollment{ID: 10, UserID: 1}
	enroll2 := &qf.Enrollment{ID: 20, UserID: 2}
	enroll3 := &qf.Enrollment{ID: 30, UserID: 3}
	enroll4 := &qf.Enrollment{ID: 40, UserID: 4}

	submissions := qf.CourseSubmissions{
		Submissions: map[uint64]*qf.Submissions{
			enroll1.GetID(): {
				Submissions: []*qf.Submission{
					// total approved = 3
					{ID: 1, AssignmentID: 1, UserID: enroll1.GetUserID(), Grades: []*qf.Grade{{UserID: enroll1.GetUserID(), Status: qf.Submission_APPROVED}}},
					{ID: 2, AssignmentID: 2, GroupID: 20, Grades: []*qf.Grade{{UserID: enroll1.GetUserID(), Status: qf.Submission_APPROVED}}},
					{ID: 3, AssignmentID: 3, UserID: enroll1.GetUserID(), Grades: []*qf.Grade{{UserID: 3, Status: qf.Submission_APPROVED}}},
					// duplicate approved assignment should be ignored
					{ID: 2, AssignmentID: 3, UserID: enroll1.GetUserID(), Grades: []*qf.Grade{{UserID: enroll1.GetUserID(), Status: qf.Submission_APPROVED}}},
				},
			},
			enroll2.GetID(): {
				Submissions: []*qf.Submission{
					// total approved = 4
					{ID: 1, AssignmentID: 1, GroupID: 30, Grades: []*qf.Grade{{UserID: enroll2.GetUserID(), Status: qf.Submission_APPROVED}}},
					{ID: 2, AssignmentID: 2, UserID: enroll2.GetUserID(), Grades: []*qf.Grade{{UserID: enroll2.GetUserID(), Status: qf.Submission_APPROVED}}},
					{ID: 3, AssignmentID: 3, UserID: enroll2.GetUserID(), Grades: []*qf.Grade{{UserID: enroll2.GetUserID(), Status: qf.Submission_APPROVED}}},
					{ID: 4, AssignmentID: 4, UserID: enroll2.GetUserID(), Grades: []*qf.Grade{{UserID: enroll2.GetUserID(), Status: qf.Submission_APPROVED}}},
				},
			},
			enroll3.GetID(): {
				Submissions: []*qf.Submission{
					// total approved = 1
					{ID: 1, AssignmentID: 1, UserID: enroll3.GetUserID(), Grades: []*qf.Grade{
						{UserID: enroll3.GetUserID(), Status: qf.Submission_APPROVED},
						// duplicate grade should be ignored
						{UserID: enroll3.GetUserID(), Status: qf.Submission_APPROVED},
					}},
				},
			},
			enroll4.GetID(): {
				Submissions: []*qf.Submission{
					// total approved = 1
					{ID: 1, AssignmentID: 1, UserID: enroll4.GetUserID(), Grades: []*qf.Grade{{UserID: enroll4.GetUserID(), Status: qf.Submission_APPROVED}}},
					// duplicate assignment should be ignored
					{ID: 1, AssignmentID: 1, GroupID: 40, Grades: []*qf.Grade{{UserID: enroll4.GetUserID(), Status: qf.Submission_APPROVED}}},
					{ID: 2, AssignmentID: 2, UserID: enroll4.GetUserID(), Grades: []*qf.Grade{{UserID: enroll4.GetUserID(), Status: qf.Submission_NONE}}},
					// user has no grade for this assignment
					{ID: 3, AssignmentID: 3, GroupID: 40, Grades: []*qf.Grade{{UserID: 10, Status: qf.Submission_APPROVED}}},
				},
			},
		},
	}

	tests := []*struct {
		enrollment *qf.Enrollment
		want       uint64
	}{
		{enroll1, 3},
		{enroll2, 4},
		{enroll3, 1},
		{enroll4, 1},
	}

	for _, test := range tests {
		enrollment := test.enrollment
		enrollment.UpdateTotalApproved(submissions.For(enrollment.GetID()))
		if enrollment.GetTotalApproved() != test.want {
			t.Errorf("expected enrollment(id=%d) total approved %d, got %d", enrollment.GetID(), test.want, enrollment.GetTotalApproved())
		}
	}
}

func TestSetGradesIfApproved(t *testing.T) {
	const (
		T = true
		F = false
	)
	auto := &qf.Assignment{AutoApprove: T, ScoreLimit: 80}
	manual := &qf.Assignment{AutoApprove: F, ScoreLimit: 80}
	sub := func(status qf.Submission_Status, score uint32) *qf.Submission {
		return &qf.Submission{Grades: []*qf.Grade{{UserID: 1, Status: status}}, Score: score}
	}
	grade := func(status qf.Submission_Status) []*qf.Grade {
		return []*qf.Grade{{UserID: 1, Status: status}}
	}
	tests := []struct {
		name       string
		assignment *qf.Assignment
		submission *qf.Submission
		score      uint32
		want       []*qf.Grade
	}{
		// Nil submission
		{name: "NilSubmission", assignment: auto, submission: nil, score: 85, want: nil},
		{name: "NilSubmission", assignment: manual, submission: nil, score: 85, want: nil},
		{name: "NilSubmission", assignment: auto, submission: nil, score: 75, want: nil},
		{name: "NilSubmission", assignment: manual, submission: nil, score: 75, want: nil},
		// AutoApprove = true
		{name: "Approved", assignment: auto, submission: sub(qf.Submission_NONE, 85), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "None", assignment: auto, submission: sub(qf.Submission_NONE, 75), score: 75, want: grade(qf.Submission_NONE)},
		{name: "AlreadyApproved", assignment: auto, submission: sub(qf.Submission_APPROVED, 75), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: auto, submission: sub(qf.Submission_REVISION, 75), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyRejected", assignment: auto, submission: sub(qf.Submission_REJECTED, 75), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyApproved", assignment: auto, submission: sub(qf.Submission_APPROVED, 85), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: auto, submission: sub(qf.Submission_REVISION, 85), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyRejected", assignment: auto, submission: sub(qf.Submission_REJECTED, 85), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyApproved", assignment: auto, submission: sub(qf.Submission_APPROVED, 75), score: 79, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: auto, submission: sub(qf.Submission_REVISION, 75), score: 79, want: grade(qf.Submission_REVISION)},
		{name: "AlreadyRejected", assignment: auto, submission: sub(qf.Submission_REJECTED, 75), score: 79, want: grade(qf.Submission_REJECTED)},
		// AutoApprove = false
		{name: "None", assignment: manual, submission: sub(qf.Submission_NONE, 85), score: 85, want: grade(qf.Submission_NONE)},
		{name: "None", assignment: manual, submission: sub(qf.Submission_NONE, 75), score: 75, want: grade(qf.Submission_NONE)},
		{name: "AlreadyApproved", assignment: manual, submission: sub(qf.Submission_APPROVED, 85), score: 85, want: grade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: manual, submission: sub(qf.Submission_REVISION, 85), score: 85, want: grade(qf.Submission_REVISION)},
		{name: "AlreadyRejected", assignment: manual, submission: sub(qf.Submission_REJECTED, 85), score: 85, want: grade(qf.Submission_REJECTED)},
	}
	for _, test := range tests {
		name := qtest.Name("User/"+test.name, []string{"AutoApprove", "ScoreLimit", "PrevStatus", "PrevScore", "Score"}, test.assignment.GetAutoApprove(), test.assignment.GetScoreLimit(), test.submission.GetGrades(), test.submission.GetScore(), test.score)
		t.Run(name, func(t *testing.T) {
			sub := test.submission
			sub.SetGradesIfApproved(test.assignment, test.score)
			got := sub.GetGrades()
			if diff := cmp.Diff(got, test.want, protocmp.Transform()); diff != "" {
				t.Errorf("SetGradesIfApproved(%v, %v, %d) mismatch (-want +got):\n%s", test.assignment, test.submission, test.score, diff)
			}
		})
	}

	groupAuto := &qf.Assignment{AutoApprove: T, ScoreLimit: 80, IsGroupLab: T}
	groupManual := &qf.Assignment{AutoApprove: F, ScoreLimit: 80, IsGroupLab: T}
	groupSub := func(status qf.Submission_Status, score uint32) *qf.Submission {
		return &qf.Submission{Grades: []*qf.Grade{
			{UserID: 1, Status: status},
			{UserID: 2, Status: status},
			{UserID: 3, Status: status},
		}, Score: score, GroupID: 1}
	}
	groupGrade := func(status qf.Submission_Status) []*qf.Grade {
		return []*qf.Grade{
			{UserID: 1, Status: status},
			{UserID: 2, Status: status},
			{UserID: 3, Status: status},
		}
	}
	groupTests := []struct {
		name       string
		assignment *qf.Assignment
		submission *qf.Submission
		score      uint32
		want       []*qf.Grade
	}{
		// Nil submission
		{name: "NilSubmission", assignment: groupAuto, submission: nil, score: 85, want: nil},
		{name: "NilSubmission", assignment: groupManual, submission: nil, score: 85, want: nil},
		{name: "NilSubmission", assignment: groupAuto, submission: nil, score: 75, want: nil},
		{name: "NilSubmission", assignment: groupManual, submission: nil, score: 75, want: nil},
		// AutoApprove = true
		{name: "Approved", assignment: groupAuto, submission: groupSub(qf.Submission_NONE, 85), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "None", assignment: groupAuto, submission: groupSub(qf.Submission_NONE, 75), score: 75, want: groupGrade(qf.Submission_NONE)},
		{name: "AlreadyApproved", assignment: groupAuto, submission: groupSub(qf.Submission_APPROVED, 75), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: groupAuto, submission: groupSub(qf.Submission_REVISION, 75), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyRejected", assignment: groupAuto, submission: groupSub(qf.Submission_REJECTED, 75), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyApproved", assignment: groupAuto, submission: groupSub(qf.Submission_APPROVED, 85), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: groupAuto, submission: groupSub(qf.Submission_REVISION, 85), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyRejected", assignment: groupAuto, submission: groupSub(qf.Submission_REJECTED, 85), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyApproved", assignment: groupAuto, submission: groupSub(qf.Submission_APPROVED, 75), score: 79, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: groupAuto, submission: groupSub(qf.Submission_REVISION, 75), score: 79, want: groupGrade(qf.Submission_REVISION)},
		{name: "AlreadyRejected", assignment: groupAuto, submission: groupSub(qf.Submission_REJECTED, 75), score: 79, want: groupGrade(qf.Submission_REJECTED)},
		// AutoApprove = false
		{name: "None", assignment: groupManual, submission: groupSub(qf.Submission_NONE, 85), score: 85, want: groupGrade(qf.Submission_NONE)},
		{name: "None", assignment: groupManual, submission: groupSub(qf.Submission_NONE, 75), score: 75, want: groupGrade(qf.Submission_NONE)},
		{name: "AlreadyApproved", assignment: groupManual, submission: groupSub(qf.Submission_APPROVED, 85), score: 85, want: groupGrade(qf.Submission_APPROVED)},
		{name: "AlreadyRevision", assignment: groupManual, submission: groupSub(qf.Submission_REVISION, 85), score: 85, want: groupGrade(qf.Submission_REVISION)},
		{name: "AlreadyRejected", assignment: groupManual, submission: groupSub(qf.Submission_REJECTED, 85), score: 85, want: groupGrade(qf.Submission_REJECTED)},
	}
	for _, test := range groupTests {
		name := qtest.Name("Group/"+test.name, []string{"AutoApprove", "ScoreLimit", "PrevStatus", "PrevScore", "Score"}, test.assignment.GetAutoApprove(), test.assignment.GetScoreLimit(), test.submission.GetGrades(), test.submission.GetScore(), test.score)
		t.Run(name, func(t *testing.T) {
			sub := test.submission
			sub.SetGradesIfApproved(test.assignment, test.score)
			got := sub.GetGrades()
			if diff := cmp.Diff(got, test.want, protocmp.Transform()); diff != "" {
				t.Errorf("SetGradesIfApproved(%v, %v, %d) mismatch (-want +got):\n%s", test.assignment, test.submission, test.score, diff)
			}
		})
	}
}
