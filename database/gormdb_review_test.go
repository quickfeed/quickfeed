package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateUpdateReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := setupCourseAssignment(t, db, false)

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
	}); err != nil {
		t.Fatal(err)
	}
	// confirm that the submission is in the database
	submissions, err := db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	review := &qf.Review{
		SubmissionID: 2,
		ReviewerID:   1,
		Feedback:     "my very good feedback",
		Ready:        false,
		Score:        95,
		Edited:       "last night",
		GradingBenchmarks: []*qf.GradingBenchmark{
			{
				ID:           1,
				AssignmentID: 1,
				ReviewID:     1,
				Heading:      "Major league baseball",
				Comment:      "wonders of the world",
				Criteria: []*qf.GradingCriterion{
					{
						ID:          1,
						BenchmarkID: 1,
						Points:      30,
						Description: "my description",
						Grade:       qf.GradingCriterion_PASSED,
						Comment:     "another comment",
					},
				},
			},
		},
	}

	submission := &qf.Submission{
		AssignmentID: 1,
		UserID:       1,
		Reviews:      []*qf.Review{review},
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

func updateSubmission(t *testing.T, db database.Database, wantReview *qf.Review) {
	if err := db.UpdateReview(wantReview); err != nil {
		t.Errorf("failed to update review: %v", err)
	}
	sub, err := db.GetSubmission(&qf.Submission{ID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(sub.Reviews) != 1 {
		t.Fatalf("have %d reviews want %d", len(sub.Reviews), 1)
	}
	var gotReview *qf.Review
	for _, r := range sub.GetReviews() {
		gotReview = r
		// fmt.Printf("sub %d: %+v, score: %d\n", sub.GetID(), r.GetReady(), r.GetScore())
	}
	if diff := cmp.Diff(gotReview, wantReview, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same review, but got (-got +want):\n%s", diff)
	}
}
