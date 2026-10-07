package database

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/quickfeed/quickfeed/qf"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDropRetired(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		"CREATE TABLE pull_requests (id INTEGER PRIMARY KEY)",
		"CREATE TABLE tasks (id INTEGER PRIMARY KEY)",
		"CREATE TABLE issues (id INTEGER PRIMARY KEY)",
		"CREATE TABLE scores (id INTEGER PRIMARY KEY, submission_id INTEGER, test_name TEXT, task_name TEXT, score INTEGER, max_score INTEGER, weight INTEGER, test_details TEXT)",
	} {
		if err := conn.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := dropRetired(conn); err != nil {
		t.Fatal(err)
	}

	m := conn.Migrator()
	for _, table := range retiredTables {
		if m.HasTable(table) {
			t.Errorf("HasTable(%q) = true, want false", table)
		}
	}
	if m.HasColumn("scores", "task_name") {
		t.Error(`HasColumn("scores", "task_name") = true, want false`)
	}
	if !m.HasColumn("scores", "test_name") {
		t.Error(`HasColumn("scores", "test_name") = false, want true`)
	}

	// Dropping again must be a no-op.
	if err := dropRetired(conn); err != nil {
		t.Fatal(err)
	}
}

func TestBackfillScores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// The assignments and submissions tables as they were before test and
	// review scores were kept apart.
	for _, stmt := range []string{
		"CREATE TABLE assignments (id INTEGER PRIMARY KEY, course_id INTEGER, name TEXT, reviewers INTEGER)",
		"CREATE TABLE submissions (id INTEGER PRIMARY KEY, assignment_id INTEGER, user_id INTEGER, group_id INTEGER, score INTEGER)",
		"INSERT INTO assignments (id, course_id, name, reviewers) VALUES (1, 1, 'tested', 0), (2, 1, 'reviewed', 2)",
		"INSERT INTO submissions (id, assignment_id, user_id, group_id, score) VALUES (1, 1, 1, 0, 85), (2, 2, 1, 0, 70)",
	} {
		if err := conn.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	if db, _ := conn.DB(); db != nil {
		_ = db.Close()
	}

	db, err := NewGormDB(path, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	wantWeights := map[uint64]uint32{1: 0, 2: 100}
	for id, want := range wantWeights {
		var assignment qf.Assignment
		if err := db.conn.First(&assignment, id).Error; err != nil {
			t.Fatal(err)
		}
		if got := assignment.GetReviewWeight(); got != want {
			t.Errorf("assignment %d: ReviewWeight = %d, want %d", id, got, want)
		}
	}
	wantSubmissions := map[uint64]*qf.Submission{
		1: {Score: 85, TestScore: 85, ReviewScore: 0},
		2: {Score: 70, TestScore: 0, ReviewScore: 70},
	}
	for id, want := range wantSubmissions {
		var got qf.Submission
		if err := db.conn.First(&got, id).Error; err != nil {
			t.Fatal(err)
		}
		if got.GetScore() != want.GetScore() || got.GetTestScore() != want.GetTestScore() || got.GetReviewScore() != want.GetReviewScore() {
			t.Errorf("submission %d: (Score, TestScore, ReviewScore) = (%d, %d, %d), want (%d, %d, %d)", id,
				got.GetScore(), got.GetTestScore(), got.GetReviewScore(), want.GetScore(), want.GetTestScore(), want.GetReviewScore())
		}
	}

	// Reopening a migrated database must not backfill again.
	if err := db.conn.Exec("UPDATE submissions SET score = 50 WHERE id = 1").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = NewGormDB(path, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	var got qf.Submission
	if err := db.conn.First(&got, 1).Error; err != nil {
		t.Fatal(err)
	}
	if got.GetTestScore() != 85 {
		t.Errorf("after reopening: TestScore = %d, want 85", got.GetTestScore())
	}
}
