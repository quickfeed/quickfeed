package database

import (
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDropRetired(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
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
