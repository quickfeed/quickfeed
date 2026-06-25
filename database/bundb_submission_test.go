package database_test

import (
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBunDBGetSubmissionForUser(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	query := &qf.Submission{AssignmentID: 10, UserID: 10}
	if _, err := db.GetSubmission(query); !isNotFound(err) {
		t.Errorf("have error '%v' wanted sql.ErrNoRows", err)
	}
}

func TestBunGetSubmissions(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	submission := &qf.Submission{AssignmentID: 1, UserID: 1}
	submission1 := &qf.Submission{AssignmentID: 1, UserID: 2}

	var wantSubmissions []*qf.Submission
	tests := []struct {
		name          string
		query         *qf.Submission
		newSubmission *qf.Submission
		wantError     error
	}{
		{name: "No Assignment ID", query: &qf.Submission{}, wantError: sql.ErrNoRows},
		{name: "Invalid assignment ID", query: &qf.Submission{AssignmentID: 4, UserID: 2}, wantError: sql.ErrNoRows},
		{name: "First submission", query: &qf.Submission{AssignmentID: 1}, newSubmission: submission},
		{name: "Second submission", query: &qf.Submission{AssignmentID: 1}, newSubmission: submission1},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if i == 1 {
				_, _, _ = qtest.SetupCourseAssignment(t, db)
				qtest.CreateFakeUser(t, db)
			}
			if test.newSubmission != nil {
				qtest.CreateSubmission(t, db, test.newSubmission)
				wantSubmissions = append(wantSubmissions, test.newSubmission)
			}

			submissions, err := db.GetSubmissions(test.query)
			qtest.CheckError(t, err, test.wantError)

			if test.wantError != nil {
				return
			}

			qtest.Diff(t, "GetSubmissions() = mismatch", submissions, wantSubmissions, protocmp.Transform())
		})
	}
}

func TestBunDBCreateSubmissionWithAutoApprove(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	user, _, assignment := qtest.SetupCourseAssignment(t, db)

	assignment.AutoApprove = true
	assignment.ScoreLimit = 1

	if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		in   *qf.Submission
		want *qf.Submission
	}{
		{name: "Approved", in: &qf.Submission{AssignmentID: assignment.ID, UserID: user.ID, Score: 1}, want: &qf.Submission{ID: 1, AssignmentID: assignment.ID, UserID: user.ID, Score: 1, Grades: []*qf.Grade{{UserID: user.ID, SubmissionID: 1, Status: qf.Submission_APPROVED}}}},
		{name: "NotApproved", in: &qf.Submission{AssignmentID: assignment.ID, UserID: user.ID, Score: 0}, want: &qf.Submission{ID: 2, AssignmentID: assignment.ID, UserID: user.ID, Score: 0, Grades: []*qf.Grade{{UserID: user.ID, SubmissionID: 2, Status: qf.Submission_NONE}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.CreateSubmission(tt.in); err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tt.in, tt.want, protocmp.Transform()); diff != "" {
				t.Errorf("CreateSubmission(): (-got +want):\n%s", diff)
			}
		})
	}
}

func TestBunDBUpdateSubmissionScore(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	user, _, assignment := qtest.SetupCourseAssignment(t, db)
	submission := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
	}
	if err := db.CreateSubmission(submission); err != nil {
		t.Fatal(err)
	}
	submission.Score = 100
	if err := db.UpdateSubmission(submission); err != nil {
		t.Fatal(err)
	}
	gotSubmission, err := db.GetSubmission(submission)
	if err != nil {
		t.Fatal(err)
	}
	qtest.Diff(t, "Expected score to be 100", gotSubmission, submission, protocmp.Transform())
}

func TestBunDBUpdateSubmissionZeroScore(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        80,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Errorf("have %d submissions want %d", len(submissions), 1)
	}
	want := &qf.Submission{
		ID:           submissions[0].GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        80,
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: submissions[0].GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        0,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	want = &qf.Submission{
		ID:           submissions[0].GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        0,
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: submissions[0].GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}
}

func TestBunDBUpdateSubmission(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	want := &qf.Submission{
		ID:           submissions[0].GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: submissions[0].GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	if submissions[0].GetStatusByUser(want.GetUserID()) != qf.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}

	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}

	if submissions[0].GetStatusByUser(want.GetUserID()) != qf.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}
	submissions[0].SetGradeByUser(user.GetID(), qf.Submission_APPROVED)
	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if submissions[0].GetStatusByUser(want.GetUserID()) != qf.Submission_APPROVED {
		t.Errorf("expected submission to be 'approved' but got 'not-approved'")
	}
}

func TestBunDBGetNonExistingSubmissions(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	if _, err := db.GetLastSubmissions(10, &qf.Submission{UserID: 10}); !isNotFound(err) {
		t.Errorf("have error '%v' wanted sql.ErrNoRows", err)
	}
}

