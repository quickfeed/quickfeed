package database

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

// TestUpdateReviewKeepsConcurrentTestRun checks that a test run recorded while
// UpdateReview runs, after it has first read the submission, is kept: the
// review's save must not restore the submission's earlier test results.
func TestUpdateReviewKeepsConcurrentTestRun(t *testing.T) {
	db, err := NewGormDB(filepath.Join(t.TempDir(), "test.db"), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	mustDo := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	user := &qf.User{Name: "Ada Lovelace", Email: "ada@example.com", StudentID: "1"}
	mustDo(db.CreateUser(user))
	course := &qf.Course{Code: "DAT100"}
	mustDo(db.CreateCourse(user.GetID(), course))
	assignment := &qf.Assignment{CourseID: course.GetID(), Name: "lab1", Order: 1, Reviewers: 1, ReviewWeight: 40}
	mustDo(db.CreateAssignment(assignment))
	mustDo(db.CreateBenchmark(&qf.GradingBenchmark{
		AssignmentID: assignment.GetID(),
		Heading:      "Design",
		Criteria:     []*qf.GradingCriterion{{Description: "Structure", Points: 10}},
	}))
	testRun := func(commit string, testScore int32) *qf.Submission {
		return &qf.Submission{
			AssignmentID: assignment.GetID(),
			UserID:       user.GetID(),
			CommitHash:   commit,
			TestScore:    uint32(testScore),
			Score:        assignment.WeightedScore(uint32(testScore), 0),
			BuildInfo:    &score.BuildInfo{BuildLog: commit},
			Scores:       []*score.Score{{TestName: "TestA", Score: testScore, MaxScore: 100, Weight: 1}},
		}
	}
	submission := testRun("old", 50)
	mustDo(db.CreateSubmission(submission))
	review := &qf.Review{SubmissionID: submission.GetID(), ReviewerID: user.GetID()}
	mustDo(db.CreateReview(review))
	review, err = db.GetReview(&qf.Review{ID: review.GetID()})
	mustDo(err)

	// Record a newer test run right after UpdateReview first reads the submission.
	recorded := false
	mustDo(db.conn.Callback().Query().After("gorm:after_query").Register("test:record_test_run", func(tx *gorm.DB) {
		if recorded || tx.Statement.Table != "submissions" {
			return
		}
		recorded = true
		newRun := testRun("new", 80)
		newRun.ID = submission.GetID()
		if err := db.CreateSubmission(newRun); err != nil {
			t.Error(err)
		}
	}))
	review.GetGradingBenchmarks()[0].GetCriteria()[0].Grade = qf.GradingCriterion_PASSED
	mustDo(db.UpdateReview(review))
	if !recorded {
		t.Fatal("no test run was recorded during UpdateReview")
	}

	got, err := db.GetSubmission(&qf.Submission{ID: submission.GetID()})
	mustDo(err)
	if got.GetCommitHash() != "new" || got.GetTestScore() != 80 || got.GetBuildInfo().GetBuildLog() != "new" {
		t.Errorf("(CommitHash, TestScore, BuildLog) = (%q, %d, %q), want the new test run's (\"new\", 80, \"new\")",
			got.GetCommitHash(), got.GetTestScore(), got.GetBuildInfo().GetBuildLog())
	}
	if len(got.GetScores()) != 1 || got.GetScores()[0].GetScore() != 80 {
		t.Errorf("Scores = %v, want only the new test run's score of 80", got.GetScores())
	}
	if want := assignment.WeightedScore(80, 100); got.GetReviewScore() != 100 || got.GetScore() != want {
		t.Errorf("(ReviewScore, Score) = (%d, %d), want (100, %d)", got.GetReviewScore(), got.GetScore(), want)
	}
}
