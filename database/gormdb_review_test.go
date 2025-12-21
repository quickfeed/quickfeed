package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func TestCreateReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := qtest.SetupCourseAssignment(t, db)
	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submission)

	assignmentWithReviewers := &qf.Assignment{CourseID: course.GetID(), Order: 2, Reviewers: 1}
	qtest.CreateAssignment(t, db, assignmentWithReviewers)
	submissionWithReviewers := &qf.Submission{AssignmentID: assignmentWithReviewers.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submissionWithReviewers)

	criteria := []*qf.GradingCriterion{{Points: 30, Description: "my description", Grade: qf.GradingCriterion_PASSED, Comment: "another comment"}}
	benchmark := &qf.GradingBenchmark{AssignmentID: assignmentWithReviewers.GetID(), Heading: "Major league baseball", Comment: "wonders of the world", Criteria: criteria}
	qtest.CreateBenchmark(t, db, benchmark)
	newBenchmark := &qf.GradingBenchmark{ID: 2, AssignmentID: assignmentWithReviewers.GetID(), ReviewID: 1, Heading: "Major league baseball", Comment: "wonders of the world", Criteria: []*qf.GradingCriterion{{ID: 2, BenchmarkID: 2, Points: 30, Description: "my description", Grade: qf.GradingCriterion_PASSED, Comment: "another comment"}}}

	tests := []struct {
		name       string
		submission *qf.Submission
		assignment *qf.Assignment
		review     *qf.Review
		wantErr    error
		wantReview *qf.Review
	}{
		{
			name:    "No submission",
			review:  &qf.Review{SubmissionID: 443},
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:       "No reviewers for assignment",
			submission: submission,
			assignment: assignment,
			review:     &qf.Review{SubmissionID: submission.GetID()},
			wantErr:    database.ErrAllReviewsCreated(submission.GetID(), assignment.GetName(), assignment.GetReviewers()),
		},
		{
			name:       "Create review, calculate score and copy benchmark",
			submission: submissionWithReviewers,
			assignment: assignmentWithReviewers,
			review:     &qf.Review{SubmissionID: submissionWithReviewers.GetID(), ReviewerID: 1, Feedback: "my very good feedback", Ready: false, Score: 95, GradingBenchmarks: []*qf.GradingBenchmark{benchmark}},
			wantReview: &qf.Review{ID: 1, Edited: timestamppb.Now(), SubmissionID: submissionWithReviewers.GetID(), ReviewerID: 1, Feedback: "my very good feedback", Ready: false, Score: 30, GradingBenchmarks: []*qf.GradingBenchmark{newBenchmark}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			qtest.CheckError(t, db.CreateReview(test.review), test.wantErr)

			// Skip comparing the reviews if we expect an error or the wanted review is nil
			if test.wantErr != nil || test.wantReview == nil {
				return
			}

			gotReview := qtest.GetReview(t, db, test.review.GetID())
			qtest.Diff(t, "Expected same review, but got", gotReview, test.wantReview, protocmp.Transform(), protocmp.IgnoreFields(test.wantReview, "edited"))
		})
	}
}

func TestUpdateReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, _ := qtest.SetupCourseAssignment(t, db)
	assignment := &qf.Assignment{CourseID: course.GetID(), Order: 2, Reviewers: 1}
	qtest.CreateAssignment(t, db, assignment)
	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submission)
	qtest.CreateReview(t, db, &qf.Review{SubmissionID: submission.GetID(), Score: 95})

	criteria := []*qf.GradingCriterion{{Points: 40, Description: "my description", Grade: qf.GradingCriterion_PASSED, Comment: "another comment"}}
	benchmark := &qf.GradingBenchmark{ID: 1, AssignmentID: assignment.GetID(), Heading: "Major league baseball", Comment: "wonders of the world", Criteria: criteria}
	qtest.CreateBenchmark(t, db, benchmark)

	newCriteria := []*qf.GradingCriterion{{Points: 88, Description: "my description 2", Grade: qf.GradingCriterion_NONE, Comment: "another comment 2"}}
	newBenchmark := &qf.GradingBenchmark{ID: 2, AssignmentID: assignment.GetID(), Heading: "Major league baseball", Comment: "wonders of the world", Criteria: newCriteria}

	tests := []struct {
		name       string
		review     *qf.Review
		wantPoints uint32
		wantErr    error
	}{
		{
			name:    "Empty review ID",
			review:  &qf.Review{},
			wantErr: database.ErrEmptyReviewID,
		},
		{
			name:    "No submission",
			review:  &qf.Review{ID: 1, SubmissionID: 443},
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:       "Add existing benchmark",
			review:     &qf.Review{ID: 1, SubmissionID: submission.GetID(), Score: 95, GradingBenchmarks: []*qf.GradingBenchmark{benchmark}},
			wantPoints: 40,
		},
		{
			name:       "Add new benchmark, expecting different score",
			review:     &qf.Review{ID: 1, SubmissionID: submission.GetID(), Score: 45, GradingBenchmarks: []*qf.GradingBenchmark{benchmark, newBenchmark}},
			wantPoints: 40,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			qtest.CheckError(t, db.UpdateReview(test.review), test.wantErr)
			gotReview := qtest.GetReview(t, db, test.review.GetID())
			if test.wantErr != nil {
				return
			}
			test.review.Score = test.wantPoints

			// Expect the score for a submission to be updated if the review score is different
			if gotReview.GetScore() != test.review.GetScore() {
				submission := qtest.GetSubmission(t, db, &qf.Submission{ID: gotReview.GetSubmissionID()})
				if submission.GetScore() != test.review.GetScore() {
					t.Errorf("Expected score %d, but got %d", gotReview.GetScore(), submission.GetScore())
				}
			}
			qtest.Diff(t, "Expected same review", gotReview, test.review, protocmp.Transform(), protocmp.IgnoreFields(test.review, "edited", "score"))
		})
	}
}

func TestCreateUpdateReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, _, assignment := qtest.SetupCourseAssignment(t, db)

	assignment.Reviewers = 1
	benchmarks := []*qf.GradingBenchmark{
		{
			Heading: "Major league baseball",
			Criteria: []*qf.GradingCriterion{
				{
					Description: "my description",
				},
			},
		},
	}
	assignment.GradingBenchmarks = benchmarks
	// Update assignments to create template benchmarks and criteria in the database.
	if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	review := &qf.Review{
		SubmissionID: 1,
		ReviewerID:   1,
	}

	// Create a new review for submission ID 1.
	// This should copy the assignment benchmarks and criteria to the review.
	if err := db.CreateReview(review); err != nil {
		t.Errorf("failed to create review: %v", err)
	}
	sub, err := db.GetSubmission(&qf.Submission{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(sub.GetReviews()) != 1 {
		t.Fatalf("have %d reviews want %d", len(sub.GetReviews()), 1)
	}
	var gotReview *qf.Review
	for _, r := range sub.GetReviews() {
		gotReview = r
	}

	if diff := cmp.Diff(gotReview, review, cmp.Options{protocmp.Transform(), protocmp.IgnoreFields(&qf.Review{}, "edited")}); diff != "" {
		t.Errorf("Expected same review, but got (-got +want):\n%s", diff)
	}

	if len(gotReview.GetGradingBenchmarks()) != 1 {
		t.Fatalf("have %d benchmarks want %d: %+v", len(gotReview.GetGradingBenchmarks()), 1, review)
	}
	if len(gotReview.GetGradingBenchmarks()[0].GetCriteria()) != 1 {
		t.Fatalf("have %d criteria want %d", len(gotReview.GetGradingBenchmarks()[0].GetCriteria()), 1)
	}

	// Set the grade of the first criterion to PASSED and update the review.
	review.GetGradingBenchmarks()[0].GetCriteria()[0].Grade = qf.GradingCriterion_PASSED
	if err := db.UpdateReview(review); err != nil {
		t.Errorf("failed to update review: %v", err)
	}
	sub, err = db.GetSubmission(&qf.Submission{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(sub.GetReviews()) != 1 {
		t.Fatalf("have %d reviews want %d", len(sub.GetReviews()), 1)
	}
	for _, r := range sub.GetReviews() {
		gotReview = r
	}

	// Verify that the updated review matches the expected review.
	if diff := cmp.Diff(gotReview, review, cmp.Options{protocmp.Transform(), protocmp.IgnoreFields(&qf.Review{}, "edited")}); diff != "" {
		t.Errorf("Expected same review, but got (-got +want):\n%s", diff)
	}
}