func TestBunDBInsertSubmissions(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: 1,
		UserID:       1,
	}); !isNotFound(err) {
		t.Fatal(err)
	}

	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       3,
	}); !isNotFound(err) {
		t.Fatal(err)
	}

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}
	gotSubmission := submissions[0]
	wantSubmission := &qf.Submission{
		ID:           gotSubmission.GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: gotSubmission.GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}

	if diff := cmp.Diff(wantSubmission, gotSubmission, protocmp.Transform()); diff != "" {
		t.Errorf("GetLastSubmissions() mismatch (-wantSubmission, +gotSubmission):\n%s", diff)
	}
}

// TestBunDBRemoveTest mirrors gormdb_submission_testremove_test.go
func TestBunDBRemoveTest(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	buildInfo := &score.BuildInfo{
		BuildDate: qtest.Timestamp(t, "2022-11-10T13:00:00"),
		BuildLog:  "Testing",
		ExecTime:  33333,
	}
	scores := []*score.Score{
		{TestName: "Test1", Score: 10, MaxScore: 15, Weight: 1},
		{TestName: "Test2", Score: 0, MaxScore: 5, Weight: 1},
		{TestName: "Test3", Score: 3, MaxScore: 5, Weight: 1},
	}
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		BuildInfo:    buildInfo,
		Scores:       scores,
	}); err != nil {
		t.Fatal(err)
	}
	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	buildInfo.SubmissionID = submissions[0].GetID()
	buildInfo.ID = 1
	if diff := cmp.Diff(buildInfo, submissions[0].GetBuildInfo(), protocmp.Transform()); diff != "" {
		t.Errorf("Expected same build info, but got (-got +want):\n%s", diff)
	}
	if diff := cmp.Diff(
		submissions[0].GetScores(),
		scores,
		protocmp.Transform(),
		protocmp.IgnoreFields(&score.Score{}, "ID", "SubmissionID")); diff != "" {
		t.Errorf("Incorrect scores after first save (-want, +got):\n%s", diff)
	}

	oldSubmissionID := submissions[0].GetID()
	updatedBuildInfo := &score.BuildInfo{
		BuildDate: qtest.Timestamp(t, "2022-11-10T15:00:00"),
		BuildLog:  "Updated",
		ExecTime:  12345,
	}
	scores = []*score.Score{
		{TestName: "Test1", Score: 10, MaxScore: 15, Weight: 1},
		{TestName: "Test3", Score: 3, MaxScore: 5, Weight: 1},
	}

	submissions[0].BuildInfo = updatedBuildInfo
	submissions[0].Scores = scores
	if err := db.CreateSubmission(submissions[0]); err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	updatedBuildInfo.ID = submissions[0].GetBuildInfo().GetID()
	updatedBuildInfo.SubmissionID = oldSubmissionID
	if diff := cmp.Diff(submissions[0].GetBuildInfo(), updatedBuildInfo, protocmp.Transform()); diff != "" {
		t.Errorf("Expected updated build info, but got (-sub +want):\n%s", diff)
	}
	if diff := cmp.Diff(submissions[0].GetScores(), scores, protocmp.Transform()); diff != "" {
		t.Errorf("Incorrect scores after update (-want, +got):\n%s", diff)
	}

	submissions[0].ID = 123
	if err := db.CreateSubmission(submissions[0]); err == nil {
		t.Fatal("expected error: record not found")
	}
}

// TestBunCreateReview mirrors gormdb_review_test.go TestCreateReview
func TestBunCreateReview(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
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
			wantErr: sql.ErrNoRows,
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
			review:     &qf.Review{SubmissionID: submissionWithReviewers.GetID(), ReviewerID: 1, Feedback: "my very good feedback", Score: 95, GradingBenchmarks: []*qf.GradingBenchmark{benchmark}},
			wantReview: &qf.Review{ID: 1, Edited: timestamppb.Now(), SubmissionID: submissionWithReviewers.GetID(), ReviewerID: 1, Feedback: "my very good feedback", Score: 30, GradingBenchmarks: []*qf.GradingBenchmark{newBenchmark}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			qtest.CheckError(t, db.CreateReview(test.review), test.wantErr)

			if test.wantErr != nil || test.wantReview == nil {
				return
			}

			gotReview := qtest.GetReview(t, db, test.review.GetID())
			qtest.Diff(t, "Expected same review, but got", gotReview, test.wantReview, protocmp.Transform(), protocmp.IgnoreFields(test.wantReview, "edited"))
		})
	}
}

// TestBunUpdateReview mirrors gormdb_review_test.go TestUpdateReview
func TestBunUpdateReview(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
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
			wantErr: sql.ErrNoRows,
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

// TestBunCreateUpdateReview mirrors gormdb_review_test.go TestCreateUpdateReview
func TestBunCreateUpdateReview(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
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

	if diff := cmp.Diff(gotReview, review, cmp.Options{protocmp.Transform(), protocmp.IgnoreFields(&qf.Review{}, "edited")}); diff != "" {
		t.Errorf("Expected same review, but got (-got +want):\n%s", diff)
	}
}
