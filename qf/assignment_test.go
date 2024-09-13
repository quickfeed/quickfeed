package qf_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSubmissionStatus(t *testing.T) {
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
		name := qtest.Name("User/"+test.name, []string{"AutoApprove", "ScoreLimit", "PrevStatus", "PrevScore", "Score"}, test.assignment.AutoApprove, test.assignment.ScoreLimit, test.submission.GetGrades(), test.submission.GetScore(), test.score)
		t.Run(name, func(t *testing.T) {
			got := test.assignment.SubmissionStatus(test.submission, test.score)
			if diff := cmp.Diff(got, test.want, protocmp.Transform()); diff != "" {
				t.Errorf("SubmissionStatus(%v, %v, %d) mismatch (-want +got):\n%s", test.assignment, test.submission, test.score, diff)
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
		name := qtest.Name("Group/"+test.name, []string{"AutoApprove", "ScoreLimit", "PrevStatus", "PrevScore", "Score"}, test.assignment.AutoApprove, test.assignment.ScoreLimit, test.submission.GetGrades(), test.submission.GetScore(), test.score)
		t.Run(name, func(t *testing.T) {
			got := test.assignment.SubmissionStatus(test.submission, test.score)
			if diff := cmp.Diff(got, test.want, protocmp.Transform()); diff != "" {
				t.Errorf("SubmissionStatus(%v, %v, %d) mismatch (-want +got):\n%s", test.assignment, test.submission, test.score, diff)
			}
		})
	}
}
