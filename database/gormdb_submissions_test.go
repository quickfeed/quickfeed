package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetCourseSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, individualAssignment := qtest.SetupCourseAssignment(t, db)
	groupAssignment := &qf.Assignment{
		CourseID:   course.GetID(),
		Order:      2,
		IsGroupLab: true,
	}
	qtest.CreateAssignment(t, db, groupAssignment)
	group := qtest.CreateGroup(t, db, &qf.Group{
		CourseID: course.GetID(),
		Name:     "Group 1",
		Users:    []*qf.User{user},
	})

	userSubmission := &qf.Submission{
		AssignmentID: individualAssignment.GetID(),
		UserID:       user.GetID(),
		Score:        42,
		Grades:       []*qf.Grade{{SubmissionID: 1, UserID: user.GetID()}},
	}
	groupSubmission := &qf.Submission{
		AssignmentID: groupAssignment.GetID(),
		GroupID:      group.GetID(),
		Score:        42,
		Grades:       []*qf.Grade{{SubmissionID: 1, UserID: user.GetID()}},
	}
	qtest.CreateSubmission(t, db, userSubmission)
	qtest.CreateSubmission(t, db, groupSubmission)

	tests := []struct {
		name             string
		request          *qf.SubmissionRequest
		want             []*qf.Submission
		submissionMapKey uint64 // A key per enrollment or group
	}{
		{
			name: "fetch user submission",
			request: &qf.SubmissionRequest{
				CourseID: course.GetID(),
				FetchMode: &qf.SubmissionRequest_Type{
					Type: qf.SubmissionRequest_USER,
				},
			},
			want:             []*qf.Submission{userSubmission},
			submissionMapKey: user.GetID(),
		},
		{
			name: "fetch group submission",
			request: &qf.SubmissionRequest{
				CourseID: course.GetID(),
				FetchMode: &qf.SubmissionRequest_Type{
					Type: qf.SubmissionRequest_GROUP,
				},
			},
			want:             []*qf.Submission{groupSubmission},
			submissionMapKey: group.GetID(),
		},
		{
			name: "fetch all submissions",
			request: &qf.SubmissionRequest{
				CourseID: course.GetID(),
				FetchMode: &qf.SubmissionRequest_Type{
					Type: qf.SubmissionRequest_ALL,
				},
			},
			want:             []*qf.Submission{userSubmission, groupSubmission},
			submissionMapKey: user.GetID(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			submissions, err := db.GetCourseSubmissions(test.request)
			if err != nil {
				t.Fatal(err)
			}
			// Map 1 is empty, so we map with 2 and index the submissions array
			qtest.Diff(t, "GetCourseSubmissions() mismatch", submissions.Submissions[test.submissionMapKey].Submissions, test.want, protocmp.Transform())
		})
	}
}

func TestGetCourseSubmissionsOmitsTestOutput(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := qtest.SetupCourseAssignment(t, db)
	submission := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Grades:       []*qf.Grade{{SubmissionID: 1, UserID: user.GetID()}},
		Scores: []*score.Score{{
			TestName:    "TestStack",
			MaxScore:    5,
			Weight:      1,
			TestDetails: "stack_test.go:12: Pop() = <nil>",
			TestOutput:  "stack: pushing 3 elements",
			Status:      score.TestStatus_FAILED,
			Elapsed:     0.25,
		}},
	}
	qtest.CreateSubmission(t, db, submission)

	courseSubmissions, err := db.GetCourseSubmissions(&qf.SubmissionRequest{
		CourseID:  course.GetID(),
		FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_USER},
	})
	if err != nil {
		t.Fatal(err)
	}
	// As in TestGetCourseSubmissions, the user's ID is also its enrollment ID.
	submissions := courseSubmissions.GetSubmissions()[user.GetID()].GetSubmissions()
	if len(submissions) != 1 || len(submissions[0].GetScores()) != 1 {
		t.Fatalf("got %v, want one submission with one score", submissions)
	}
	got := submissions[0].GetScores()[0]
	if got.GetTestOutput() != "" {
		t.Errorf("course submissions carry test output %q, want none", got.GetTestOutput())
	}
	if got.GetTestDetails() == "" || got.GetStatus() != score.TestStatus_FAILED || got.GetElapsed() != 0.25 {
		t.Errorf("course submissions lost what the results table shows: %v", got)
	}

	full, err := db.GetSubmission(&qf.Submission{ID: submission.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := full.GetScores()[0].GetTestOutput(), "stack: pushing 3 elements"; got != want {
		t.Errorf("full submission test output = %q, want %q", got, want)
	}
}
