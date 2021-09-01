package database_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGormDBRemoveTest(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, course, assignment := setupCourseAssignment(t, db)

	// create a new submission, ensure that build info and scores are saved as well
	buildInfo := &score.BuildInfo{
		BuildDate: "2022-11-10T13:00:00",
		BuildLog:  "Testing",
		ExecTime:  33333,
	}
	scores := []*score.Score{
		{TestName: "Test1", Score: 10, MaxScore: 15, Weight: 1},
		{TestName: "Test2", Score: 0, MaxScore: 5, Weight: 1},
		{TestName: "Test3", Score: 3, MaxScore: 5, Weight: 1},
	}
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		BuildInfo:    buildInfo,
		Scores:       scores,
	}); err != nil {
		t.Fatal(err)
	}
	submissions, err := db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	buildInfo.SubmissionID = submissions[0].ID
	buildInfo.ID = 1
	if diff := cmp.Diff(buildInfo, submissions[0].BuildInfo, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same build info, but got (-got +want):\n%s", diff)
	}
	if diff := cmp.Diff(
		submissions[0].Scores,
		scores,
		protocmp.Transform(),
		protocmp.IgnoreFields(&score.Score{}, "ID", "SubmissionID")); diff != "" {
		t.Errorf("Incorrect scores after first save (-want, +got):\n%s", diff)
	}

	// buildInfo record must be updated (have the same ID as before) instead
	// of saving a duplicate
	oldSubmissionID := submissions[0].ID
	updatedBuildInfo := &score.BuildInfo{
		BuildDate: "2022-11-10T15:00:00",
		BuildLog:  "Updated",
		ExecTime:  12345,
	}
	scores = []*score.Score{
		{TestName: "Test1", Score: 10, MaxScore: 15, Weight: 1},
		// Test2 is removed from the tests repository and should be removed in the output from GetLastSubmissions
		{TestName: "Test3", Score: 3, MaxScore: 5, Weight: 1},
	}

	submissions[0].BuildInfo = updatedBuildInfo
	submissions[0].Scores = scores
	if err := db.CreateSubmission(submissions[0]); err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	updatedBuildInfo.ID = submissions[0].BuildInfo.ID
	updatedBuildInfo.SubmissionID = oldSubmissionID
	if diff := cmp.Diff(submissions[0].BuildInfo, updatedBuildInfo, protocmp.Transform()); diff != "" {
		t.Errorf("Expected updated build info, but got (-sub +want):\n%s", diff)
	}
	if diff := cmp.Diff(submissions[0].Scores, scores, protocmp.Transform()); diff != "" {
		t.Errorf("Incorrect scores after update (-want, +got):\n%s", diff)
	}

	// attempting to update build info and scores with wrong submission ID must return an error
	submissions[0].ID = 123
	if err := db.CreateSubmission(submissions[0]); err == nil {
		t.Fatal("expected error: record not found")
	}
}
