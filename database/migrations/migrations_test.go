package migrations_test

// go test ./database/migrations/... -v

// these tests check all 21 tables from migrations

// rollback test checks the down migration, Expected to fail query after rollback
import (
	"context"
	"database/sql"
	"testing"

	"github.com/quickfeed/quickfeed/database/migrations"
	"github.com/quickfeed/quickfeed/database/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	bunmigrate "github.com/uptrace/bun/migrate"
)

func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.ShimName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	db.RegisterModel((*models.GroupUser)(nil))
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrations(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	migrator := bunmigrate.NewMigrator(db, migrations.Migrations)
	if err := migrator.Init(ctx); err != nil {
		t.Fatal("init:", err)
	}
	if _, err := migrator.Migrate(ctx); err != nil {
		t.Fatal("migrate up:", err)
	}

	t.Run("users", func(t *testing.T) {
		row := &models.User{Login: "alice", Name: "Alice"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.User)
		if err := db.NewSelect().Model(got).Where("login = ?", "alice").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Login != row.Login {
			t.Errorf("got login %q, want %q", got.Login, row.Login)
		}
	})

	t.Run("courses", func(t *testing.T) {
		row := &models.Course{Code: "DAT320", Year: 2026}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Course)
		if err := db.NewSelect().Model(got).Where("code = ?", "DAT320").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Code != row.Code {
			t.Errorf("got code %q, want %q", got.Code, row.Code)
		}
	})

	t.Run("groups", func(t *testing.T) {
		row := &models.Group{Name: "team-alpha", CourseID: 1}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Group)
		if err := db.NewSelect().Model(got).Where("name = ?", "team-alpha").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Name != row.Name {
			t.Errorf("got name %q, want %q", got.Name, row.Name)
		}
	})

	t.Run("group_users", func(t *testing.T) {
		row := &models.GroupUser{GroupID: 1, UserID: 1}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.GroupUser)
		if err := db.NewSelect().Model(got).Where("group_id = ? AND user_id = ?", 1, 1).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.GroupID != row.GroupID || got.UserID != row.UserID {
			t.Errorf("got (%d,%d), want (%d,%d)", got.GroupID, got.UserID, row.GroupID, row.UserID)
		}
	})

	t.Run("repositories", func(t *testing.T) {
		row := &models.Repository{HTMLURL: "https://github.com/org/repo", ScmRepositoryID: 42}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Repository)
		if err := db.NewSelect().Model(got).Where("html_url = ?", row.HTMLURL).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.HTMLURL != row.HTMLURL {
			t.Errorf("got html_url %q, want %q", got.HTMLURL, row.HTMLURL)
		}
	})

	t.Run("enrollments", func(t *testing.T) {
		row := &models.Enrollment{CourseID: 1, UserID: 1}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Enrollment)
		if err := db.NewSelect().Model(got).Where("course_id = ? AND user_id = ?", 1, 1).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.CourseID != row.CourseID {
			t.Errorf("got course_id %d, want %d", got.CourseID, row.CourseID)
		}
	})

	t.Run("assignments", func(t *testing.T) {
		row := &models.Assignment{Name: "lab1", CourseID: 1, Order: 1}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Assignment)
		if err := db.NewSelect().Model(got).Where("name = ?", "lab1").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Name != row.Name {
			t.Errorf("got name %q, want %q", got.Name, row.Name)
		}
	})

	t.Run("test_infos", func(t *testing.T) {
		row := &models.TestInfo{TestName: "TestFoo", AssignmentID: 1, MaxScore: 100}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.TestInfo)
		if err := db.NewSelect().Model(got).Where("test_name = ?", "TestFoo").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.MaxScore != row.MaxScore {
			t.Errorf("got max_score %d, want %d", got.MaxScore, row.MaxScore)
		}
	})

	t.Run("used_slip_days", func(t *testing.T) {
		row := &models.UsedSlipDay{EnrollmentID: 1, AssignmentID: 1, UsedDays: 2}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.UsedSlipDay)
		if err := db.NewSelect().Model(got).Where("enrollment_id = ?", 1).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.UsedDays != row.UsedDays {
			t.Errorf("got used_days %d, want %d", got.UsedDays, row.UsedDays)
		}
	})

	t.Run("submissions", func(t *testing.T) {
		row := &models.Submission{AssignmentID: 1, UserID: 1, CommitHash: "abc123", Score: 80}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Submission)
		if err := db.NewSelect().Model(got).Where("commit_hash = ?", "abc123").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Score != row.Score {
			t.Errorf("got score %d, want %d", got.Score, row.Score)
		}
	})

	t.Run("feedback_receipts", func(t *testing.T) {
		row := &models.FeedbackReceipt{AssignmentID: 1, UserID: 1}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.FeedbackReceipt)
		if err := db.NewSelect().Model(got).Where("assignment_id = ? AND user_id = ?", 1, 1).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.AssignmentID != row.AssignmentID {
			t.Errorf("got assignment_id %d, want %d", got.AssignmentID, row.AssignmentID)
		}
	})

	t.Run("reviews", func(t *testing.T) {
		row := &models.Review{SubmissionID: 1, ReviewerID: 1, Feedback: "good work"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Review)
		if err := db.NewSelect().Model(got).Where("submission_id = ?", 1).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Feedback != row.Feedback {
			t.Errorf("got feedback %q, want %q", got.Feedback, row.Feedback)
		}
	})

	t.Run("grading_benchmarks", func(t *testing.T) {
		row := &models.GradingBenchmark{CourseID: 1, AssignmentID: 1, ReviewID: 1, Heading: "correctness"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.GradingBenchmark)
		if err := db.NewSelect().Model(got).Where("heading = ?", "correctness").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Heading != row.Heading {
			t.Errorf("got heading %q, want %q", got.Heading, row.Heading)
		}
	})

	t.Run("grading_criterions", func(t *testing.T) {
		row := &models.GradingCriterion{BenchmarkID: 1, CourseID: 1, Description: "tests pass"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.GradingCriterion)
		if err := db.NewSelect().Model(got).Where("description = ?", "tests pass").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Description != row.Description {
			t.Errorf("got description %q, want %q", got.Description, row.Description)
		}
	})

	t.Run("tasks", func(t *testing.T) {
		row := &models.Task{AssignmentID: 1, Title: "task1", Name: "task-1"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Task)
		if err := db.NewSelect().Model(got).Where("title = ?", "task1").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Title != row.Title {
			t.Errorf("got title %q, want %q", got.Title, row.Title)
		}
	})

	t.Run("issues", func(t *testing.T) {
		row := &models.Issue{RepositoryID: 1, TaskID: 1, ScmIssueNumber: 42}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Issue)
		if err := db.NewSelect().Model(got).Where("scm_issue_number = ?", 42).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.ScmIssueNumber != row.ScmIssueNumber {
			t.Errorf("got scm_issue_number %d, want %d", got.ScmIssueNumber, row.ScmIssueNumber)
		}
	})

	t.Run("pull_requests", func(t *testing.T) {
		row := &models.PullRequest{TaskID: 1, IssueID: 1, UserID: 1, SourceBranch: "feature-x"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.PullRequest)
		if err := db.NewSelect().Model(got).Where("source_branch = ?", "feature-x").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.SourceBranch != row.SourceBranch {
			t.Errorf("got source_branch %q, want %q", got.SourceBranch, row.SourceBranch)
		}
	})

	t.Run("build_infos", func(t *testing.T) {
		row := &models.BuildInfo{SubmissionID: 1, BuildLog: "ok"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.BuildInfo)
		if err := db.NewSelect().Model(got).Where("submission_id = ?", 1).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.BuildLog != row.BuildLog {
			t.Errorf("got build_log %q, want %q", got.BuildLog, row.BuildLog)
		}
	})

	t.Run("scores", func(t *testing.T) {
		row := &models.Score{SubmissionID: 1, TestName: "TestBar", MaxScore: 50, Score: 45}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Score)
		if err := db.NewSelect().Model(got).Where("test_name = ?", "TestBar").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.Score != row.Score {
			t.Errorf("got score %d, want %d", got.Score, row.Score)
		}
	})

	t.Run("grades", func(t *testing.T) {
		row := &models.Grade{SubmissionID: 1, UserID: 1}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.Grade)
		if err := db.NewSelect().Model(got).Where("submission_id = ? AND user_id = ?", 1, 1).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.SubmissionID != row.SubmissionID {
			t.Errorf("got submission_id %d, want %d", got.SubmissionID, row.SubmissionID)
		}
	})

	t.Run("assignment_feedbacks", func(t *testing.T) {
		row := &models.AssignmentFeedback{AssignmentID: 1, CourseID: 1, LikedContent: "exercises"}
		if _, err := db.NewInsert().Model(row).Exec(ctx); err != nil {
			t.Fatal(err)
		}
		got := new(models.AssignmentFeedback)
		if err := db.NewSelect().Model(got).Where("liked_content = ?", "exercises").Scan(ctx); err != nil {
			t.Fatal(err)
		}
		if got.LikedContent != row.LikedContent {
			t.Errorf("got liked_content %q, want %q", got.LikedContent, row.LikedContent)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		if _, err := migrator.Rollback(ctx); err != nil {
			t.Fatal("rollback:", err)
		}
		if _, err := db.ExecContext(ctx, "SELECT 1 FROM users"); err == nil {
			t.Error("expected error querying users after rollback, got nil")
		}
	})
}
