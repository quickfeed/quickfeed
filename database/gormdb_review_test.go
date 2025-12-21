package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

func TestCreateReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := qtest.SetupCourseAssignment(t, db)
	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submission)

	// Create assignment with reviewers, but WITHOUT benchmarks first
	// (benchmarks must be added before submission to auto-create reviews with benchmarks)
	assignmentWithReviewers := &qf.Assignment{CourseID: course.GetID(), Order: 2, Reviewers: 1}
	qtest.CreateAssignment(t, db, assignmentWithReviewers)

	// Add benchmark/criteria to the assignment
	criteria := []*qf.GradingCriterion{{Points: 30, Description: "my description", Grade: qf.GradingCriterion_PASSED, Comment: "another comment"}}
	benchmark := &qf.GradingBenchmark{AssignmentID: assignmentWithReviewers.GetID(), Heading: "Major league baseball", Comment: "wonders of the world", Criteria: criteria}
	qtest.CreateBenchmark(t, db, benchmark)

	// Now create submission - this will auto-create the review
	submissionWithReviewers := &qf.Submission{AssignmentID: assignmentWithReviewers.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submissionWithReviewers)

	tests := []struct {
		name       string
		submission *qf.Submission
		assignment *qf.Assignment
		review     *qf.Review
		wantErr    error
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
			name:       "All reviews already created (auto-created on submission)",
			submission: submissionWithReviewers,
			assignment: assignmentWithReviewers,
			review:     &qf.Review{SubmissionID: submissionWithReviewers.GetID(), ReviewerID: 1, Feedback: "my very good feedback", Ready: false, Score: 95, GradingBenchmarks: []*qf.GradingBenchmark{benchmark}},
			wantErr:    database.ErrAllReviewsCreated(submissionWithReviewers.GetID(), assignmentWithReviewers.GetName(), assignmentWithReviewers.GetReviewers()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			qtest.CheckError(t, db.CreateReview(test.review), test.wantErr)
		})
	}
}

func TestUpdateReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, _ := qtest.SetupCourseAssignment(t, db)
	assignment := &qf.Assignment{CourseID: course.GetID(), Order: 2, Reviewers: 1}
	qtest.CreateAssignment(t, db, assignment)

	// Create benchmark before creating submission so it gets copied to auto-created review
	criteria := []*qf.GradingCriterion{{Points: 40, Description: "my description", Grade: qf.GradingCriterion_NONE, Comment: ""}}
	benchmark := &qf.GradingBenchmark{AssignmentID: assignment.GetID(), Heading: "Major league baseball", Comment: "", Criteria: criteria}
	qtest.CreateBenchmark(t, db, benchmark)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submission)

	// Get the auto-created review
	sub, err := db.GetSubmission(&qf.Submission{ID: submission.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(sub.GetReviews()) != 1 {
		t.Fatalf("expected 1 auto-created review, got %d", len(sub.GetReviews()))
	}
	autoCreatedReview := sub.GetReviews()[0]

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
			name:       "Update auto-created review - set criterion grade to PASSED",
			review:     &qf.Review{ID: autoCreatedReview.GetID(), SubmissionID: submission.GetID(), GradingBenchmarks: autoCreatedReview.GetGradingBenchmarks()},
			wantPoints: 40, // Score is 40 because the criterion has 40 points and is graded as PASSED
		},
	}

	// Set the grade of the criterion to PASSED for the last test
	if len(autoCreatedReview.GetGradingBenchmarks()) > 0 && len(autoCreatedReview.GetGradingBenchmarks()[0].GetCriteria()) > 0 {
		autoCreatedReview.GetGradingBenchmarks()[0].GetCriteria()[0].Grade = qf.GradingCriterion_PASSED
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			qtest.CheckError(t, db.UpdateReview(test.review), test.wantErr)
			if test.wantErr != nil {
				return
			}
			gotReview := qtest.GetReview(t, db, test.review.GetID())

			// Expect the score for a submission to be updated if the review score is different
			if gotReview.GetScore() != test.wantPoints {
				t.Errorf("Expected score %d, but got %d", test.wantPoints, gotReview.GetScore())
			}
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

	// The review should be auto-created when the submission is created
	sub, err := db.GetSubmission(&qf.Submission{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(sub.GetReviews()) != 1 {
		t.Fatalf("have %d reviews want %d (should be auto-created)", len(sub.GetReviews()), 1)
	}
	var gotReview *qf.Review
	for _, r := range sub.GetReviews() {
		gotReview = r
	}

	// Verify that the auto-created review has ReviewerID = 0
	if gotReview.GetReviewerID() != 0 {
		t.Errorf("Expected auto-created review to have ReviewerID = 0, got %d", gotReview.GetReviewerID())
	}

	if len(gotReview.GetGradingBenchmarks()) != 1 {
		t.Fatalf("have %d benchmarks want %d: %+v", len(gotReview.GetGradingBenchmarks()), 1, gotReview)
	}
	if len(gotReview.GetGradingBenchmarks()[0].GetCriteria()) != 1 {
		t.Fatalf("have %d criteria want %d", len(gotReview.GetGradingBenchmarks()[0].GetCriteria()), 1)
	}

	// Set the grade of the first criterion to PASSED and update the review.
	gotReview.GetGradingBenchmarks()[0].GetCriteria()[0].Grade = qf.GradingCriterion_PASSED
	if err := db.UpdateReview(gotReview); err != nil {
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

	// Verify that the updated review has the expected score (100 since all criteria are passed)
	if gotReview.GetScore() != 100 {
		t.Errorf("Expected score 100 after grading all criteria as PASSED, got %d", gotReview.GetScore())
	}
}
