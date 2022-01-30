package database_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateUpdateReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := setupCourseAssignment(t, db)

	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
	}); err != nil {
		t.Fatal(err)
	}
	// confirm that the submission is in the database
	submissions, err := db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	review := &pb.Review{
		SubmissionID: 2,
		ReviewerID:   1,
		Feedback:     "my very good feedback",
		Ready:        false,
		Score:        95,
		Edited:       "last night",
		GradingBenchmarks: []*pb.GradingBenchmark{
			{
				ID:           1,
				AssignmentID: 1,
				ReviewID:     1,
				Heading:      "Major league baseball",
				Comment:      "wonders of the world",
				Criteria: []*pb.GradingCriterion{
					{
						ID:          1,
						BenchmarkID: 1,
						Points:      30,
						Description: "my description",
						Grade:       pb.GradingCriterion_PASSED,
						Comment:     "another comment",
					},
				},
			},
		},
	}

	submission := &pb.Submission{
		AssignmentID: 1,
		UserID:       1,
		Reviews:      []*pb.Review{review},
	}
	if err := db.CreateSubmission(submission); err != nil {
		t.Fatal(err)
	}
	updateSubmission(t, db, review)

	review.Edited = "today"
	review.Score = 90
	review.Ready = true
	updateSubmission(t, db, review)

	review.Edited = "now"
	review.Score = 50
	review.Ready = false
	updateSubmission(t, db, review)
}

func updateSubmission(t *testing.T, db database.Database, wantReview *pb.Review) {
	if err := db.UpdateReview(wantReview); err != nil {
		t.Errorf("failed to update review: %v", err)
	}
	sub, err := db.GetSubmission(&pb.Submission{ID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(sub.Reviews) != 1 {
		t.Fatalf("have %d reviews want %d", len(sub.Reviews), 1)
	}
	var gotReview *pb.Review
	for _, r := range sub.GetReviews() {
		gotReview = r
		// fmt.Printf("sub %d: %+v, score: %d\n", sub.GetID(), r.GetReady(), r.GetScore())
	}
	if diff := cmp.Diff(gotReview, wantReview, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same review, but got (-got +want):\n%s", diff)
	}
}
