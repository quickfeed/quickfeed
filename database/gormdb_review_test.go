package database_test

import (
	"testing"

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
