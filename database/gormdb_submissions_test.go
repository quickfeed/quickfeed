package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetCourseSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// create teacher, course, user (student) and assignment
	user, course, assignment := setupCourseAssignment(t, db)

	wantStruct := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        42,
		Reviews:      []*qf.Review{},
		BuildInfo: &score.BuildInfo{
			BuildDate:      qtest.Timestamp(t, "2021-01-21T18:00:00"),
			SubmissionDate: qtest.Timestamp(t, "2021-01-21T18:00:00"),
			BuildLog:       "what do you say",
			ExecTime:       50,
		},
		Scores: []*score.Score{
			{TestName: "TestBigNum", MaxScore: 100, Score: 60, Weight: 10},
			{TestName: "TestDigNum", MaxScore: 100, Score: 70, Weight: 10},
		},
	}
	if err := db.CreateSubmission(wantStruct); err != nil {
		t.Fatal(err)
	}
	request := &qf.SubmissionRequest{
		CourseID: course.GetID(),
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}
	submissions, err := db.GetCourseSubmissions(request)
	if err != nil {
		t.Fatal(err)
	}
	wantStruct.BuildInfo = nil
	wantAssignment := (proto.Clone(assignment)).(*qf.Assignment)
	wantAssignment.Submissions = append(wantAssignment.GetSubmissions(), wantStruct)
	if diff := cmp.Diff(wantAssignment.GetSubmissions(), submissions, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourseSubmissions() mismatch (-want +got):\n%s", diff)
	}

	// Submission with Review
	wantReview := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        45,
		Reviews: []*qf.Review{
			{
				ReviewerID: 1, Feedback: "SGTM!", Score: 42, Ready: true,
				GradingBenchmarks: []*qf.GradingBenchmark{
					{
						Heading: "Ding Dong", Comment: "Communication",
						Criteria: []*qf.GradingCriterion{
							{Points: 50, Description: "Loads of ding"},
						},
					},
				},
			},
		},
	}
	if err := db.CreateSubmission(wantReview); err != nil {
		t.Fatal(err)
	}
	request = &qf.SubmissionRequest{
		CourseID: course.GetID(),
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}
	submissions, err = db.GetCourseSubmissions(request)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment = (proto.Clone(assignment)).(*qf.Assignment)
	wantAssignment.Submissions = append(wantAssignment.GetSubmissions(), wantStruct, wantReview)
	if diff := cmp.Diff(wantAssignment.GetSubmissions(), submissions, protocmp.Transform()); diff != "" {
		t.Errorf("GetCourseSubmissions() mismatch (-want +got):\n%s", diff)
	}
}
