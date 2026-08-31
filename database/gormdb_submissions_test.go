package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
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
